package packets

import (
	"github.com/threeidiotsonegamejam/gmtk26/src/util/jsonutil"
)

type RankEntry struct {
	FactionIdx int    `json:"faction_idx"`
	PlayerName string `json:"player_name"`
	Points     int32  `json:"points"`
	Alive      bool   `json:"alive"`
}

type S2CGameEndPacket struct {
	WinnerFaction int         `json:"winner_faction"`
	WinnerName    string      `json:"winner_name"`
	Rankings      []RankEntry `json:"rankings"`
}

func init() {
	mustRegisterPacket(
		S2CGameEndPacketType,
		func() Packet { return &S2CGameEndPacket{} },
		func(packet Packet) bool {
			_, ok := packet.(*S2CGameEndPacket)
			return ok
		},
	)
}

func (*S2CGameEndPacket) PacketType() PacketType { return S2CGameEndPacketType }
func (*S2CGameEndPacket) isS2C()                 {}

func (p *S2CGameEndPacket) UnmarshalJSON(data []byte) error {
	type endPayload struct {
		WinnerFaction *int         `json:"winner_faction"`
		WinnerName    *string      `json:"winner_name"`
		Rankings      *[]RankEntry `json:"rankings"`
	}
	var payload endPayload
	if err := jsonutil.DecodeStrict(data, &payload); err != nil {
		return err
	}
	if payload.WinnerFaction == nil {
		return errMissingField("s2c_game_end", "winner_faction")
	}
	if payload.WinnerName == nil {
		return errMissingField("s2c_game_end", "winner_name")
	}
	if payload.Rankings == nil {
		return errMissingField("s2c_game_end", "rankings")
	}

	p.WinnerFaction = *payload.WinnerFaction
	p.WinnerName = *payload.WinnerName
	p.Rankings = *payload.Rankings
	return nil
}
