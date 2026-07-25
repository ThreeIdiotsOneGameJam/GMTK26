package net

import (
	"context"
	"fmt"
	stdnet "net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
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
	return client.send(packet)
}

func DrainEvents(onPacket func(packets.S2CPacket)) {
	client.drainEvents(onPacket)
}

func State() ConnectionState {
	return ConnectionState(client.stateAtomic.Load())
}

func Close() {
	client.close()
}

type WSClient struct {
	conn   *Connection
	connMu sync.Mutex
	dialMu sync.Mutex
	dial   stdnet.Conn

	stateAtomic atomic.Int32

	events   []packets.S2CPacket
	eventsMu sync.Mutex

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
	for {
		if c.ctx.Err() != nil {
			c.setState(ConnectionDisconnected)
			return
		}

		c.setState(ConnectionConnecting)

		dialer := *websocket.DefaultDialer
		dialer.NetDialContext = c.dialContext
		conn, _, err := dialer.DialContext(c.ctx, "ws://"+addr, nil)
		c.dialMu.Lock()
		c.dial = nil
		c.dialMu.Unlock()
		if err != nil {
			c.setState(ConnectionDisconnected)
			if !c.waitForRetry() {
				return
			}
			continue
		}

		connection := NewConnection(conn)
		connection.Start(c.handlePacket)
		if err := connection.SendPacket(&packets.C2SConnectPacket{
			Player: *game.PlayerData,
		}); err != nil {
			connection.Close()
			connection.wait()
			c.setState(ConnectionDisconnected)
			if !c.waitForRetry() {
				return
			}
			continue
		}

		if !c.setConnection(connection) {
			connection.Close()
			connection.wait()
			c.setState(ConnectionDisconnected)
			return
		}

		select {
		case <-connection.Done():
		case <-c.ctx.Done():
			connection.Close()
		}
		connection.wait()

		c.connMu.Lock()
		c.conn = nil
		c.connMu.Unlock()
		c.setState(ConnectionDisconnected)

		if !c.waitForRetry() {
			return
		}
	}
}

func (c *WSClient) dialContext(
	ctx context.Context,
	network string,
	address string,
) (stdnet.Conn, error) {
	var dialer stdnet.Dialer
	conn, err := dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}

	c.dialMu.Lock()
	if c.ctx.Err() != nil {
		c.dialMu.Unlock()
		_ = conn.Close()
		return nil, c.ctx.Err()
	}
	c.dial = conn
	c.dialMu.Unlock()

	return conn, nil
}

func (c *WSClient) waitForRetry() bool {
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
		return true
	case <-c.ctx.Done():
		return false
	}
}

func (c *WSClient) setConnection(conn *Connection) bool {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.ctx.Err() != nil {
		return false
	}

	c.conn = conn
	c.setState(ConnectionConnected)
	return true
}

func (c *WSClient) close() {
	c.closeOnce.Do(func() {
		c.lifecycleMu.Lock()
		c.cancel()
		waitForRun := c.running
		c.lifecycleMu.Unlock()

		c.dialMu.Lock()
		dial := c.dial
		c.dial = nil
		c.dialMu.Unlock()
		if dial != nil {
			_ = dial.Close()
		}

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

func (c *WSClient) handlePacket(packet packets.Packet) error {
	serverPacket, ok := packet.(packets.S2CPacket)
	if !ok {
		return fmt.Errorf("received client packet %T from server", packet)
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
