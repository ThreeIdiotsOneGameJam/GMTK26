package packets

import (
	"github.com/threeidiotsonegamejam/gmtk26/src/util/jsonutil"
)

type C2SUpdatePlayerNamePacket struct {
	PlayerName string `json:"player_name"`
}

func init() {
	mustRegisterPacket(
		C2SUpdatePlayerNamePacketType,
		func() Packet { return &C2SUpdatePlayerNamePacket{} },
		func(packet Packet) bool {
			value, ok := packet.(*C2SUpdatePlayerNamePacket)
			return ok && value != nil
		},
	)
}

func (*C2SUpdatePlayerNamePacket) PacketType() PacketType {
	return C2SUpdatePlayerNamePacketType
}
func (*C2SUpdatePlayerNamePacket) isC2S() {}

func (p *C2SUpdatePlayerNamePacket) UnmarshalJSON(data []byte) error {
	type payload struct {
		PlayerName *string `json:"player_name"`
	}
	var decoded payload
	if err := jsonutil.DecodeStrict(data, &decoded); err != nil {
		return err
	}
	if decoded.PlayerName == nil {
		return errMissingField("c2s_update_player_name", "player_name")
	}
	p.PlayerName = *decoded.PlayerName
	return nil
}
