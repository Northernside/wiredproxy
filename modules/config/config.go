package config

// JSON based configuration file

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"wired.rip/wiredutils/protocol"
	"wired.rip/wiredutils/utils"
)

// Configuration struct
type RoutesConfig struct {
	Routes []protocol.Route `json:"routes"`
}

type Node struct {
	Id             string `json:"id"`
	Passphrase     string `json:"passphrase"`
	LastConnection int64  `json:"last_connection"`
}

type SystemConfig struct {
	WiredHost           string           `json:"wired_host"`
	SystemKey           string           `json:"system_key"`
	CurrentAmd64Hash    string           `json:"current_amd64_hash"`
	CurrentArm64Hash    string           `json:"current_arm64_hash"`
	DiscordClientId     string           `json:"discord_client_id"`
	DiscordClientSecret string           `json:"discord_client_secret"`
	DiscordRedirectUri  string           `json:"discord_redirect_uri"`
	JwtSigningKey       string           `json:"jwt_signing_key"`
	AdminDiscordId      string           `json:"admin_discord_id"`
	Passphrase          string           `json:"passphrase"`
	Mode                string           `json:"mode"`
	Nodes               []Node           `json:"nodes"`
	Routes              []protocol.Route `json:"routes"`
}

var config SystemConfig

func AddRoute(route protocol.Route) int {
	config.Routes = append(config.Routes, route)
	saveConfigFile("config.json")

	return http.StatusOK
}

func SetRoutes(routes []protocol.Route) {
	config.Routes = routes
	saveConfigFile("config.json")
}

func DeleteRoute(routeId string) int {
	for i, r := range config.Routes {
		if r.RouteId == routeId {
			config.Routes = append(config.Routes[:i], config.Routes[i+1:]...)
			saveConfigFile("config.json")
			return http.StatusOK
		}
	}

	return http.StatusNotFound
}

func GetRoutes() []protocol.Route {
	return config.Routes
}

func GetNodes() []Node {
	return config.Nodes
}

func GetNode(nodeId string) (Node, bool) {
	for _, n := range config.Nodes {
		if n.Id == nodeId {
			return n, true
		}
	}

	return Node{}, false
}

func AddNode(node Node) int {
	config.Nodes = append(config.Nodes, node)
	saveConfigFile("config.json")

	return http.StatusOK
}

func DeleteNode(nodeId string) int {
	for i, n := range config.Nodes {
		if n.Id == nodeId {
			config.Nodes = append(config.Nodes[:i], config.Nodes[i+1:]...)
			saveConfigFile("config.json")
			return http.StatusOK
		}
	}

	return http.StatusNotFound
}

func SetSystemKey(key string) {
	config.SystemKey = key
	saveConfigFile("config.json")
}

func GetSystemKey() string {
	return config.SystemKey
}

func SetPassphrase(passphrase string) {
	config.Passphrase = passphrase
	saveConfigFile("config.json")
}

func GetPassphrase() string {
	return config.Passphrase
}

func GetWiredHost() string {
	return config.WiredHost
}

func GetRouteByProxyDomain(proxyDomain string) (protocol.Route, bool) {
	for _, r := range config.Routes {
		if r.ProxyDomain == proxyDomain {
			return r, true
		}
	}

	return protocol.Route{}, false
}

func SetCurrentNodeHash(hash string, arch string) {
	switch arch {
	case "amd64":
		config.CurrentAmd64Hash = hash
	case "arm64":
		config.CurrentArm64Hash = hash
	}

	saveConfigFile("config.json")
}

func GetCurrentNodeHash(arch string) string {
	switch arch {
	case "amd64":
		return config.CurrentAmd64Hash
	case "arm64":
		return config.CurrentArm64Hash
	}

	return ""
}

func GetDiscordClientId() string {
	return config.DiscordClientId
}

func GetDiscordClientSecret() string {
	return config.DiscordClientSecret
}

func GetDiscordRedirectUri() string {
	return config.DiscordRedirectUri
}

func GetJwtSigningKey() string {
	return config.JwtSigningKey
}

func GetAdminDiscordId() string {
	return config.AdminDiscordId
}

func GetMode() string {
	return config.Mode
}

func Init() {
	// create if not exists
	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
		config = SystemConfig{
			WiredHost:        "wired.rip",
			SystemKey:        fmt.Sprintf("node-%s", utils.GenerateString(8)),
			CurrentAmd64Hash: "",
			CurrentArm64Hash: "",
			Routes:           []protocol.Route{},
		}

		saveConfigFile("config.json")
	}

	config = readConfigFile("config.json")
}

func readConfigFile(configFile string) SystemConfig {
	var config SystemConfig

	file, err := os.Open(configFile)
	if err != nil {
		log.Println("Error opening configuration file:", err)
		return config
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("Error reading configuration file:", err)
		return config
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Println("Error unmarshalling configuration file:", err)
		return config
	}

	return config
}

func saveConfigFile(configFile string) {
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		log.Println("Error marshalling configuration file:", err)
		return
	}

	err = ioutil.WriteFile(configFile, data, 0644)
	if err != nil {
		log.Println("Error writing configuration file:", err)
	}
}
