package server

import (
	"sync"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

// instanceManager tracks running GameInstances, as opposed to GameManager
// which handles the pre-game lobby lifecycle.
type instanceManager struct {
	mu     sync.RWMutex
	games  map[uint64]*GameInstance
	nextID uint64
}

var GameInstances = &instanceManager{
	games: make(map[uint64]*GameInstance),
}

func (gm *instanceManager) StartGame(clients []*Client, maxPlayers uint8, seed int64) *GameInstance {
	gm.mu.Lock()
	id := gm.nextID
	gm.nextID++
	gm.mu.Unlock()

	g := &game.Game{
		GameID:      id,
		Multiplayer: maxPlayers > 1,
		MaxPlayers:  maxPlayers,
		Map: game.Map{
			Seed: seed,
		},
	}

	for i := uint8(0); i < 4; i++ {
		if i < maxPlayers {
			g.Factions[i] = game.Faction{
				Index:     int(i),
				Resources: make(game.Resources),
			}
		} else {
			g.Factions[i] = game.Faction{
				Index:     int(i),
				AI:        true,
				Resources: make(game.Resources),
			}
		}
	}

	gi := NewGameInstance(id, g, clients)

	gm.mu.Lock()
	gm.games[id] = gi
	gm.mu.Unlock()

	for _, c := range clients {
		if c != nil {
			c.JoinGame(gi)
		}
	}

	go gi.Run()

	return gi
}

func (gm *instanceManager) GameForClient(client *Client) *GameInstance {
	if client == nil {
		return nil
	}
	return client.GameInstance()
}

func (gm *instanceManager) RemoveGame(id uint64) {
	gm.mu.Lock()
	delete(gm.games, id)
	gm.mu.Unlock()
}
