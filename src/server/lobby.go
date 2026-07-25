package server

import (
	"fmt"
	"sync"
)

var Lobby = &lobby{}

type lobby struct {
	mu          sync.RWMutex
	connections []*Client
}

func (l *lobby) AddConnection(connection *Client) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	player, identified := connection.Player()
	if !identified {
		return fmt.Errorf("add connection: client is not identified")
	}

	for _, existing := range l.connections {
		existingPlayer, _ := existing.Player()
		if existingPlayer.ClientID == player.ClientID {
			return fmt.Errorf("add connection: client ID %q is already connected", player.ClientID)
		}
	}

	l.connections = append(l.connections, connection)
	return nil
}

func (l *lobby) RemoveConnection(connection *Client) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i, existing := range l.connections {
		if existing == connection {
			last := len(l.connections) - 1
			copy(l.connections[i:], l.connections[i+1:])
			l.connections[last] = nil
			l.connections = l.connections[:last]
			return
		}
	}
}

func (l *lobby) Connections() []*Client {
	l.mu.RLock()
	defer l.mu.RUnlock()

	connections := make([]*Client, 0, len(l.connections))
	for _, connection := range l.connections {
		if connection.Ready() {
			connections = append(connections, connection)
		}
	}

	return connections
}
