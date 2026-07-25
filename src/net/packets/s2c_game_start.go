package packets

import (
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/jsonutil"
)

type S2CGameStartPacket struct {
	FactionIdx int            `json:"faction_idx"`
	Map        game.Map       `json:"map"`
	Coins      int32          `json:"coins"`
	Points     int32          `json:"points"`
	Resources  game.Resources `json:"resources"`
	Round      int32          `json:"round"`
	Deadline   int64          `json:"deadline"`
}

func init() {
	mustRegisterPacket(
		S2CGameStartPacketType,
		func() Packet { return &S2CGameStartPacket{} },
		func(packet Packet) bool {
			_, ok := packet.(*S2CGameStartPacket)
			return ok
		},
	)
}

func (*S2CGameStartPacket) PacketType() PacketType { return S2CGameStartPacketType }
func (*S2CGameStartPacket) isS2C()                 {}

func (p *S2CGameStartPacket) UnmarshalJSON(data []byte) error {
	type startPayload struct {
		FactionIdx *int            `json:"faction_idx"`
		Map        *game.Map       `json:"map"`
		Coins      *int32          `json:"coins"`
		Points     *int32          `json:"points"`
		Resources  *game.Resources `json:"resources"`
		Round      *int32          `json:"round"`
		Deadline   *int64          `json:"deadline"`
	}
	var payload startPayload
	if err := jsonutil.DecodeStrict(data, &payload); err != nil {
		return err
	}
	if payload.FactionIdx == nil {
		return errMissingField("s2c_game_start", "faction_idx")
	}
	if payload.Map == nil {
		return errMissingField("s2c_game_start", "map")
	}
	if payload.Coins == nil {
		return errMissingField("s2c_game_start", "coins")
	}
	if payload.Points == nil {
		return errMissingField("s2c_game_start", "points")
	}
	if payload.Resources == nil {
		return errMissingField("s2c_game_start", "resources")
	}
	if payload.Round == nil {
		return errMissingField("s2c_game_start", "round")
	}
	if payload.Deadline == nil {
		return errMissingField("s2c_game_start", "deadline")
	}

	p.FactionIdx = *payload.FactionIdx
	p.Map = *payload.Map
	p.Coins = *payload.Coins
	p.Points = *payload.Points
	p.Resources = *payload.Resources
	p.Round = *payload.Round
	p.Deadline = *payload.Deadline
	return nil
}
