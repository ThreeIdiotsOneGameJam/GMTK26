package packets

import (
	"fmt"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/jsonutil"
)

type C2SActionPacket struct {
	Round   int32                      `json:"round"`
	Type    game.ActionType            `json:"type"`
	Build   *game.BuildActionPayload   `json:"build,omitempty"`
	Move    *game.MoveActionPayload    `json:"move,omitempty"`
	Recruit *game.RecruitActionPayload `json:"recruit,omitempty"`
	Attack  *game.AttackActionPayload  `json:"attack,omitempty"`
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
		Round   *int32                     `json:"round"`
		Type    *game.ActionType           `json:"type"`
		Build   *game.BuildActionPayload   `json:"build,omitempty"`
		Move    *game.MoveActionPayload    `json:"move,omitempty"`
		Recruit *game.RecruitActionPayload `json:"recruit,omitempty"`
		Attack  *game.AttackActionPayload  `json:"attack,omitempty"`
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
		if payload.Build != nil || payload.Move != nil || payload.Recruit != nil || payload.Attack != nil {
			return fmt.Errorf("c2s_action: type pass does not accept a payload")
		}
	case game.ActionBuild:
		if payload.Build == nil {
			return fmt.Errorf("c2s_action: type build requires build payload")
		}
		if payload.Move != nil || payload.Recruit != nil || payload.Attack != nil {
			return fmt.Errorf("c2s_action: type build has mismatched payload")
		}
	case game.ActionMove:
		if payload.Move == nil {
			return fmt.Errorf("c2s_action: type move requires move payload")
		}
		if payload.Build != nil || payload.Recruit != nil || payload.Attack != nil {
			return fmt.Errorf("c2s_action: type move has mismatched payload")
		}
	case game.ActionRecruit:
		if payload.Recruit == nil {
			return fmt.Errorf("c2s_action: type recruit requires recruit payload")
		}
		if payload.Build != nil || payload.Move != nil || payload.Attack != nil {
			return fmt.Errorf("c2s_action: type recruit has mismatched payload")
		}
	case game.ActionAttack:
		if payload.Attack == nil {
			return fmt.Errorf("c2s_action: type attack requires attack payload")
		}
		if payload.Build != nil || payload.Move != nil || payload.Recruit != nil {
			return fmt.Errorf("c2s_action: type attack has mismatched payload")
		}
	default:
		return fmt.Errorf("c2s_action: unknown action type %d", *payload.Type)
	}

	p.Round = *payload.Round
	p.Type = *payload.Type
	p.Build = payload.Build
	p.Move = payload.Move
	p.Recruit = payload.Recruit
	p.Attack = payload.Attack
	return nil
}
