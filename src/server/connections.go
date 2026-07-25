package server

import (
	"fmt"
	"sync"
)

// Connections is the registry of every identified client connected to the
// server, regardless of whether they are in a lobby or a running game.
var Connections = &connectionRegistry{}

type connectionRegistry struct {
	mu          sync.RWMutex
	connections []*Client
}

func (r *connectionRegistry) Register(connection *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	player, identified := connection.Player()
	if !identified {
		return fmt.Errorf("register connection: client is not identified")
	}

	for _, existing := range r.connections {
		existingPlayer, _ := existing.Player()
		if existingPlayer.ClientID == player.ClientID {
			return fmt.Errorf("register connection: client ID %q is already connected", player.ClientID)
		}
	}

	r.connections = append(r.connections, connection)
	return nil
}

func (r *connectionRegistry) Unregister(connection *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, existing := range r.connections {
		if existing == connection {
			last := len(r.connections) - 1
			copy(r.connections[i:], r.connections[i+1:])
			r.connections[last] = nil
			r.connections = r.connections[:last]
			return
		}
	}
}

// All returns the ready (fully connected) clients.
func (r *connectionRegistry) All() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	connections := make([]*Client, 0, len(r.connections))
	for _, connection := range r.connections {
		if connection.Ready() {
			connections = append(connections, connection)
		}
	}

	return connections
}
