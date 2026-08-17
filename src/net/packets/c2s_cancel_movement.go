package packets

import (
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/jsonutil"
)

type C2SCancelMovementOrderPacket struct {
	Round int32    `json:"round"`
	From  game.Hex `json:"from"`
}

func init() {
	mustRegisterPacket(
		C2SCancelMovementOrderPacketType,
		func() Packet { return &C2SCancelMovementOrderPacket{} },
		func(packet Packet) bool {
			_, ok := packet.(*C2SCancelMovementOrderPacket)
			return ok
		},
	)
}

func (*C2SCancelMovementOrderPacket) PacketType() PacketType {
	return C2SCancelMovementOrderPacketType
}
func (*C2SCancelMovementOrderPacket) isC2S() {}

func (p *C2SCancelMovementOrderPacket) UnmarshalJSON(data []byte) error {
	type payload struct {
		Round *int32    `json:"round"`
		From  *game.Hex `json:"from"`
	}
	var decoded payload
	if err := jsonutil.DecodeStrict(data, &decoded); err != nil {
		return err
	}
	if decoded.Round == nil {
		return errMissingField("c2s_cancel_movement_order", "round")
	}
	if decoded.From == nil {
		return errMissingField("c2s_cancel_movement_order", "from")
	}
	p.Round = *decoded.Round
	p.From = *decoded.From
	return nil
}
