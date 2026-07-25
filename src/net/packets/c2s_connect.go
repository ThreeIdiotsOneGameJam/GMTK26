package packets

import (
	"fmt"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/jsonutil"
)

type C2SConnectPacket struct {
	Player game.Player `json:"player"`
}

func init() {
	mustRegisterPacket(
		C2SConnectPacketType,
		func() Packet {
			return &C2SConnectPacket{}
		},
		func(packet Packet) bool {
			value, ok := packet.(*C2SConnectPacket)
			return ok && value != nil
		},
	)
}

func (*C2SConnectPacket) PacketType() PacketType {
	return C2SConnectPacketType
}

func (*C2SConnectPacket) isC2S() {}

func (p *C2SConnectPacket) UnmarshalJSON(data []byte) error {
	type connectPayload struct {
		Player *game.Player `json:"player"`
	}

	var payload connectPayload
	if err := jsonutil.DecodeStrict(data, &payload); err != nil {
		return err
	}
	if payload.Player == nil {
		return fmt.Errorf("connect packet: missing or null player")
	}

	p.Player = *payload.Player
	return nil
}
