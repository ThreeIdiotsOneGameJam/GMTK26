package packets

type S2CServerInfoPacket struct {
	Hello string `json:"hello"`
}

func init() {
	mustRegisterPacket(
		S2CPongPacketType,
		func() Packet {
			return &S2CServerInfoPacket{}
		},
		func(packet Packet) bool {
			value, ok := packet.(*S2CServerInfoPacket)
			return ok && value != nil
		},
	)
}

func (*S2CServerInfoPacket) PacketType() PacketType {
	return S2CPongPacketType
}

func (p *S2CServerInfoPacket) Handle(c *Connection) {
	// Handle p.HelloToClient here.
}

func (*S2CServerInfoPacket) isS2C() {}
