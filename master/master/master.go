package master

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
	"wiredmaster/routes"

	"wired.rip/wiredutils/config"
	"wired.rip/wiredutils/packet"
	"wired.rip/wiredutils/protocol"
	"wired.rip/wiredutils/terminal"
)

var clients = make(map[string]protocol.Conn)

func Run() {
	config.Init()
	log.SetFlags(0)
	prefix := fmt.Sprintf("%s.%s » ", config.GetSystemKey(), config.GetWiredHost())
	log.SetPrefix(terminal.PrefixColor + prefix + terminal.Reset)

	go startHttpServer()
	go configChecker()
	loadWiredKeyPair()
	startServer()
}

var currentRoutes []protocol.Route
var connected = false

func configChecker() {
	log.Println("Starting config checker")
	for {
		if !connected {
			log.Println("Not connected to any clients, waiting...")
			time.Sleep(1 * time.Second)
			continue
		}

		routes := config.GetRoutes()
		if !routesEqual(currentRoutes, routes) {
			currentRoutes = routes

			log.Println("Sending updated routes to clients")
			for _, client := range clients {
				err := client.SendPacket(packet.Id_Routes, packet.Routes{
					Routes: routes,
				})

				if err != nil {
					log.Println("Error sending routes packet to client:", err)
				}
			}
		}

		time.Sleep(5 * time.Second)
	}
}

func routesEqual(a, b []protocol.Route) bool {
	if len(a) != len(b) {
		return false
	}

	for i, route := range a {
		if route != b[i] {
			return false
		}
	}

	return true
}

func startHttpServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "Not found"}`))
			return
		}

		routes.IndexRoute(w, r)
	})

	customHandler("/api/routes", routes.GetRoutes, http.MethodGet)
	customHandler("/api/routes/add", routes.AddRoute, http.MethodGet)
	customHandler("/api/routes/remove", routes.RemoveRoute, http.MethodDelete)

	http.HandleFunc("/api/connect/publickey", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		w.Write(pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(&wiredKey.PublicKey),
		}))
	})

	log.Println("HTTP server listening on 127.0.0.1:37421")
	http.ListenAndServe("127.0.0.1:37421", nil)
}

func customHandler(path string, handler http.HandlerFunc, method string) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"message": "Method not allowed"}`))
			return
		}

		handler(w, r)
	})
}

func startServer() {
	server, err := net.Listen("tcp", ":37420")
	if err != nil {
		panic(err)
	}

	connected = true
	log.Println("Communication server listening on *:37420")
	for {
		conn, err := server.Accept()
		if err != nil {
			continue
		}

		go handleConnection(protocol.NewConn(conn, wiredKey, nil))
	}
}

func handleConnection(conn *protocol.Conn) {
	defer conn.Close()

	var pp protocol.Packet
	err := pp.Read(conn)
	if err != nil {
		return
	}

	var sharedSecret []byte
	err = protocol.DecodePacket(pp.Data, &sharedSecret)
	if err != nil {
		log.Println("Error decoding shared secret:", err)
		return
	}

	err = conn.EnableEncryption(sharedSecret)
	if err != nil {
		log.Println("Error enabling encryption:", err)
		return
	}

	err = conn.SendPacket(packet.Id_Ready, nil)
	if err != nil {
		log.Println("Error sending ready packet:", err)
		return
	}

	// add client to clients map
	clients[string(conn.Address)] = *conn

	for {
		var pp protocol.Packet
		err := pp.Read(conn)
		if err != nil {
			// log.Println("Error reading packet:", err)
			continue
		}

		// log.Println("Received packet:", pp.ID)
		// log.Println("Data:", string(pp.Data))

		switch pp.ID {
		case packet.Id_Hello:
			log.Printf("Received hello packet at %s\n", time.Now().Format("15:04:05"))
			var hello packet.Hello
			err := protocol.DecodePacket(pp.Data, &hello)
			if err != nil {
				log.Println("Error decoding hello packet:", err)
				continue
			}

			log.Printf("Client %s.%s connected with version %s\n", hello.Key, config.GetWiredHost(), hello.Version)

			routes := config.GetRoutes()

			// send routes packet
			err = conn.SendPacket(packet.Id_Routes, packet.Routes{
				Routes: routes,
			})

			if err != nil {
				log.Println("Error sending routes packet:", err)
				continue
			}
		case packet.Id_Ping:
			err = conn.SendPacket(packet.Id_Pong, nil)
			if err != nil {
				log.Println("Error sending pong packet:", err)
				continue
			}

			// log.Println("Sent pong")
		}
	}
}
