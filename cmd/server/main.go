package main

import (
	"fmt"
	"time"

	"github.com/threeidiotsonegamejam/gmtk26/src/net"
	"github.com/threeidiotsonegamejam/gmtk26/src/server"
)

func main() {
	fmt.Println("server starting")

	go func() {
		time.Sleep(2 * time.Second)
		clients := server.Lobby.Connections()
		if len(clients) > 0 {
			fmt.Printf("starting game with %d players\n", len(clients))
			server.GameManager.StartGame(clients, uint8(len(clients)), time.Now().UnixNano())
		}
	}()

	net.StartWebSocketServer("0.0.0.0", 58008)
}
