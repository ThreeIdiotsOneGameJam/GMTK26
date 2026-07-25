package server

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
)

const (
	gameCodeLength  = 6
	gameCodeCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	ErrAlreadyInGame = errors.New("client is already in a game")
	ErrGameNotFound  = errors.New("game not found")
	ErrGameFull      = errors.New("game is full")

	GameLifecycle = NewGameManager()
)

type activeGame struct {
	state   game.Game
	clients []*Client
}

type GameManager struct {
	mu          sync.RWMutex
	nextGameID  uint64
	games       map[uint64]*activeGame
	gameCodes   map[string]uint64
	clientGames map[*Client]uint64
}

func NewGameManager() *GameManager {
	return &GameManager{
		games:       make(map[uint64]*activeGame),
		gameCodes:   make(map[string]uint64),
		clientGames: make(map[*Client]uint64),
	}
}

func (m *GameManager) CreateGame(client *Client, public bool, maxPlayers uint8, seed int64) (game.Game, error) {
	if maxPlayers < 2 || maxPlayers > 4 {
		return game.Game{}, fmt.Errorf("max players must be 2, 3, or 4")
	}

	player, identified := client.Player()
	if !identified {
		return game.Game{}, fmt.Errorf("client is not identified")
	}

	m.mu.Lock()
	if _, exists := m.clientGames[client]; exists {
		m.mu.Unlock()
		return game.Game{}, ErrAlreadyInGame
	}

	gameCode, err := m.newGameCodeLocked()
	if err != nil {
		m.mu.Unlock()
		return game.Game{}, err
	}

	m.nextGameID++
	gameID := m.nextGameID
	state := game.Game{
		GameID:      gameID,
		GameCode:    gameCode,
		Public:      public,
		HostID:      player.ClientID,
		Multiplayer: maxPlayers > 1,
		MaxPlayers:  maxPlayers,
		Round:       1,
		Map: game.Map{
			Seed: seed,
		},
	}
	state.Factions[0].Player = playerPointer(player)
	for i := int(maxPlayers); i < len(state.Factions); i++ {
		state.Factions[i].AI = true
	}

	m.games[gameID] = &activeGame{
		state:   state,
		clients: []*Client{client},
	}
	m.gameCodes[gameCode] = gameID
	m.clientGames[client] = gameID
	m.mu.Unlock()

	return state, nil
}

func (m *GameManager) JoinGame(client *Client, code string) (game.Game, error) {
	player, identified := client.Player()
	if !identified {
		return game.Game{}, fmt.Errorf("client is not identified")
	}

	code = strings.ToUpper(strings.TrimSpace(code))

	m.mu.Lock()
	if _, exists := m.clientGames[client]; exists {
		m.mu.Unlock()
		return game.Game{}, ErrAlreadyInGame
	}

	active := m.findJoinableGameLocked(code)
	if active == nil {
		m.mu.Unlock()
		return game.Game{}, ErrGameNotFound
	}

	factionIndex := availableFaction(active.state)
	if factionIndex < 0 {
		m.mu.Unlock()
		return game.Game{}, ErrGameFull
	}

	active.state.Factions[factionIndex].Player = playerPointer(player)
	active.clients = append(active.clients, client)
	m.clientGames[client] = active.state.GameID
	state := active.state
	existingClients := clientsExcept(active.clients, client)
	m.mu.Unlock()

	broadcast(existingClients, &packets.S2CGameUpdatePacket{Game: state})
	return state, nil
}

func (m *GameManager) LeaveGame(client *Client) {
	m.removeClient(client, "host left the game")
}

func (m *GameManager) Disconnect(client *Client) {
	m.removeClient(client, "host disconnected")
}

func (m *GameManager) ActiveGames() []game.Game {
	m.mu.RLock()
	defer m.mu.RUnlock()

	games := make([]game.Game, 0, len(m.games))
	for _, active := range m.games {
		games = append(games, active.state)
	}
	return games
}

func (m *GameManager) removeClient(client *Client, hostReason string) {
	m.mu.Lock()
	gameID, exists := m.clientGames[client]
	if !exists {
		m.mu.Unlock()
		return
	}

	active := m.games[gameID]
	delete(m.clientGames, client)
	if active == nil {
		m.mu.Unlock()
		return
	}

	player, _ := client.Player()
	if active.state.HostID == player.ClientID {
		delete(m.games, gameID)
		delete(m.gameCodes, active.state.GameCode)
		remainingClients := clientsExcept(active.clients, client)
		for _, remaining := range remainingClients {
			delete(m.clientGames, remaining)
		}
		m.mu.Unlock()

		broadcast(remainingClients, &packets.S2CGameClosedPacket{
			GameID: gameID,
			Reason: hostReason,
		})
		return
	}

	for i := range active.state.Factions {
		factionPlayer := active.state.Factions[i].Player
		if factionPlayer != nil && factionPlayer.ClientID == player.ClientID {
			active.state.Factions[i].Player = nil
			break
		}
	}
	active.clients = clientsExcept(active.clients, client)
	state := active.state
	remainingClients := append([]*Client(nil), active.clients...)
	m.mu.Unlock()

	broadcast(remainingClients, &packets.S2CGameUpdatePacket{Game: state})
}

func (m *GameManager) findJoinableGameLocked(code string) *activeGame {
	if code != "" {
		gameID, exists := m.gameCodes[code]
		if !exists {
			return nil
		}
		return m.games[gameID]
	}

	var selected *activeGame
	for _, active := range m.games {
		if !active.state.Public || availableFaction(active.state) < 0 {
			continue
		}
		if selected == nil || active.state.GameID < selected.state.GameID {
			selected = active
		}
	}
	return selected
}

func (m *GameManager) newGameCodeLocked() (string, error) {
	for range 100 {
		bytes := make([]byte, gameCodeLength)
		if _, err := rand.Read(bytes); err != nil {
			return "", fmt.Errorf("generate game code: %w", err)
		}
		for i := range bytes {
			bytes[i] = gameCodeCharset[int(bytes[i])%len(gameCodeCharset)]
		}
		code := string(bytes)
		if _, exists := m.gameCodes[code]; !exists {
			return code, nil
		}
	}
	return "", fmt.Errorf("generate game code: no unique code available")
}

func availableFaction(state game.Game) int {
	for i := 0; i < int(state.MaxPlayers); i++ {
		if !state.Factions[i].AI && state.Factions[i].Player == nil {
			return i
		}
	}
	return -1
}

func clientsExcept(clients []*Client, excluded *Client) []*Client {
	result := make([]*Client, 0, len(clients))
	for _, client := range clients {
		if client != excluded {
			result = append(result, client)
		}
	}
	return result
}

func playerPointer(player game.Player) *game.Player {
	return &player
}

func broadcast(clients []*Client, packet packets.S2CPacket) {
	for _, client := range clients {
		_ = client.SendPacket(packet)
	}
}
