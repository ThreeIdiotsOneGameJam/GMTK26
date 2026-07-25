package main

import (
	"fmt"

	"github.com/threeidiotsonegamejam/gmtk26/src/net"
)

func main() {
	fmt.Println("server starting")

	net.StartWebSocketServer("0.0.0.0", 58008)
}
