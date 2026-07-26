package packets

import (
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/jsonutil"
)

type S2CGameStatePacket struct {
	Round        int32                `json:"round"`
	Deadline     int64                `json:"deadline"`
	Map          game.Map             `json:"map"`
	Coins        int32                `json:"coins"`
	Points       int32                `json:"points"`
	Resources    game.Resources       `json:"resources"`
	Orders       []game.MovementOrder `json:"orders"`
	Result       *game.ActionResult   `json:"result,omitempty"`
	Movements    []game.MovementEvent `json:"movements"`
	AttackOrders []game.AttackOrder   `json:"attack_orders"`
	AttackEvents []game.AttackEvent   `json:"attack_events"`
}

func init() {
	mustRegisterPacket(
		S2CGameStatePacketType,
		func() Packet { return &S2CGameStatePacket{} },
		func(packet Packet) bool {
			_, ok := packet.(*S2CGameStatePacket)
			return ok
		},
	)
}

func (*S2CGameStatePacket) PacketType() PacketType { return S2CGameStatePacketType }
func (*S2CGameStatePacket) isS2C()                 {}

func (p *S2CGameStatePacket) UnmarshalJSON(data []byte) error {
	type statePayload struct {
		Round        *int32                `json:"round"`
		Deadline     *int64                `json:"deadline"`
		Map          *game.Map             `json:"map"`
		Coins        *int32                `json:"coins"`
		Points       *int32                `json:"points"`
		Resources    *game.Resources       `json:"resources"`
		Orders       *[]game.MovementOrder `json:"orders"`
		Result       *game.ActionResult    `json:"result,omitempty"`
		Movements    *[]game.MovementEvent `json:"movements"`
		AttackOrders *[]game.AttackOrder   `json:"attack_orders"`
		AttackEvents *[]game.AttackEvent   `json:"attack_events"`
	}
	var payload statePayload
	if err := jsonutil.DecodeStrict(data, &payload); err != nil {
		return err
	}
	if payload.Round == nil {
		return errMissingField("s2c_game_state", "round")
	}
	if payload.Deadline == nil {
		return errMissingField("s2c_game_state", "deadline")
	}
	if payload.Map == nil {
		return errMissingField("s2c_game_state", "map")
	}
	if payload.Coins == nil {
		return errMissingField("s2c_game_state", "coins")
	}
	if payload.Points == nil {
		return errMissingField("s2c_game_state", "points")
	}
	if payload.Resources == nil {
		return errMissingField("s2c_game_state", "resources")
	}
	if payload.Orders == nil {
		return errMissingField("s2c_game_state", "orders")
	}
	if payload.Movements == nil {
		return errMissingField("s2c_game_state", "movements")
	}
	if payload.AttackOrders == nil {
		return errMissingField("s2c_game_state", "attack_orders")
	}
	if payload.AttackEvents == nil {
		return errMissingField("s2c_game_state", "attack_events")
	}

	p.Round = *payload.Round
	p.Deadline = *payload.Deadline
	p.Map = *payload.Map
	p.Coins = *payload.Coins
	p.Points = *payload.Points
	p.Resources = *payload.Resources
	p.Orders = *payload.Orders
	p.Result = payload.Result
	p.Movements = *payload.Movements
	p.AttackOrders = *payload.AttackOrders
	p.AttackEvents = *payload.AttackEvents
	return nil
}
