package net

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
)

const (
	writeWait      = 10 * time.Second // 10s
	pingPeriod     = 54 * time.Second // 54s (0.9min)
	// Maps are ~200KB as JSON (96x96 grid); 64KB rejected start/state
	// packets and dropped the socket, which bounced clients to the menu.
	maxMessageSize  = 1 << 20 // 1 MiB
	sendQueueLength = 64
)

var (
	ErrConnectionClosed = errors.New("websocket connection is closed")
	ErrSendQueueFull    = errors.New("websocket send queue is full")
)

type Connection struct {
	conn     *websocket.Conn
	onPacket func(packets.Packet) error
	send     chan []byte
	done     chan struct{}

	// pings enables periodic liveness pings. The server always enables
	// them. Native clients do too so a black-holed peer is detected;
	// js/wasm cannot send WebSocket pings, so those clients rely on the
	// server (and the browser close event) instead.
	pings bool

	startOnce sync.Once
	closeOnce sync.Once
	pumpWG    sync.WaitGroup
	stateMu   sync.RWMutex
	closed    bool
}

func NewConnection(conn *websocket.Conn) *Connection {
	return &Connection{
		conn:  conn,
		send:  make(chan []byte, sendQueueLength),
		done:  make(chan struct{}),
		pings: runtime.GOOS != "js",
	}
}

func NewServerConnection(conn *websocket.Conn) *Connection {
	c := NewConnection(conn)
	c.pings = true
	return c
}

func (c *Connection) Start(onPacket func(packets.Packet) error) {
	c.startOnce.Do(func() {
		c.onPacket = onPacket
		c.conn.SetReadLimit(maxMessageSize)

		c.pumpWG.Add(2)
		go func() {
			defer c.pumpWG.Done()
			c.writePump()
		}()
		go func() {
			defer c.pumpWG.Done()
			c.readPump()
		}()

		if c.pings {
			c.pumpWG.Add(1)
			go func() {
				defer c.pumpWG.Done()
				c.pingPump()
			}()
		}
	})
}

func (c *Connection) SendPacket(packet packets.Packet) error {
	message, err := packets.Serialize(packet)
	if err != nil {
		return err
	}

	c.stateMu.RLock()
	defer c.stateMu.RUnlock()

	if c.closed {
		return ErrConnectionClosed
	}

	select {
	case c.send <- message:
		return nil
	default:
		return ErrSendQueueFull
	}
}

func (c *Connection) Done() <-chan struct{} {
	return c.done
}

func (c *Connection) wait() {
	c.pumpWG.Wait()
}

func (c *Connection) Close() {
	c.closeWithStatus(websocket.StatusNormalClosure, "")
}

func (c *Connection) closeWithStatus(code websocket.StatusCode, reason string) {
	c.closeOnce.Do(func() {
		c.markClosed()
		_ = c.conn.Close(code, reason)
	})
}

func (c *Connection) closeOnError() {
	c.closeOnce.Do(func() {
		c.markClosed()
		_ = c.conn.CloseNow()
	})
}

func (c *Connection) markClosed() {
	c.stateMu.Lock()
	c.closed = true
	close(c.done)
	c.stateMu.Unlock()
}

func (c *Connection) readPump() {
	for {
		_, message, err := c.conn.Read(context.Background())
		if err != nil {
			c.closeOnError()
			return
		}

		packet, err := packets.Deserialize(message)
		if err != nil {
			c.closeWithStatus(websocket.StatusUnsupportedData, "invalid packet")
			return
		}

		if c.onPacket != nil {
			if err := c.onPacket(packet); err != nil {
				c.closeWithStatus(websocket.StatusPolicyViolation, "packet rejected")
				return
			}
		}
	}
}

func (c *Connection) writePump() {
	for {
		select {
		case message := <-c.send:
			ctx, cancel := context.WithTimeout(context.Background(), writeWait)
			err := c.conn.Write(ctx, websocket.MessageText, message)
			cancel()
			if err != nil {
				c.closeOnError()
				return
			}
		case <-c.done:
			return
		}
	}
}

// pingPump detects dead peers: Ping sends a ping frame and waits for the
// pong, so a peer that went away gets the connection closed within
// pingPeriod+writeWait.
func (c *Connection) pingPump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), writeWait)
			err := c.conn.Ping(ctx)
			cancel()
			if err != nil {
				c.closeOnError()
				return
			}
		case <-c.done:
			return
		}
	}
}
