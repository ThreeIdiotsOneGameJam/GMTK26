package packets

import (
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/jsonutil"
)

// C2SCancelBuildActionPacket withdraws the pending build aimed at To. The
// target makes cancellation safe if another action has already replaced it.
type C2SCancelBuildActionPacket struct {
	Round int32    `json:"round"`
	To    game.Hex `json:"to"`
}

func init() {
	mustRegisterPacket(
		C2SCancelBuildActionPacketType,
		func() Packet { return &C2SCancelBuildActionPacket{} },
		func(packet Packet) bool {
			_, ok := packet.(*C2SCancelBuildActionPacket)
			return ok
		},
	)
}

func (*C2SCancelBuildActionPacket) PacketType() PacketType {
	return C2SCancelBuildActionPacketType
}
func (*C2SCancelBuildActionPacket) isC2S() {}

func (p *C2SCancelBuildActionPacket) UnmarshalJSON(data []byte) error {
	type payload struct {
		Round *int32    `json:"round"`
		To    *game.Hex `json:"to"`
	}
	var decoded payload
	if err := jsonutil.DecodeStrict(data, &decoded); err != nil {
		return err
	}
	if decoded.Round == nil {
		return errMissingField("c2s_cancel_build_action", "round")
	}
	if decoded.To == nil {
		return errMissingField("c2s_cancel_build_action", "to")
	}
	p.Round = *decoded.Round
	p.To = *decoded.To
	return nil
}
