package packets

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/jsonutil"
)

type S2CConnectAcceptedPacket struct {
	ClientID   game.ClientID `json:"client_id"`
	Persistent bool          `json:"persistent"`
}

func init() {
	mustRegisterPacket(
		S2CConnectAcceptedPacketType,
		func() Packet {
			return &S2CConnectAcceptedPacket{}
		},
		func(packet Packet) bool {
			value, ok := packet.(*S2CConnectAcceptedPacket)
			return ok && value != nil
		},
	)
}

func (*S2CConnectAcceptedPacket) PacketType() PacketType {
	return S2CConnectAcceptedPacketType
}

func (*S2CConnectAcceptedPacket) isS2C() {}

func (p *S2CConnectAcceptedPacket) UnmarshalJSON(data []byte) error {
	type acceptedPayload struct {
		ClientID   *game.ClientID `json:"client_id"`
		Persistent *bool          `json:"persistent"`
	}

	var payload acceptedPayload
	if err := jsonutil.DecodeStrict(data, &payload); err != nil {
		return err
	}
	if payload.ClientID == nil {
		return fmt.Errorf("connect accepted packet: missing or null client_id")
	}
	if payload.Persistent == nil {
		return fmt.Errorf("connect accepted packet: missing or null persistent")
	}
	rawClientID := string(*payload.ClientID)
	clientID, err := uuid.Parse(rawClientID)
	if err != nil || clientID == uuid.Nil || clientID.String() != rawClientID {
		return fmt.Errorf(
			"connect accepted packet: client_id must be a canonical, non-nil UUID",
		)
	}

	p.ClientID = *payload.ClientID
	p.Persistent = *payload.Persistent
	return nil
}
