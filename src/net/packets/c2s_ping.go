package packets

type C2SPingPacket struct {
	HelloToServer string `json:"hellotoserver"`
}

func init() {
	mustRegisterPacket(
		C2SPingPacketType,
		func() Packet {
			return &C2SPingPacket{}
		},
		func(packet Packet) bool {
			value, ok := packet.(*C2SPingPacket)
			return ok && value != nil
		},
	)
}

func (*C2SPingPacket) PacketType() PacketType {
	return C2SPingPacketType
}

func (p *C2SPingPacket) Handle() {
	// Handle p.HelloToServer here.
}

func (*C2SPingPacket) isC2S() {}
