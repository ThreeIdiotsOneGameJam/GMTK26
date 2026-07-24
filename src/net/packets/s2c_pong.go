package packets

type S2CPongPacket struct {
	HelloToClient string `json:"hellotoclient"`
}

func init() {
	mustRegisterPacket(
		S2CPongPacketType,
		func() Packet {
			return &S2CPongPacket{}
		},
		func(packet Packet) bool {
			value, ok := packet.(*S2CPongPacket)
			return ok && value != nil
		},
	)
}

func (*S2CPongPacket) PacketType() PacketType {
	return S2CPongPacketType
}

func (p *S2CPongPacket) Handle() {
	// Handle p.HelloToClient here.
}

func (*S2CPongPacket) isS2C() {}
