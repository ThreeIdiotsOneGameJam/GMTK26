// coder/websocket only provides the client API on js/wasm, so the server
// endpoint is excluded from web builds.

//go:build !js

package net

import (
	"fmt"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/server"
)

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	socket, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Browser clients are served from a different origin than this
		// server (e.g. localhost:8080 vs localhost:58008), so same-origin
		// verification must be skipped or the upgrade gets a 403.
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Println(err)
		return
	}

	connection := NewServerConnection(socket)
	client := server.NewClient(connection)
	registered := false
	defer func() {
		if registered {
			server.Lobbies.Disconnect(client)
			server.Connections.Unregister(client)
			client.LeaveGame()
		}
	}()

	connection.Start(func(packet packets.Packet) error {
		clientPacket, ok := packet.(packets.C2SPacket)
		if !ok {
			return fmt.Errorf("received server packet %T from client", packet)
		}

		response, err := client.HandlePacket(clientPacket)
		if err != nil {
			log.Printf("failed to handle client packet: %v", err)
			if server.IsFatalPacketError(err) {
				return err
			}
			if response != nil {
				if sendErr := client.SendPacket(response); sendErr != nil {
					log.Printf("failed to send error response %T: %v", response, sendErr)
					return sendErr
				}
			}
			return nil
		}

		_, isConnectPacket := clientPacket.(*packets.C2SConnectPacket)
		newlyRegistered := !registered && isConnectPacket
		if !registered && !isConnectPacket {
			return fmt.Errorf("client sent %T before connecting", clientPacket)
		}
		if newlyRegistered {
			if err := server.Connections.Register(client); err != nil {
				log.Printf("failed to register client: %v", err)
				return err
			}
			registered = true
		}

		if response != nil {
			if err := client.SendPacket(response); err != nil {
				log.Printf("failed to send response %T: %v", response, err)
				return err
			}
			if newlyRegistered {
				client.MarkReady()
			}
		}

		return nil
	})
	<-connection.Done()
	connection.wait()
}

func StartWebSocketServer(ip string, port uint16) {
	http.HandleFunc("/", websocketHandler)

	host := fmt.Sprintf("%s:%d", ip, port)

	fmt.Println("server listening on", host)
	if err := http.ListenAndServe(host, nil); err != nil {
		log.Fatal(err)
	}
}
