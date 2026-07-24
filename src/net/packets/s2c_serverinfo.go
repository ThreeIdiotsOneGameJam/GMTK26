package packets

type S2CServerInfoPacket struct {
	Hello string `json:"hello"`
}

func init() {
	mustRegisterPacket(
		S2CServerInfoPacketType,
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
	return S2CServerInfoPacketType
}

func (p *S2CServerInfoPacket) Handle(c *Connection) {
	// Handle p.HelloToClient here.
}

func (*S2CServerInfoPacket) isS2C() {}
