package packet

import "wired.rip/wiredutils/protocol"

const (
	Id_SharedSecret protocol.VarInt = 0
	Id_Hello        protocol.VarInt = 1
	Id_Ready        protocol.VarInt = 2
	Id_Ping         protocol.VarInt = 3
	Id_Pong         protocol.VarInt = 4
	Id_Routes       protocol.VarInt = 5
)

type Hello struct {
	Key     string
	Version string
	Hash    []byte
}

type Routes struct {
	Routes []protocol.Route
}
