package packet

import "wired.rip/wiredutils/protocol"

const (
	Id_SharedSecret     protocol.VarInt = 0
	Id_Hello            protocol.VarInt = 1
	Id_Ready            protocol.VarInt = 2
	Id_Ping             protocol.VarInt = 3
	Id_Pong             protocol.VarInt = 4
	Id_Routes           protocol.VarInt = 5
	Id_BinaryData       protocol.VarInt = 6
	Id_BinaryEnd        protocol.VarInt = 7
	Id_PlayerAdd        protocol.VarInt = 8
	Id_PlayerRemove     protocol.VarInt = 9
	Id_DisconnectPlayer protocol.VarInt = 10
)

type Hello struct {
	Key        string
	Version    string
	Passphrase string
	Arch       string
	Hash       []byte
}

type BinaryData struct {
	Label string
	Data  []byte
}

type Routes struct {
	Routes []protocol.Route
}

type Disconnect struct {
	PlayerUUID string
	ProxyHost  string
}
