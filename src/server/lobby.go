package server

import "github.com/threeidiotsonegamejam/gmtk26/src/net/packets"

var Lobby = &lobby{}

type lobby struct {
	Connections []*packets.Connection
}

func (l *lobby) AddConnection(connection *packets.Connection) {
	l.Connections = append(l.Connections, connection)
}
