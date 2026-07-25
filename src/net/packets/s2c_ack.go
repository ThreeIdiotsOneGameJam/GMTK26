package packets

import "github.com/threeidiotsonegamejam/gmtk26/src/util/jsonutil"

type S2CAckPacket struct {
	OK bool `json:"ok"`
}

func init() {
	mustRegisterPacket(
		S2CAckPacketType,
		func() Packet { return &S2CAckPacket{} },
		func(packet Packet) bool {
			_, ok := packet.(*S2CAckPacket)
			return ok
		},
	)
}

func (*S2CAckPacket) PacketType() PacketType { return S2CAckPacketType }
func (*S2CAckPacket) isS2C()                 {}

func (p *S2CAckPacket) UnmarshalJSON(data []byte) error {
	type ackPayload struct {
		OK *bool `json:"ok"`
	}
	var payload ackPayload
	if err := jsonutil.DecodeStrict(data, &payload); err != nil {
		return err
	}
	if payload.OK == nil {
		return errMissingField("ack", "ok")
	}
	p.OK = *payload.OK
	return nil
}
