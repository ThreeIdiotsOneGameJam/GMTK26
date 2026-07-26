package screens

import (
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/net"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
)

func TestResultPlacement(t *testing.T) {
	rankings := []packets.RankEntry{
		{FactionIdx: 2},
		{FactionIdx: 0},
		{FactionIdx: 1},
	}

	if got := resultPlacement(rankings, 0); got != 2 {
		t.Fatalf("resultPlacement() = %d, want 2", got)
	}
	if got := resultPlacement(rankings, 3); got != 0 {
		t.Fatalf("missing resultPlacement() = %d, want 0", got)
	}
}

func TestOrdinal(t *testing.T) {
	tests := map[int]string{
		1:   "1st",
		2:   "2nd",
		3:   "3rd",
		4:   "4th",
		11:  "11th",
		12:  "12th",
		13:  "13th",
		21:  "21st",
		102: "102nd",
	}
	for value, want := range tests {
		if got := ordinal(value); got != want {
			t.Errorf("ordinal(%d) = %q, want %q", value, got, want)
		}
	}
}

func TestResultEntryNameFallback(t *testing.T) {
	entry := packets.RankEntry{FactionIdx: 2}
	if got := resultEntryName(entry); got != "Faction 3" {
		t.Fatalf("resultEntryName() = %q, want %q", got, "Faction 3")
	}
}

func TestTrimResultName(t *testing.T) {
	if got := trimResultName("A very long kingdom name", 12); got != "A very lo..." {
		t.Fatalf("trimResultName() = %q, want %q", got, "A very lo...")
	}
	if got := trimResultName("short", 12); got != "short" {
		t.Fatalf("short trimResultName() = %q, want %q", got, "short")
	}
}

func TestApplyServerGameEndQueuesResultsScreen(t *testing.T) {
	originalActive := activeScreen
	originalPending := pendingScreen
	originalGameScreen := gameScreen
	originalPrevious := gamePreviousScreen
	originalServerGameActive := serverGameActive
	originalActionsEnabled := gameWorld.Renderer.ActionsEnabled
	t.Cleanup(func() {
		activeScreen = originalActive
		pendingScreen = originalPending
		gameScreen = originalGameScreen
		gamePreviousScreen = originalPrevious
		serverGameActive = originalServerGameActive
		gameWorld.Renderer.ActionsEnabled = originalActionsEnabled
		net.LocalGameState.Reset()
	})

	activeGameScreen := ui.Screen()
	previousScreen := ui.Screen()
	activeScreen = activeGameScreen
	pendingScreen = nil
	gameScreen = activeGameScreen
	gamePreviousScreen = previousScreen
	serverGameActive = true
	gameWorld.Renderer.ActionsEnabled = true
	net.LocalGameState.ApplyStartPacket(&packets.S2CGameStartPacket{FactionIdx: 1})

	ApplyServerGameEnd(&packets.S2CGameEndPacket{
		WinnerFaction: 1,
		WinnerName:    "Local player",
		Rankings: []packets.RankEntry{
			{FactionIdx: 1, PlayerName: "Local player", Points: 12, Alive: true},
		},
	})

	if pendingScreen == nil {
		t.Fatal("game end did not queue a results screen")
	}
	if pendingScreen == activeGameScreen {
		t.Fatal("game end left the live game screen pending")
	}
	if serverGameActive {
		t.Fatal("game remained active after receiving the end packet")
	}
	if gameWorld.Renderer.ActionsEnabled {
		t.Fatal("world actions remained enabled after receiving the end packet")
	}
	if pendingScreen.OnBack == nil {
		t.Fatal("results screen has no route back to the play screen")
	}

	pendingScreen.OnBack()
	if pendingScreen != previousScreen {
		t.Fatal("results screen back action did not queue the previous play screen")
	}
}
