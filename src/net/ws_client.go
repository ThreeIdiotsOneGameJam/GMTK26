package net

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/threeidiotsonegamejam/gmtk26/src/constants"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/settings"
)

const (
	initialReconnectDelay = time.Second
	maxReconnectDelay     = 30 * time.Second
	offlinePollPeriod     = 250 * time.Millisecond
)

//go:generate stringer -type=ConnectionState -trimprefix=Connection

type ConnectionState int32

const (
	ConnectionDisconnected ConnectionState = iota
	ConnectionConnecting
	ConnectionConnected
)

var client = newWSClient()

func newWSClient() *WSClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &WSClient{
		ctx:     ctx,
		cancel:  cancel,
		runDone: make(chan struct{}),
		retry:   make(chan struct{}, 1),
	}
}

func Connect(addr string) {
	client.run(addr)
}

func (c *WSClient) run(addr string) {
	c.connectOnce.Do(func() {
		if !c.beginRun() {
			return
		}
		defer c.endRun()

		c.connect(addr)
	})
}

func Send(packet packets.C2SPacket) error {
	if local := activeLocalTransport(); local != nil {
		return local.send(packet)
	}
	return client.send(packet)
}

func DrainEvents(onPacket func(packets.S2CPacket)) {
	if local := activeLocalTransport(); local != nil {
		local.drainEvents(onPacket)
		return
	}
	client.drainEvents(onPacket)
}

func State() ConnectionState {
	return ConnectionState(client.stateAtomic.Load())
}

func RetryConnection() {
	client.retryConnection()
}

func suspendRemote() {
	client.suspend()
}

func resumeRemote() {
	client.resume()
}

func Close() {
	stopLocalGame(false)
	client.close()
}

type WSClient struct {
	conn   *Connection
	connMu sync.Mutex

	stateAtomic atomic.Int32
	suspended   atomic.Bool

	events   []packets.S2CPacket
	eventsMu sync.Mutex
	retry    chan struct{}

	ctx         context.Context
	cancel      context.CancelFunc
	connectOnce sync.Once
	closeOnce   sync.Once

	lifecycleMu sync.Mutex
	running     bool
	runDone     chan struct{}
}

func (c *WSClient) beginRun() bool {
	c.lifecycleMu.Lock()
	defer c.lifecycleMu.Unlock()

	if c.ctx.Err() != nil {
		return false
	}

	c.running = true
	return true
}

func (c *WSClient) endRun() {
	c.lifecycleMu.Lock()
	c.running = false
	close(c.runDone)
	c.lifecycleMu.Unlock()
}

func (c *WSClient) connect(addr string) {
	retryDelay := initialReconnectDelay

	for {
		if c.ctx.Err() != nil {
			c.setState(ConnectionDisconnected)
			return
		}
		if !c.waitUntilAvailable() {
			return
		}

		c.clearRetryRequest()
		c.setState(ConnectionConnecting)

		// coder/websocket dials through the browser WebSocket API on
		// js/wasm; canceling c.ctx aborts an in-flight dial.
		url := fmt.Sprintf("%s://%s", constants.ClientWebSocketScheme, addr)
		conn, _, err := websocket.Dial(c.ctx, url, nil)
		if err != nil {
			c.setState(ConnectionDisconnected)
			if !c.waitForRetry(retryDelay) {
				return
			}
			retryDelay = nextReconnectDelay(retryDelay)
			continue
		}

		connection := NewConnection(conn)
		connection.Start(func(packet packets.Packet) error {
			return c.handlePacket(connection, packet)
		})
		if !c.setConnection(connection) {
			connection.Close()
			connection.wait()
			c.setState(ConnectionDisconnected)
			return
		}
		if err := connection.SendPacket(&packets.C2SConnectPacket{
			Player: *game.PlayerData,
		}); err != nil {
			c.clearConnection(connection)
			connection.Close()
			connection.wait()
			c.setState(ConnectionDisconnected)
			if !c.waitForRetry(retryDelay) {
				return
			}
			retryDelay = nextReconnectDelay(retryDelay)
			continue
		}

		select {
		case <-connection.Done():
		case <-c.ctx.Done():
			connection.Close()
		}
		connection.wait()

		wasConnected := c.state() == ConnectionConnected
		if closeErr := connection.CloseError(); closeErr != nil {
			if isMessageTooBig(closeErr) {
				fmt.Printf("disconnected: oversized websocket message: %v\n", closeErr)
			} else if wasConnected {
				fmt.Printf("disconnected: %v\n", closeErr)
			}
		}
		c.clearConnection(connection)
		c.setState(ConnectionDisconnected)
		if wasConnected {
			retryDelay = initialReconnectDelay
		}

		if !c.waitForRetry(retryDelay) {
			return
		}
		retryDelay = nextReconnectDelay(retryDelay)
	}
}

