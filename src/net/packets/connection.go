package packets

import "github.com/gorilla/websocket"

type Connection struct {
	Conn *websocket.Conn
}

func (c *Connection) SendPacket(packet Packet) {}
