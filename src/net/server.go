package net

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}
	}
}

func StartWebSocketServer(ip string, port uint16) {
	http.HandleFunc("/", websocketHandler)

	host := fmt.Sprintf("%s:%d", ip, port)

	fmt.Println("server listening on", host)
	if err := http.ListenAndServe(host, nil); err != nil {
		log.Fatal(err)
	}
}
