package packets

import "github.com/threeidiotsonegamejam/gmtk26/src/game"

type C2SCreateGamePacket struct {
	Public     bool  `json:"public"`
	MaxPlayers uint8 `json:"max_players"`
	Seed       int64 `json:"seed"`
}

type C2SJoinGamePacket struct {
	// An empty code requests any available public game.
	GameCode string `json:"game_code"`
}

type C2SLeaveGamePacket struct{}

type S2CGameJoinedPacket struct {
	Game game.Game `json:"game"`
}

type S2CGameUpdatePacket struct {
	Game game.Game `json:"game"`
}

type S2CGameRejectedPacket struct {
	Operation string `json:"operation"`
	Message   string `json:"message"`
}

type S2CGameClosedPacket struct {
	GameID uint64 `json:"game_id"`
	Reason string `json:"reason"`
}

func init() {
	mustRegisterPacket(C2SCreateGamePacketType, func() Packet {
		return &C2SCreateGamePacket{}
	}, func(packet Packet) bool {
		value, ok := packet.(*C2SCreateGamePacket)
		return ok && value != nil
	})
	mustRegisterPacket(C2SJoinGamePacketType, func() Packet {
		return &C2SJoinGamePacket{}
	}, func(packet Packet) bool {
		value, ok := packet.(*C2SJoinGamePacket)
		return ok && value != nil
	})
	mustRegisterPacket(C2SLeaveGamePacketType, func() Packet {
		return &C2SLeaveGamePacket{}
	}, func(packet Packet) bool {
		value, ok := packet.(*C2SLeaveGamePacket)
		return ok && value != nil
	})
	mustRegisterPacket(S2CGameJoinedPacketType, func() Packet {
		return &S2CGameJoinedPacket{}
	}, func(packet Packet) bool {
		value, ok := packet.(*S2CGameJoinedPacket)
		return ok && value != nil
	})
	mustRegisterPacket(S2CGameUpdatePacketType, func() Packet {
		return &S2CGameUpdatePacket{}
	}, func(packet Packet) bool {
		value, ok := packet.(*S2CGameUpdatePacket)
		return ok && value != nil
	})
	mustRegisterPacket(S2CGameRejectedPacketType, func() Packet {
		return &S2CGameRejectedPacket{}
	}, func(packet Packet) bool {
		value, ok := packet.(*S2CGameRejectedPacket)
		return ok && value != nil
	})
	mustRegisterPacket(S2CGameClosedPacketType, func() Packet {
		return &S2CGameClosedPacket{}
	}, func(packet Packet) bool {
		value, ok := packet.(*S2CGameClosedPacket)
		return ok && value != nil
	})
}

func (*C2SCreateGamePacket) PacketType() PacketType { return C2SCreateGamePacketType }
func (*C2SCreateGamePacket) isC2S()                 {}

func (*C2SJoinGamePacket) PacketType() PacketType { return C2SJoinGamePacketType }
func (*C2SJoinGamePacket) isC2S()                 {}

func (*C2SLeaveGamePacket) PacketType() PacketType { return C2SLeaveGamePacketType }
func (*C2SLeaveGamePacket) isC2S()                 {}

func (*S2CGameJoinedPacket) PacketType() PacketType { return S2CGameJoinedPacketType }
func (*S2CGameJoinedPacket) isS2C()                 {}

func (*S2CGameUpdatePacket) PacketType() PacketType { return S2CGameUpdatePacketType }
func (*S2CGameUpdatePacket) isS2C()                 {}

func (*S2CGameRejectedPacket) PacketType() PacketType { return S2CGameRejectedPacketType }
func (*S2CGameRejectedPacket) isS2C()                 {}

func (*S2CGameClosedPacket) PacketType() PacketType { return S2CGameClosedPacketType }
func (*S2CGameClosedPacket) isS2C()                 {}
