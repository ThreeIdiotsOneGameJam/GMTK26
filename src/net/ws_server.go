// coder/websocket only provides the client API on js/wasm, so the server
// endpoint is excluded from web builds.

//go:build !js

package net

import (
	"fmt"
	"log"
	stdnet "net"
	"net/http"
	"strconv"

	"github.com/coder/websocket"
	"github.com/threeidiotsonegamejam/gmtk26/src/server"
)

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	socket, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Browser clients are served from a different origin than this
		// server (e.g. localhost:8080 vs the multiplayer port), so same-origin
		// verification must be skipped or the upgrade gets a 403.
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Println(err)
		return
	}

	connection := NewServerConnection(socket)
	session := server.NewSession(connection)
	defer session.Close()

	connection.Start(session.HandlePacket)
	<-connection.Done()
	connection.wait()
}

func StartWebSocketServer(host string, port uint16) {
	http.HandleFunc("/", websocketHandler)

	address := stdnet.JoinHostPort(host, strconv.FormatUint(uint64(port), 10))

	fmt.Println("server listening on", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatal(err)
	}
}
