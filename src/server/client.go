package server

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
)

type fatalPacketError struct {
	err error
}

func (e *fatalPacketError) Error() string {
	return e.err.Error()
}

func (e *fatalPacketError) Unwrap() error {
	return e.err
}

func fatalPacketErrorf(format string, args ...any) error {
	return &fatalPacketError{err: fmt.Errorf(format, args...)}
}

func IsFatalPacketError(err error) bool {
	var fatalErr *fatalPacketError
	return errors.As(err, &fatalErr)
}

var temporaryClientCounter atomic.Uint64

type Client struct {
	mu         sync.RWMutex
	player     game.Player
	persistent bool
	ready      bool
	sender     PacketSender
	game       *GameInstance
}

func (c *Client) JoinGame(gi *GameInstance) {
	c.mu.Lock()
	c.game = gi
	c.mu.Unlock()
}

func (c *Client) LeaveGame() {
	c.mu.Lock()
	gi := c.game
	c.game = nil
	c.mu.Unlock()

	if gi != nil {
		gi.clientLeft()
	}
}

// LeaveGameInstance detaches the client only if it is still attached to the
// given instance, so a finished game cannot clobber a newer one.
func (c *Client) LeaveGameInstance(gi *GameInstance) {
	c.mu.Lock()
	if c.game == gi {
		c.game = nil
	}
	c.mu.Unlock()
}

func (c *Client) GameInstance() *GameInstance {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.game
}

func NewClient(sender PacketSender) *Client {
	return &Client{sender: sender}
}

func (c *Client) SendPacket(packet packets.S2CPacket) error {
	return c.sender.SendPacket(packet)
}

func (c *Client) Player() (game.Player, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.player, c.player.ClientID != ""
}

// HasPersistentIdentity reports whether the client supplied a stable UUID that
// can be used by long-term storage such as leaderboards.
func (c *Client) HasPersistentIdentity() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.persistent
}

func (c *Client) MarkReady() {
	c.mu.Lock()
	c.ready = true
	c.mu.Unlock()
}

func (c *Client) Ready() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.ready
}

func (c *Client) HandlePacket(packet packets.C2SPacket) (packets.S2CPacket, error) {
	switch packet := packet.(type) {
	case *packets.C2SConnectPacket:
		return c.handleConnectPacket(packet)
	case *packets.C2SCreateGamePacket:
		return c.handleCreateGamePacket(packet)
	case *packets.C2SJoinGamePacket:
		return c.handleJoinGamePacket(packet)
	case *packets.C2SLeaveGamePacket:
		Lobbies.LeaveGame(c)
		c.LeaveGame()
		return nil, nil
	case *packets.C2SStartGamePacket:
		return c.handleStartGamePacket()
	case *packets.C2SActionPacket:
		gi := c.GameInstance()
		if gi == nil {
			return nil, fatalPacketErrorf("handle action packet: not in a game")
		}
		if err := gi.SubmitAction(c, packet.Round, packet.Type, packet.Build, packet.Dispatch); err != nil {
			// Late or invalid actions are dropped; the next state broadcast
			// resynchronizes the client.
			return nil, fmt.Errorf("handle action packet: %w", err)
		}
		return nil, nil
	default:
		return nil, fatalPacketErrorf(
			"handle client packet: unsupported packet %T",
			packet,
		)
	}
}

func (c *Client) handleCreateGamePacket(packet *packets.C2SCreateGamePacket) (packets.S2CPacket, error) {
	if !c.Ready() {
		return nil, fatalPacketErrorf("handle create game packet: client is not ready")
	}

	state, err := Lobbies.CreateGame(c, packet.Public, packet.MaxPlayers, packet.Seed)
	if err != nil {
		return &packets.S2CGameRejectedPacket{
			Operation: "create",
			Message:   err.Error(),
		}, nil
	}
	return &packets.S2CGameJoinedPacket{Game: state}, nil
}

func (c *Client) handleJoinGamePacket(packet *packets.C2SJoinGamePacket) (packets.S2CPacket, error) {
	if !c.Ready() {
		return nil, fatalPacketErrorf("handle join game packet: client is not ready")
	}

	outcome, err := Lobbies.JoinGame(c, packet.GameCode)
	if err != nil {
		return &packets.S2CGameRejectedPacket{
			Operation: "join",
			Message:   err.Error(),
		}, nil
	}
	if outcome.Waiting {
		return &packets.S2CMatchmakingWaitingPacket{
			QueuePosition: outcome.QueuePosition,
		}, nil
	}
	return &packets.S2CGameJoinedPacket{Game: outcome.Game}, nil
}

func (c *Client) handleStartGamePacket() (packets.S2CPacket, error) {
	if !c.Ready() {
		return nil, fatalPacketErrorf("handle start game packet: client is not ready")
	}

	if err := Lobbies.StartGame(c); err != nil {
		return &packets.S2CGameRejectedPacket{
			Operation: "start",
			Message:   err.Error(),
		}, nil
	}
	return nil, nil
}

func (c *Client) handleConnectPacket(packet *packets.C2SConnectPacket) (packets.S2CPacket, error) {
	rawClientID := string(packet.Player.ClientID)
	clientID, err := uuid.Parse(rawClientID)
	if err != nil {
		return nil, fatalPacketErrorf(
			"handle connect packet: invalid client ID: %v",
			err,
		)
	}
	if clientID.String() != rawClientID {
		return nil, fatalPacketErrorf(
			"handle connect packet: client ID must use canonical UUID format",
		)
	}
	persistent := clientID != uuid.Nil
	if !persistent {
		clientID = newTemporaryClientID()
	}

	c.mu.Lock()
	if c.player.ClientID != "" {
		c.mu.Unlock()
		return nil, fatalPacketErrorf(
			"handle connect packet: client is already identified",
		)
	}
	c.player = packet.Player
	c.player.ClientID = game.ClientID(clientID.String())
	c.persistent = persistent
	c.mu.Unlock()

	return &packets.S2CConnectAcceptedPacket{
		ClientID:   game.ClientID(clientID.String()),
		Persistent: persistent,
	}, nil
}

func newTemporaryClientID() uuid.UUID {
	clientID, err := uuid.NewRandom()
	if err == nil {
		return clientID
	}

	seed := fmt.Sprintf(
		"%d:%d",
		time.Now().UnixNano(),
		temporaryClientCounter.Add(1),
	)
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(seed))
}
