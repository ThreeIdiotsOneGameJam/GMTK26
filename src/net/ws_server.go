package net

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/server"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	connection := NewConnection(socket)
	client := server.NewClient(connection)
	registered := false
	defer func() {
		if registered {
			server.GameLifecycle.Disconnect(client)
			server.Lobby.RemoveConnection(client)
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
				return client.SendPacket(response)
			}
			return nil
		}

		_, isConnectPacket := clientPacket.(*packets.C2SConnectPacket)
		newlyRegistered := !registered && isConnectPacket
		if !registered && !isConnectPacket {
			return fmt.Errorf("client sent %T before connecting", clientPacket)
		}
		if newlyRegistered {
			if err := server.Lobby.AddConnection(client); err != nil {
				log.Printf("failed to register client: %v", err)
				return err
			}
			registered = true
		}

		if response != nil {
			if err := client.SendPacket(response); err != nil {
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
