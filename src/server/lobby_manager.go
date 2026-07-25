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
	ErrAlreadyInGame    = errors.New("client is already in a game")
	ErrGameNotFound     = errors.New("game not found")
	ErrGameFull         = errors.New("game is full")
	ErrNotHost          = errors.New("only the host can start the game")
	ErrNotEnoughPlayers = errors.New("need at least 2 players to start")

	Lobbies = NewLobbyManager()
)

// lobby is a created game waiting to start: players gather here until the
// host launches it, at which point it is handed over to GameInstances and
// becomes a running GameInstance.
type lobby struct {
	state   game.Game
	clients []*Client
}

// LobbyManager tracks lobbies, i.e. games that have not started yet.
// Running games are owned by GameInstances; the set of all server
// connections lives in Connections.
type LobbyManager struct {
	mu            sync.RWMutex
	nextGameID    uint64
	lobbies       map[uint64]*lobby
	gameCodes     map[string]uint64
	clientLobbies map[*Client]uint64
}

func NewLobbyManager() *LobbyManager {
	return &LobbyManager{
		lobbies:       make(map[uint64]*lobby),
		gameCodes:     make(map[string]uint64),
		clientLobbies: make(map[*Client]uint64),
	}
}

func (m *LobbyManager) CreateGame(client *Client, public bool, maxPlayers uint8, seed int64) (game.Game, error) {
	if maxPlayers < 2 || maxPlayers > 4 {
		return game.Game{}, fmt.Errorf("max players must be 2, 3, or 4")
	}

	player, identified := client.Player()
	if !identified {
		return game.Game{}, fmt.Errorf("client is not identified")
	}

	if client.GameInstance() != nil {
		return game.Game{}, ErrAlreadyInGame
	}

	m.mu.Lock()
	if _, exists := m.clientLobbies[client]; exists {
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

	m.lobbies[gameID] = &lobby{
		state:   state,
		clients: []*Client{client},
	}
	m.gameCodes[gameCode] = gameID
	m.clientLobbies[client] = gameID
	m.mu.Unlock()

	return state, nil
}

func (m *LobbyManager) JoinGame(client *Client, code string) (game.Game, error) {
	player, identified := client.Player()
	if !identified {
		return game.Game{}, fmt.Errorf("client is not identified")
	}

	if client.GameInstance() != nil {
		return game.Game{}, ErrAlreadyInGame
	}

	code = strings.ToUpper(strings.TrimSpace(code))

	m.mu.Lock()
	if _, exists := m.clientLobbies[client]; exists {
		m.mu.Unlock()
		return game.Game{}, ErrAlreadyInGame
	}

	active := m.findJoinableLobbyLocked(code)
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
	m.clientLobbies[client] = active.state.GameID
	state := active.state
	existingClients := clientsExcept(active.clients, client)
	m.mu.Unlock()

	broadcast(existingClients, &packets.S2CGameUpdatePacket{Game: state})
	return state, nil
}

// StartGame hands a lobby over to the game loop. Only the host may start,
// and at least two connected players are required. On success the lobby is
// dissolved (it can no longer be joined or host-closed); the running
// GameInstance takes over and notifies every player via S2CGameStartPacket.
// Connection tracking (server.Connections) is unaffected.
func (m *LobbyManager) StartGame(client *Client) error {
	m.mu.Lock()
	gameID, exists := m.clientLobbies[client]
	if !exists {
		m.mu.Unlock()
		return ErrGameNotFound
	}
	active := m.lobbies[gameID]
	if active == nil {
		m.mu.Unlock()
		return ErrGameNotFound
	}

	player, _ := client.Player()
	if active.state.HostID != player.ClientID {
		m.mu.Unlock()
		return ErrNotHost
	}

	clients := clientsByFaction(active.state, active.clients)
	playerCount := 0
	for _, c := range clients {
		if c != nil {
			playerCount++
		}
	}
	if playerCount < 2 {
		m.mu.Unlock()
		return ErrNotEnoughPlayers
	}

	delete(m.lobbies, gameID)
	delete(m.gameCodes, active.state.GameCode)
	for _, c := range active.clients {
		delete(m.clientLobbies, c)
	}
	state := active.state
	m.mu.Unlock()

	GameInstances.StartGame(&state, clients)
	return nil
}

// clientsByFaction returns a slice with one slot per faction, holding the
// client controlling that faction or nil for AI and unfilled slots.
func clientsByFaction(state game.Game, clients []*Client) []*Client {
	ordered := make([]*Client, len(state.Factions))
	for i := range state.Factions {
		factionPlayer := state.Factions[i].Player
		if factionPlayer == nil {
			continue
		}
		for _, client := range clients {
			player, _ := client.Player()
			if player.ClientID == factionPlayer.ClientID {
				ordered[i] = client
				break
			}
		}
	}
	return ordered
}

func (m *LobbyManager) LeaveGame(client *Client) {
	m.removeClient(client, "host left the game")
}

func (m *LobbyManager) Disconnect(client *Client) {
	m.removeClient(client, "host disconnected")
}

func (m *LobbyManager) OpenLobbies() []game.Game {
	m.mu.RLock()
	defer m.mu.RUnlock()

	games := make([]game.Game, 0, len(m.lobbies))
	for _, active := range m.lobbies {
		games = append(games, active.state)
	}
	return games
}

func (m *LobbyManager) removeClient(client *Client, hostReason string) {
	m.mu.Lock()
	gameID, exists := m.clientLobbies[client]
	if !exists {
		m.mu.Unlock()
		return
	}

	active := m.lobbies[gameID]
	delete(m.clientLobbies, client)
	if active == nil {
		m.mu.Unlock()
		return
	}

	player, _ := client.Player()
	if active.state.HostID == player.ClientID {
		delete(m.lobbies, gameID)
		delete(m.gameCodes, active.state.GameCode)
		remainingClients := clientsExcept(active.clients, client)
		for _, remaining := range remainingClients {
			delete(m.clientLobbies, remaining)
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

func (m *LobbyManager) findJoinableLobbyLocked(code string) *lobby {
	if code != "" {
		gameID, exists := m.gameCodes[code]
		if !exists {
			return nil
		}
		return m.lobbies[gameID]
	}

	var selected *lobby
	for _, active := range m.lobbies {
		if !active.state.Public || availableFaction(active.state) < 0 {
			continue
		}
		if selected == nil || active.state.GameID < selected.state.GameID {
			selected = active
		}
	}
	return selected
}

func (m *LobbyManager) newGameCodeLocked() (string, error) {
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
