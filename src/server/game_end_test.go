package server

import (
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
)

type gameEndPacketRecorder struct {
	packet *packets.S2CGameEndPacket
}

func (r *gameEndPacketRecorder) SendPacket(packet packets.Packet) error {
	if end, ok := packet.(*packets.S2CGameEndPacket); ok {
		r.packet = end
	}
	return nil
}

func TestGameEndSurvivorWinsAheadOfHigherScoringEliminatedFaction(t *testing.T) {
	g := &game.Game{}
	g.Factions[0] = game.Faction{
		Player: &game.Player{PlayerName: "Survivor"},
		Points: 10,
		Alive:  true,
	}
	g.Factions[1] = game.Faction{
		Player: &game.Player{PlayerName: "Eliminated"},
		Points: 50,
		Alive:  false,
	}

	recorder := &gameEndPacketRecorder{}
	client := NewClient(recorder)
	gi := NewGameInstance(1, g, []*Client{client})
	client.JoinGame(gi)

	gi.broadcastGameEnd()

	if recorder.packet == nil {
		t.Fatal("game end packet was not sent")
	}
	if recorder.packet.WinnerFaction != 0 {
		t.Fatalf("winner faction = %d, want surviving faction 0", recorder.packet.WinnerFaction)
	}
	if recorder.packet.WinnerName != "Survivor" {
		t.Fatalf("winner name = %q, want %q", recorder.packet.WinnerName, "Survivor")
	}
	if len(recorder.packet.Rankings) == 0 || recorder.packet.Rankings[0].FactionIdx != 0 {
		t.Fatalf("rankings = %+v, want surviving faction first", recorder.packet.Rankings)
	}
}

func TestGameEndUsesScoreWhenMultipleFactionsRemain(t *testing.T) {
	g := &game.Game{}
	g.Factions[0] = game.Faction{Points: 10, Alive: true}
	g.Factions[1] = game.Faction{Points: 50, Alive: true}

	recorder := &gameEndPacketRecorder{}
	client := NewClient(recorder)
	gi := NewGameInstance(1, g, []*Client{client})
	client.JoinGame(gi)

	gi.broadcastGameEnd()

	if recorder.packet == nil {
		t.Fatal("game end packet was not sent")
	}
	if recorder.packet.WinnerFaction != 1 {
		t.Fatalf("winner faction = %d, want highest-scoring faction 1", recorder.packet.WinnerFaction)
	}
	if len(recorder.packet.Rankings) == 0 || recorder.packet.Rankings[0].FactionIdx != 1 {
		t.Fatalf("rankings = %+v, want highest-scoring faction first", recorder.packet.Rankings)
	}
}

func TestGameEndRanksLivingFactionsAheadOfEliminatedScoreLeader(t *testing.T) {
	g := &game.Game{}
	g.Factions[0] = game.Faction{Points: 10, Alive: true}
	g.Factions[1] = game.Faction{Points: 50, Alive: true}
	g.Factions[2] = game.Faction{Points: 500, Alive: false}

	recorder := &gameEndPacketRecorder{}
	client := NewClient(recorder)
	gi := NewGameInstance(1, g, []*Client{client})
	client.JoinGame(gi)

	gi.broadcastGameEnd()

	if recorder.packet.WinnerFaction != 1 {
		t.Fatalf("winner faction = %d, want living score leader 1", recorder.packet.WinnerFaction)
	}
	if got := recorder.packet.Rankings[2].FactionIdx; got != 2 {
		t.Fatalf("eliminated score leader ranked at faction %d position, want third", got)
	}
}

func TestGameEndUsesDrawForTiedEligibleLeaders(t *testing.T) {
	g := &game.Game{}
	g.Factions[0] = game.Faction{Points: 50, Alive: true}
	g.Factions[1] = game.Faction{Points: 50, Alive: true}

	recorder := &gameEndPacketRecorder{}
	client := NewClient(recorder)
	gi := NewGameInstance(1, g, []*Client{client})
	client.JoinGame(gi)

	gi.broadcastGameEnd()

	if recorder.packet.WinnerFaction != -1 {
		t.Fatalf("winner faction = %d, want draw -1", recorder.packet.WinnerFaction)
	}
	if recorder.packet.WinnerName != "" {
		t.Fatalf("draw winner name = %q, want empty", recorder.packet.WinnerName)
	}
}
