package packets

import (
	"github.com/gorilla/websocket"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

type Connection struct {
	Player game.Player
	Conn   *websocket.Conn
}

func (c *Connection) SendPacket(packet Packet) {
	data, err := Serialize(packet)
	if err != nil {
		return
	}
	err = c.Conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return
	}
}
