package net

import (
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
)

const (
	writeWait       = 10 * time.Second    // 10s
	pongWait        = 60 * time.Second    // 60s
	pingPeriod      = (pongWait * 9) / 10 // 54s (0.9min)
	maxMessageSize  = 64 * 1024
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

	startOnce sync.Once
	closeOnce sync.Once
	pumpWG    sync.WaitGroup
	stateMu   sync.RWMutex
	closed    bool
}

func NewConnection(conn *websocket.Conn) *Connection {
	return &Connection{
		conn: conn,
		send: make(chan []byte, sendQueueLength),
		done: make(chan struct{}),
	}
}

func (c *Connection) Start(onPacket func(packets.Packet) error) {
	c.startOnce.Do(func() {
		c.onPacket = onPacket
		c.pumpWG.Add(2)
		go func() {
			defer c.pumpWG.Done()
			c.writePump()
		}()
		go func() {
			defer c.pumpWG.Done()
			c.readPump()
		}()
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
	c.closeWithStatus(websocket.CloseNormalClosure, "")
}

func (c *Connection) closeWithStatus(code int, reason string) {
	c.closeOnce.Do(func() {
		c.markClosed()
		_ = c.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(code, reason),
			time.Now().Add(writeWait),
		)
		_ = c.conn.Close()
	})
}

func (c *Connection) closeOnError() {
	c.closeOnce.Do(func() {
		c.markClosed()
		_ = c.conn.Close()
	})
}

func (c *Connection) markClosed() {
	c.stateMu.Lock()
	c.closed = true
	close(c.done)
	c.stateMu.Unlock()
}

func (c *Connection) readPump() {
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			c.closeOnError()
			return
		}

		packet, err := packets.Deserialize(message)
		if err != nil {
			c.closeWithStatus(websocket.CloseUnsupportedData, "invalid packet")
			return
		}

		if c.onPacket != nil {
			if err := c.onPacket(packet); err != nil {
				c.closeWithStatus(websocket.ClosePolicyViolation, "packet rejected")
				return
			}
		}
	}
}

func (c *Connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case message := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.closeOnError()
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.closeOnError()
				return
			}
		case <-c.done:
			return
		}
	}
}
