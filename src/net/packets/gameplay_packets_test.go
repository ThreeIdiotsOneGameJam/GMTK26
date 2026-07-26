package packets

import (
	"strings"
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

func TestGameplayPacketRoundTrips(t *testing.T) {
	tests := []Packet{
		&C2SActionPacket{
			Round: 3,
			Type:  game.ActionMove,
			Move: &game.MoveActionPayload{
				From: game.NewHex(1, 2),
				To:   game.NewHex(4, 5),
			},
		},
		&C2SActionPacket{
			Round: 3,
			Type:  game.ActionRecruit,
			Recruit: &game.RecruitActionPayload{
				From: game.NewHex(1, 2),
				To:   game.NewHex(2, 2),
				Unit: game.UnitScout,
			},
		},
		&C2SActionPacket{
			Round: 3,
			Type:  game.ActionBuild,
			Build: &game.BuildActionPayload{
				From:     game.NewHex(1, 2),
				To:       game.NewHex(2, 2),
				Building: game.BuildingBarracks,
			},
		},
		&C2SActionPacket{
			Round: 3,
			Type:  game.ActionAttack,
			Attack: &game.AttackActionPayload{
				From: game.NewHex(1, 2),
				To:   game.NewHex(2, 2),
			},
		},
		&C2SCancelMovementOrderPacket{
			Round: 3,
			From:  game.NewHex(1, 2),
		},
		&C2SCancelBuildActionPacket{
			Round: 3,
			To:    game.NewHex(2, 2),
		},
		&S2CGameStartPacket{
			FactionIdx:  0,
			Map:         game.Map{},
			Coins:       100,
			Points:      0,
			Resources:   game.Resources{},
			Round:       1,
			Deadline:    100,
			GameEndTime: 200,
			Orders: []game.MovementOrder{{
				Current:     game.NewHex(1, 2),
				Destination: game.NewHex(4, 5),
			}},
		},
		&S2CGameStatePacket{
			Round:     4,
			Deadline:  100,
			Map:       game.Map{},
			Coins:     90,
			Points:    2,
			Resources: game.Resources{},
			Orders: []game.MovementOrder{{
				Current:     game.NewHex(2, 2),
				Destination: game.NewHex(5, 5),
			}},
			Result: &game.ActionResult{
				Round:   3,
				Type:    game.ActionMove,
				Status:  game.ActionResultSucceeded,
				From:    game.NewHex(1, 2),
				To:      game.NewHex(5, 5),
				Message: "Movement order queued",
			},
			Movements: []game.MovementEvent{{
				Unit:  game.UnitScout,
				Owner: 0,
				Path: []game.Hex{
					game.NewHex(1, 2),
					game.NewHex(2, 2),
				},
			}},
		},
		&S2CGameEndPacket{
			WinnerFaction: 1,
			WinnerName:    "Winner",
			Rankings: []RankEntry{
				{FactionIdx: 1, PlayerName: "Winner", Points: 12, Alive: true},
				{FactionIdx: 0, PlayerName: "Runner-up", Points: 8, Alive: false},
			},
		},
	}

	for _, packet := range tests {
		encoded, err := Serialize(packet)
		if err != nil {
			t.Fatalf("serialize %T: %v", packet, err)
		}
		decoded, err := Deserialize(encoded)
		if err != nil {
			t.Fatalf("deserialize %T: %v", packet, err)
		}
		if decoded.PacketType() != packet.PacketType() {
			t.Fatalf("decoded type = %d, want %d", decoded.PacketType(), packet.PacketType())
		}
	}
}

func TestActionRejectsMismatchedPayload(t *testing.T) {
	packet := &C2SActionPacket{
		Round: 1,
		Type:  game.ActionMove,
		Move: &game.MoveActionPayload{
			From: game.NewHex(0, 0),
			To:   game.NewHex(0, 1),
		},
		Attack: &game.AttackActionPayload{
			From: game.NewHex(0, 0),
			To:   game.NewHex(0, 1),
		},
	}
	encoded, err := Serialize(packet)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Deserialize(encoded); err == nil || !strings.Contains(err.Error(), "mismatched") {
		t.Fatalf("error = %v, want mismatched payload", err)
	}
}

func TestGameStartRoundTripPreservesTimerAndOrders(t *testing.T) {
	want := &S2CGameStartPacket{
		FactionIdx:  2,
		Map:         game.Map{},
		Coins:       90,
		Points:      4,
		Resources:   game.Resources{},
		Round:       3,
		Deadline:    100,
		GameEndTime: 200,
		Orders: []game.MovementOrder{{
			Current:     game.NewHex(1, 2),
			Destination: game.NewHex(4, 5),
		}},
	}
	encoded, err := Serialize(want)
	if err != nil {
		t.Fatal(err)
	}
	packet, err := Deserialize(encoded)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := packet.(*S2CGameStartPacket)
	if !ok {
		t.Fatalf("decoded %T, want *S2CGameStartPacket", packet)
	}
	if got.GameEndTime != want.GameEndTime {
		t.Fatalf("game end time = %d, want %d", got.GameEndTime, want.GameEndTime)
	}
	if len(got.Orders) != 1 || got.Orders[0] != want.Orders[0] {
		t.Fatalf("orders = %v, want %v", got.Orders, want.Orders)
	}
}

func TestProtocolV1IsRejected(t *testing.T) {
	data := []byte(`{"version":1,"type":10,"data":{"round":1,"type":1}}`)
	if _, err := Deserialize(data); err == nil || !strings.Contains(err.Error(), "protocol version 1") {
		t.Fatalf("error = %v, want protocol version rejection", err)
	}
}
