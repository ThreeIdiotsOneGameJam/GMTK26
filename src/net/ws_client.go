package net

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
)

type ConnectionState int32

const (
	Disconnected ConnectionState = iota
	Connecting
	Connected
)

var client = &WSClient{
	stopCh: make(chan struct{}),
}

func Connect(addr string, onPacket func(packets.Packet)) {
	client.connect(addr, onPacket)
}

func Send(packet packets.Packet) error {
	return client.send(packet)
}

func DrainEvents(onPacket func(packets.Packet)) {
	client.drainEvents(onPacket)
}

func State() ConnectionState {
	return ConnectionState(client.stateAtomic.Load())
}

func Close() {
	client.close()
}

type WSClient struct {
	conn   *websocket.Conn
	connMu sync.Mutex

	stateAtomic atomic.Int32

	events   []packets.Packet
	eventsMu sync.Mutex

	sendMu sync.Mutex

	stopCh    chan struct{}
	closeOnce sync.Once
}

func (c *WSClient) connect(addr string, onPacket func(packets.Packet)) {
	for {
		select {
		case <-c.stopCh:
			return
		default:
		}

		c.setState(Connecting)

		conn, _, err := websocket.DefaultDialer.Dial("ws://"+addr, nil)
		if err != nil {
			c.setState(Disconnected)
			select {
			case <-time.After(5 * time.Second):
			case <-c.stopCh:
				return
			}
			continue
		}

		c.connMu.Lock()
		c.conn = conn
		c.connMu.Unlock()

		c.setState(Connected)

		c.send(&packets.C2SHelloPacket{
			Player: *game.PlayerData,
		})

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				c.closeConn(conn)
				select {
				case <-time.After(5 * time.Second):
				case <-c.stopCh:
					return
				}
				break
			}

			packet, err := packets.Deserialize(message)
			if err != nil {
				continue
			}

			c.eventsMu.Lock()
			c.events = append(c.events, packet)
			c.eventsMu.Unlock()
		}
	}
}

func (c *WSClient) closeConn(conn *websocket.Conn) {
	c.setState(Disconnected)
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	conn.Close()
	c.connMu.Lock()
	c.conn = nil
	c.connMu.Unlock()
}

func (c *WSClient) close() {
	c.closeOnce.Do(func() {
		close(c.stopCh)
	})

	c.connMu.Lock()
	conn := c.conn
	c.connMu.Unlock()

	if conn != nil {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		conn.Close()
	}

	c.connMu.Lock()
	c.conn = nil
	c.connMu.Unlock()

	c.setState(Disconnected)
}

func (c *WSClient) send(packet packets.Packet) error {
	data, err := packets.Serialize(packet)
	if err != nil {
		return err
	}

	c.sendMu.Lock()
	defer c.sendMu.Unlock()

	c.connMu.Lock()
	conn := c.conn
	c.connMu.Unlock()

	if conn == nil {
		return nil
	}

	return conn.WriteMessage(websocket.TextMessage, data)
}

func (c *WSClient) drainEvents(onPacket func(packets.Packet)) {
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
	global.WSState.Store(s.String())
}

func (s ConnectionState) String() string {
	switch s {
	case Disconnected:
		return "Disconnected"
	case Connecting:
		return "Connecting"
	case Connected:
		return "Connected"
	default:
		return "Unknown"
	}
}
