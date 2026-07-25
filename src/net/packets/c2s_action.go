package packets

import (
	"fmt"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/jsonutil"
)

type C2SActionPacket struct {
	Round    int32                       `json:"round"`
	Type     game.ActionType             `json:"type"`
	Build    *game.BuildActionPayload    `json:"build,omitempty"`
	Dispatch *game.DispatchActionPayload `json:"dispatch,omitempty"`
}

func init() {
	mustRegisterPacket(
		C2SActionPacketType,
		func() Packet { return &C2SActionPacket{} },
		func(packet Packet) bool {
			_, ok := packet.(*C2SActionPacket)
			return ok
		},
	)
}

func (*C2SActionPacket) PacketType() PacketType { return C2SActionPacketType }
func (*C2SActionPacket) isC2S()                 {}

func (p *C2SActionPacket) UnmarshalJSON(data []byte) error {
	type actionPayload struct {
		Round    *int32                      `json:"round"`
		Type     *game.ActionType            `json:"type"`
		Build    *game.BuildActionPayload    `json:"build,omitempty"`
		Dispatch *game.DispatchActionPayload `json:"dispatch,omitempty"`
	}
	var payload actionPayload
	if err := jsonutil.DecodeStrict(data, &payload); err != nil {
		return err
	}
	if payload.Round == nil {
		return errMissingField("c2s_action", "round")
	}
	if payload.Type == nil {
		return errMissingField("c2s_action", "type")
	}

	switch *payload.Type {
	case game.ActionPass:
	case game.ActionBuild:
		if payload.Build == nil {
			return fmt.Errorf("c2s_action: type build requires build payload")
		}
	case game.ActionDispatch:
		if payload.Dispatch == nil {
			return fmt.Errorf("c2s_action: type dispatch requires dispatch payload")
		}
	default:
		return fmt.Errorf("c2s_action: unknown action type %d", *payload.Type)
	}

	p.Round = *payload.Round
	p.Type = *payload.Type
	p.Build = payload.Build
	p.Dispatch = payload.Dispatch
	return nil
}
