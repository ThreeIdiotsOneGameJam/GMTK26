package server

import (
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

func TestNewGameInstanceLimitsClientsToFactionSlots(t *testing.T) {
	g := &game.Game{GameID: 1}
	clients := make([]*Client, len(g.Factions)+1)
	for i := range clients {
		clients[i] = &Client{}
	}

	gi := NewGameInstance(g.GameID, g, clients)

	if len(gi.clients) != len(g.Factions) {
		t.Fatalf("client slots = %d, want %d", len(gi.clients), len(g.Factions))
	}
	if _, exists := gi.factionClients[clients[len(g.Factions)]]; exists {
		t.Fatal("overflow client was assigned a faction")
	}
}

func TestInstanceManagerDoesNotAttachOverflowClient(t *testing.T) {
	g := &game.Game{GameID: 1}
	clients := make([]*Client, len(g.Factions)+1)
	for i := range clients {
		clients[i] = &Client{}
	}
	manager := &instanceManager{games: make(map[uint64]*GameInstance)}

	gi := manager.Prepare(g, clients)

	for i, client := range clients[:len(g.Factions)] {
		if client.GameInstance() != gi {
			t.Fatalf("client %d was not attached to the game", i)
		}
	}
	if clients[len(g.Factions)].GameInstance() != nil {
		t.Fatal("overflow client was attached to the game")
	}
}

func TestAvailableFactionClampsMalformedMaxPlayers(t *testing.T) {
	state := game.Game{MaxPlayers: 255}
	for i := range state.Factions {
		state.Factions[i].AI = true
	}

	if got := availableFaction(state); got != -1 {
		t.Fatalf("available faction = %d, want -1", got)
	}
}
