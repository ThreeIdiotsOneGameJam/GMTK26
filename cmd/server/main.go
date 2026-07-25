package main

import (
	"fmt"

	"github.com/threeidiotsonegamejam/gmtk26/src/constants"
	"github.com/threeidiotsonegamejam/gmtk26/src/net"
)

func main() {
	fmt.Println("server starting")

	net.StartWebSocketServer(constants.ServerBindHost, constants.ServerPort)
}
