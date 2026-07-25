package server

import (
	"sync"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

// instanceManager tracks running GameInstances, as opposed to LobbyManager
// which tracks lobbies that have not started yet.
type instanceManager struct {
	mu    sync.RWMutex
	games map[uint64]*GameInstance
}

var GameInstances = &instanceManager{
	games: make(map[uint64]*GameInstance),
}

// Prepare registers a game instance and attaches clients without starting the
// loop. Used by LobbyManager while it still holds its lock so concurrent
// leave/create requests observe GameInstance immediately.
func (gm *instanceManager) Prepare(g *game.Game, clients []*Client) *GameInstance {
	gi := NewGameInstance(g.GameID, g, clients)

	gm.mu.Lock()
	gm.games[g.GameID] = gi
	gm.mu.Unlock()

	for _, c := range clients {
		if c != nil {
			c.JoinGame(gi)
		}
	}

	return gi
}

// StartGame launches the game loop for a dissolved lobby. clients must have
// one slot per faction (nil for AI and unfilled slots), as produced by
// clientsByFaction. Faction gameplay fields are initialized by the game
// loop itself; lobby-assigned data such as faction players is preserved.
func (gm *instanceManager) StartGame(g *game.Game, clients []*Client) *GameInstance {
	gi := gm.Prepare(g, clients)
	go gi.Run()
	return gi
}

func (gm *instanceManager) RemoveGame(id uint64) {
	gm.mu.Lock()
	delete(gm.games, id)
	gm.mu.Unlock()
}