func (c *WSClient) waitForRetry(delay time.Duration) bool {
	if c.unavailable() {
		return c.waitUntilAvailable()
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()
	offlineTicker := time.NewTicker(offlinePollPeriod)
	defer offlineTicker.Stop()

	for {
		select {
		case <-timer.C:
			if c.unavailable() {
				return c.waitUntilAvailable()
			}
			return true
		case <-c.retry:
			if c.unavailable() {
				continue
			}
			return true
		case <-offlineTicker.C:
			if c.unavailable() {
				return c.waitUntilAvailable()
			}
		case <-c.ctx.Done():
			return false
		}
	}
}

func (c *WSClient) waitUntilAvailable() bool {
	if !c.unavailable() {
		return true
	}

	c.setState(ConnectionDisconnected)
	ticker := time.NewTicker(offlinePollPeriod)
	defer ticker.Stop()

	for c.unavailable() {
		select {
		case <-ticker.C:
		case <-c.retry:
		case <-c.ctx.Done():
			return false
		}
	}
	return true
}

func (c *WSClient) setConnection(conn *Connection) bool {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.ctx.Err() != nil || c.suspended.Load() {
		return false
	}

	c.conn = conn
	return true
}

func (c *WSClient) clearConnection(conn *Connection) {
	c.connMu.Lock()
	if c.conn == conn {
		c.conn = nil
	}
	c.connMu.Unlock()
}

func (c *WSClient) close() {
	c.closeOnce.Do(func() {
		c.lifecycleMu.Lock()
		c.cancel()
		waitForRun := c.running
		c.lifecycleMu.Unlock()

		c.connMu.Lock()
		conn := c.conn
		c.conn = nil
		c.connMu.Unlock()

		if conn != nil {
			conn.Close()
		}

		if waitForRun {
			<-c.runDone
		}

		c.eventsMu.Lock()
		c.events = nil
		c.eventsMu.Unlock()
		c.setState(ConnectionDisconnected)
	})
}

func (c *WSClient) send(packet packets.C2SPacket) error {
	c.connMu.Lock()
	conn := c.conn
	c.connMu.Unlock()

	if conn == nil {
		return ErrConnectionClosed
	}

	return conn.SendPacket(packet)
}

func (c *WSClient) handlePacket(connection *Connection, packet packets.Packet) error {
	c.connMu.Lock()
	current := c.conn
	c.connMu.Unlock()
	if current != connection || c.suspended.Load() {
		return nil
	}

	serverPacket, ok := packet.(packets.S2CPacket)
	if !ok {
		return fmt.Errorf("received client packet %T from server", packet)
	}
	if _, accepted := serverPacket.(*packets.S2CConnectAcceptedPacket); accepted {
		c.setState(ConnectionConnected)
	}

	c.eventsMu.Lock()
	c.events = append(c.events, serverPacket)
	c.eventsMu.Unlock()

	return nil
}

func (c *WSClient) drainEvents(onPacket func(packets.S2CPacket)) {
	c.eventsMu.Lock()
	events := c.events
	c.events = nil
	c.eventsMu.Unlock()

	for _, p := range events {
		onPacket(p)
	}
}

func (c *WSClient) setState(s ConnectionState) {
	c.stateAtomic.Store(int32(s))
}

func (c *WSClient) state() ConnectionState {
	return ConnectionState(c.stateAtomic.Load())
}

func (c *WSClient) retryConnection() {
	if c.unavailable() || c.state() != ConnectionDisconnected {
		return
	}

	c.signalRetry()
}

func (c *WSClient) signalRetry() {
	select {
	case c.retry <- struct{}{}:
	default:
	}
}

func (c *WSClient) clearRetryRequest() {
	select {
	case <-c.retry:
	default:
	}
}

func nextReconnectDelay(delay time.Duration) time.Duration {
	return min(delay*2, maxReconnectDelay)
}

func (c *WSClient) unavailable() bool {
	return settings.Current.Offline || c.suspended.Load()
}

func (c *WSClient) suspend() {
	if c.suspended.Swap(true) {
		return
	}

	c.clearEvents()

	c.connMu.Lock()
	conn := c.conn
	c.conn = nil
	c.connMu.Unlock()
	if conn != nil {
		// Solo startup runs on the UI thread, so tear the socket down
		// immediately instead of waiting for a graceful close handshake.
		conn.closeOnError()
	}

	c.signalRetry()
}

func (c *WSClient) resume() {
	if !c.suspended.Swap(false) {
		return
	}

	// Nothing received before or during suspension belongs to the new remote
	// session that will be established after solo play.
	c.clearEvents()
	c.signalRetry()
}

func (c *WSClient) clearEvents() {
	c.eventsMu.Lock()
	c.events = nil
	c.eventsMu.Unlock()
}
