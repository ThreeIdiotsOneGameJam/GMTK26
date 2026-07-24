package packets

import (
	"github.com/google/uuid"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

type C2SHelloPacket struct {
	Player game.Player `json:"player"`
}

func init() {
	mustRegisterPacket(
		C2SPingPacketType,
		func() Packet {
			return &C2SHelloPacket{}
		},
		func(packet Packet) bool {
			value, ok := packet.(*C2SHelloPacket)
			return ok && value != nil
		},
	)
}

func (*C2SHelloPacket) PacketType() PacketType {
	return C2SPingPacketType
}

func (p *C2SHelloPacket) Handle(c *Connection) {
	uid, err := uuid.Parse(string(p.Player.ClientID))
	if err != nil {
		// idk
	}

	if uid == uuid.Nil {
		// idk
	}

	// TODO: check other fields

	c.SendPacket(&S2CServerInfoPacket{
		Hello: "hi",
	})
}

func (*C2SHelloPacket) isC2S() {}
