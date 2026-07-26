package net

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/server"
)

var ErrLocalGameActive = errors.New("a local game is already active")

var localGame = struct {
	sync.RWMutex
	transport *localTransport
}{}

// StartLocalGame starts a private one-player server in this process. The
// multiplayer socket is suspended until StopLocalGame is called.
func StartLocalGame(seed int64) error {
	if game.PlayerData == nil {
		return fmt.Errorf("start local game: player data is not loaded")
	}

	localGame.Lock()
	if localGame.transport != nil {
		localGame.Unlock()
		return ErrLocalGameActive
	}

	suspendRemote()
	transport := newLocalTransport()
	localGame.transport = transport
	localGame.Unlock()

	transport.start()
	if err := transport.send(&packets.C2SConnectPacket{
		Player: *game.PlayerData,
	}); err != nil {
		stopLocalGame(true)
		return fmt.Errorf("start local game handshake: %w", err)
	}
	if err := transport.send(&packets.C2SCreateGamePacket{
		Public:     false,
		MaxPlayers: 1,
		Seed:       seed,
	}); err != nil {
		stopLocalGame(true)
		return fmt.Errorf("create local game: %w", err)
	}

	return nil
}

func StopLocalGame() {
	stopLocalGame(true)
}

func stopLocalGame(resume bool) {
	stopLocalTransport(nil, resume)
}

// stopLocalTransport stops the current local transport. If expected is
// non-nil, cleanup only proceeds while it is still the active transport.
// This prevents delayed terminal cleanup from affecting a newer local game.
func stopLocalTransport(expected *localTransport, resume bool) {
	localGame.Lock()
	transport := localGame.transport
	if expected != nil && transport != expected {
		localGame.Unlock()
		return
	}
	localGame.transport = nil
	localGame.Unlock()

	if transport != nil {
		transport.close()
	}
	if resume {
		resumeRemote()
	}
}

func LocalGameActive() bool {
	localGame.RLock()
	defer localGame.RUnlock()
	return localGame.transport != nil
}

func activeLocalTransport() *localTransport {
	localGame.RLock()
	defer localGame.RUnlock()
	return localGame.transport
}

type localTransport struct {
	session  *server.Session
	inbound  chan []byte
	outbound chan []byte
	stop     chan struct{}
	runDone  chan struct{}

	closeOnce sync.Once
	joinedMu  sync.RWMutex
	joined    bool
	terminal  atomic.Bool
}

func newLocalTransport() *localTransport {
	transport := &localTransport{
		inbound:  make(chan []byte, sendQueueLength),
		outbound: make(chan []byte, sendQueueLength),
		stop:     make(chan struct{}),
		runDone:  make(chan struct{}),
	}
	transport.session = server.NewSession(transport)
	return transport
}

func (t *localTransport) start() {
	go t.run()
}

func (t *localTransport) run() {
	defer close(t.runDone)

	for {
		select {
		case message := <-t.inbound:
			packet, err := packets.Deserialize(message)
			if err != nil {
				t.reportFailure(fmt.Errorf("decode local client packet: %w", err))
				t.terminal.Store(true)
				return
			}
			if err := t.session.HandlePacket(packet); err != nil {
				t.reportFailure(err)
				t.terminal.Store(true)
				return
			}
		case <-t.stop:
			return
		}
	}
}

func (t *localTransport) reportFailure(err error) {
	operation := "create"
	t.joinedMu.RLock()
	if t.joined {
		operation = "start"
	}
	t.joinedMu.RUnlock()

	rejection := &packets.S2CGameRejectedPacket{
		Operation: operation,
		Message:   "Local server failed: " + err.Error(),
	}
	if sendErr := t.enqueueOutbound(rejection); sendErr != nil {
		fmt.Printf("local server failed: %v (could not report: %v)\n", err, sendErr)
	}
}

// SendPacket implements server.PacketSender.
func (t *localTransport) SendPacket(packet packets.Packet) error {
	message, err := serializeLocalPacket(packet)
	if err != nil {
		return err
	}

	// The local handshake is transport-internal. In particular, a temporary
	// guest ID must not replace the ID assigned by the remote server.
	if _, accepted := packet.(*packets.S2CConnectAcceptedPacket); accepted {
		decoded, err := packets.Deserialize(message)
		if err != nil {
			return err
		}
		if _, ok := decoded.(*packets.S2CConnectAcceptedPacket); !ok {
			return fmt.Errorf("local handshake decoded as %T", decoded)
		}
		return nil
	}

	if err := t.enqueueOutboundMessage(message); err != nil {
		return err
	}

	if _, joined := packet.(*packets.S2CGameJoinedPacket); joined {
		t.joinedMu.Lock()
		firstJoin := !t.joined
		t.joined = true
		t.joinedMu.Unlock()
		if firstJoin {
			// Queue the automatic start only after the joined response so the
			// UI always establishes currentGame before receiving game state.
			return t.send(&packets.C2SStartGamePacket{})
		}
	}

	return nil
}

func (t *localTransport) send(packet packets.C2SPacket) error {
	message, err := serializeLocalPacket(packet)
	if err != nil {
		return err
	}

	select {
	case <-t.stop:
		return ErrConnectionClosed
	case <-t.runDone:
		return ErrConnectionClosed
	default:
	}

	select {
	case t.inbound <- message:
		return nil
	case <-t.stop:
		return ErrConnectionClosed
	case <-t.runDone:
		return ErrConnectionClosed
	default:
		return ErrSendQueueFull
	}
}

func (t *localTransport) enqueueOutbound(packet packets.S2CPacket) error {
	message, err := serializeLocalPacket(packet)
	if err != nil {
		return err
	}
	return t.enqueueOutboundMessage(message)
}

func (t *localTransport) enqueueOutboundMessage(message []byte) error {
	select {
	case <-t.stop:
		return ErrConnectionClosed
	default:
	}

	select {
	case t.outbound <- message:
		return nil
	case <-t.stop:
		return ErrConnectionClosed
	default:
		return ErrSendQueueFull
	}
}

func serializeLocalPacket(packet packets.Packet) ([]byte, error) {
	message, err := packets.Serialize(packet)
	if err != nil {
		return nil, err
	}
	if err := outboundMessageSizeError(packet, len(message)); err != nil {
		return nil, err
	}
	return message, nil
}

func (t *localTransport) drainEvents(onPacket func(packets.S2CPacket)) {
	var events []packets.S2CPacket
	for {
		select {
		case message := <-t.outbound:
			packet, err := packets.Deserialize(message)
			if err != nil {
				fmt.Printf("failed to decode local server packet: %v\n", err)
				continue
			}
			serverPacket, ok := packet.(packets.S2CPacket)
			if !ok {
				fmt.Printf("local server sent client packet %T\n", packet)
				continue
			}
			events = append(events, serverPacket)
		default:
			for _, packet := range events {
				onPacket(packet)
			}
			if t.terminal.Load() {
				stopLocalTransport(t, true)
			}
			return
		}
	}
}

func (t *localTransport) close() {
	t.closeOnce.Do(func() {
		close(t.stop)
		<-t.runDone
		t.session.Close()
	})
}
