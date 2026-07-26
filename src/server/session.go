package server

import (
	"fmt"
	"log"
	"sync"

	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
)

// PacketSender is the transport-facing half of a server session.
// Implementations may be backed by a WebSocket or an in-memory packet queue.
type PacketSender interface {
	SendPacket(packets.Packet) error
}

// LocalPacketSender marks the trusted in-process transport used for solo games.
// Ordinary network transports must not implement this interface.
type LocalPacketSender interface {
	PacketSender
	IsLocalTransport() bool
}

// Session owns the protocol lifecycle shared by every server transport.
type Session struct {
	mu         sync.Mutex
	client     *Client
	registered bool
	closeOnce  sync.Once
}

func NewSession(sender PacketSender) *Session {
	return &Session{client: NewClient(sender)}
}

// HandlePacket validates and dispatches one packet received from a client.
// Recoverable gameplay errors are logged and dropped; fatal protocol and
// transport errors are returned so the caller can close the connection.
func (s *Session) HandlePacket(packet packets.Packet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	clientPacket, ok := packet.(packets.C2SPacket)
	if !ok {
		return fmt.Errorf("received server packet %T from client", packet)
	}

	response, err := s.client.HandlePacket(clientPacket)
	if err != nil {
		log.Printf("failed to handle client packet: %v", err)
		if IsFatalPacketError(err) {
			return err
		}
		if response != nil {
			if sendErr := s.client.SendPacket(response); sendErr != nil {
				return fmt.Errorf("send error response %T: %w", response, sendErr)
			}
		}
		return nil
	}

	_, isConnectPacket := clientPacket.(*packets.C2SConnectPacket)
	newlyRegistered := !s.registered && isConnectPacket
	if !s.registered && !isConnectPacket {
		return fmt.Errorf("client sent %T before connecting", clientPacket)
	}
	if newlyRegistered {
		if err := Connections.Register(s.client); err != nil {
			return err
		}
		s.registered = true
	}

	if response != nil {
		if err := s.client.SendPacket(response); err != nil {
			return fmt.Errorf("send response %T: %w", response, err)
		}
		if newlyRegistered {
			s.client.MarkReady()
		}
	}

	return nil
}

func (s *Session) Close() {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		if s.registered {
			Lobbies.Disconnect(s.client)
			Connections.Unregister(s.client)
		}
		s.client.LeaveGame()
	})
}
