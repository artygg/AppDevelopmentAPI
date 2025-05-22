package websocket

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

var clients = make(map[*websocket.Conn]bool)
var Broadcast = make(chan []byte)

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	clients[conn] = true

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("Client disconnected:", err)
			delete(clients, conn)
			break
		}
	}
}

func HandleMessages() {
	for {
		msg := <-Broadcast
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("Write error:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
