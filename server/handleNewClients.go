package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func handleNewClients(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error: Could not upgrade websocket, error was: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		response := OKResponse{
			Code:    400,
			Message: "Could not upgrade your socket, are you using the right standards?",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	defer ws.Close()

	index := len(clients)
	clients = append(clients, ws)

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				log.Println("There was an error on the socket!")
				copy(clients[index:], clients[index+1:])
				clients[len(clients)-1] = nil
				clients = clients[:len(clients)-1]
				return
			}
		}
	}()

	for {
		select {
		case _ = <-ticker.C:
			err := ws.WriteMessage(websocket.TextMessage, []byte("[]"))
			if err != nil {
				log.Println("Error with sending heartbeat: %v", err)
				return
			}
		}
	}
}
