package server

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
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
	ErrAlreadyInGame      = errors.New("client is already in a game")
	ErrAlreadyMatchmaking = errors.New("client is already searching for a game")
	ErrGameNotFound       = errors.New("game not found")
	ErrGameFull           = errors.New("game is full")
	ErrNotHost            = errors.New("only the host can start the game")
	ErrNotEnoughPlayers   = errors.New("need at least 2 players to start")

	Lobbies = NewLobbyManager()
)

// JoinOutcome is the result of JoinGame: either an immediate lobby join or
// placement into the random-matchmaking queue.
type JoinOutcome struct {
	Game          game.Game
	Waiting       bool
	QueuePosition int
}

// lobby is a created game waiting to start: players gather here until the
// host launches it, at which point it is handed over to GameInstances and
// becomes a running GameInstance.
type lobby struct {
	state   game.Game
	clients []*Client
}

// LobbyManager tracks lobbies, i.e. games that have not started yet.
// Running games are owned by GameInstances; the set of all server
// connections lives in Connections. Clients waiting for a public lobby
// sit in the matchmaking queue.
type LobbyManager struct {
	mu            sync.RWMutex
	nextGameID    uint64
	lobbies       map[uint64]*lobby
	gameCodes     map[string]uint64
	clientLobbies map[*Client]uint64
	matchmaking   []*Client
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

	m.mu.Lock()
	if err := m.ensureAvailableLocked(client); err != nil {
		m.mu.Unlock()
		return game.Game{}, err
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

	var matched []joinNotification
	if public {
		matched = m.fillMatchmakingLocked()
		state = m.lobbies[gameID].state
	}
	m.mu.Unlock()

	flushJoinNotifications(matched)
	return state, nil
}

func (m *LobbyManager) JoinGame(client *Client, code string) (JoinOutcome, error) {
	player, identified := client.Player()
	if !identified {
		return JoinOutcome{}, fmt.Errorf("client is not identified")
	}

	code = strings.ToUpper(strings.TrimSpace(code))
	randomJoin := code == ""

	m.mu.Lock()
	if err := m.ensureAvailableLocked(client); err != nil {
		m.mu.Unlock()
		return JoinOutcome{}, err
	}

	active := m.findJoinableLobbyLocked(code)
	if active == nil {
		if !randomJoin {
			m.mu.Unlock()
			return JoinOutcome{}, ErrGameNotFound
		}

		m.matchmaking = append(m.matchmaking, client)
		position := len(m.matchmaking)
		m.mu.Unlock()
		return JoinOutcome{Waiting: true, QueuePosition: position}, nil
	}

	factionIndex := availableFaction(active.state)
	if factionIndex < 0 {
		m.mu.Unlock()
		return JoinOutcome{}, ErrGameFull
	}

	active.state.Factions[factionIndex].Player = playerPointer(player)
	active.clients = append(active.clients, client)
	m.clientLobbies[client] = active.state.GameID
	state := active.state
	existingClients := clientsExcept(active.clients, client)
	m.mu.Unlock()

	broadcast(existingClients, &packets.S2CGameUpdatePacket{Game: state})
	return JoinOutcome{Game: state}, nil
}

// StartGame hands a lobby over to the game loop. Only the host may start,
// and at least two connected players are required. Clients are attached to
// the GameInstance before the lobby lock is released so a concurrent leave
// cannot miss the handoff.
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

	// Attach under the lobby lock so LeaveGame/CreateGame see GameInstance
	// before this handoff becomes visible as "not in a lobby".
	gi := GameInstances.Prepare(&state, clients)
	m.mu.Unlock()

	go gi.Run()
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
	m.removeFromMatchmakingLocked(client)

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
	matched := m.fillMatchmakingLocked()
	state := active.state
	lobbyClients := append([]*Client(nil), active.clients...)
	m.mu.Unlock()

	broadcast(lobbyClients, &packets.S2CGameUpdatePacket{Game: state})
	flushJoinNotifications(matched)
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

type joinNotification struct {
	client *Client
	state  game.Game
}

// fillMatchmakingLocked pulls queued clients into public lobbies with free
// seats. Caller must hold m.mu. Returned notifications must be flushed after
// unlock so sends do not run under the lobby lock. Each notification carries
// the lobby state after every successful dequeue in this pass.
func (m *LobbyManager) fillMatchmakingLocked() []joinNotification {
	if len(m.matchmaking) == 0 {
		return nil
	}

	joined := make([]*Client, 0, len(m.matchmaking))
	remaining := make([]*Client, 0, len(m.matchmaking))

	for i := 0; i < len(m.matchmaking); i++ {
		client := m.matchmaking[i]
		if client.GameInstance() != nil {
			continue
		}
		if _, exists := m.clientLobbies[client]; exists {
			continue
		}

		active := m.findJoinableLobbyLocked("")
		if active == nil {
			remaining = append(remaining, m.matchmaking[i:]...)
			break
		}

		player, identified := client.Player()
		if !identified {
			continue
		}
		factionIndex := availableFaction(active.state)
		if factionIndex < 0 {
			remaining = append(remaining, m.matchmaking[i:]...)
			break
		}

		active.state.Factions[factionIndex].Player = playerPointer(player)
		active.clients = append(active.clients, client)
		m.clientLobbies[client] = active.state.GameID
		joined = append(joined, client)
	}
	m.matchmaking = remaining

	matched := make([]joinNotification, 0, len(joined))
	for _, client := range joined {
		gameID := m.clientLobbies[client]
		matched = append(matched, joinNotification{
			client: client,
			state:  m.lobbies[gameID].state,
		})
	}
	return matched
}

func flushJoinNotifications(matched []joinNotification) {
	for _, note := range matched {
		_ = note.client.SendPacket(&packets.S2CGameJoinedPacket{Game: note.state})
	}
}

func (m *LobbyManager) ensureAvailableLocked(client *Client) error {
	if client.GameInstance() != nil {
		return ErrAlreadyInGame
	}
	if _, exists := m.clientLobbies[client]; exists {
		return ErrAlreadyInGame
	}
	if m.isMatchmakingLocked(client) {
		return ErrAlreadyMatchmaking
	}
	return nil
}

func (m *LobbyManager) isMatchmakingLocked(client *Client) bool {
	for _, queued := range m.matchmaking {
		if queued == client {
			return true
		}
	}
	return false
}

func (m *LobbyManager) removeFromMatchmakingLocked(client *Client) {
	for i, queued := range m.matchmaking {
		if queued != client {
			continue
		}
		copy(m.matchmaking[i:], m.matchmaking[i+1:])
		m.matchmaking[len(m.matchmaking)-1] = nil
		m.matchmaking = m.matchmaking[:len(m.matchmaking)-1]
		return
	}
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
		if err := client.SendPacket(packet); err != nil {
			log.Printf("lobby broadcast failed for %T: %v", packet, err)
		}
	}
}
