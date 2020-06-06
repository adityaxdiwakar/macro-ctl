package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// websocket upgrader (http -> ws)
var upgrader = websocket.Upgrader{}

// torrent between home server and global server
var torrent = make(chan CompInstruction)

// list of clients connected (would act on torrent channel)
var clients = []*websocket.Conn{}

// CompInstruction: all instructions to be sent into websocket
type CompInstruction struct {
	Type string `json:"type"`
}

// OKResponse: all endpoints send 200
type OKResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	go handleAllClients()

	router := mux.NewRouter()
	router.HandleFunc("/", rootPage)
	router.HandleFunc("/ws/", handleNewClients)
	router.HandleFunc("/test/", testMessage)

	router.Use(loggingMiddleware)

	log.Fatal(http.ListenAndServe(":8000", router))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("HTTP Request Received on:", r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func rootPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	response := OKResponse{
		Code:    200,
		Message: "Received Request",
	}
	json.NewEncoder(w).Encode(response)

}

func testMessage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	response := OKResponse{
		Code:    200,
		Message: "Proxying in progress!",
	}
	json.NewEncoder(w).Encode(response)

	message := CompInstruction{
		Type: "Test",
	}

	torrent <- message
}

func handleAllClients() {
	for {
		msg := <-torrent // grab the latest message from the torrent
		for _, session := range clients {
			err := session.WriteJSON(msg)
			if err != nil {
				log.Printf("Could not send CompInstruction, errored: %v", err)
			}
		}
	}
}

func handleNewClients(w http.ResponseWriter, r *http.Request) {

	if _, ok := r.URL.Query()["auth"]; !ok {
		log.Println("Error: Authentication not provided")
		w.WriteHeader(http.StatusUnauthorized)
		response := OKResponse{
			Code:    401,
			Message: "Authentication key not provided!",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	authKey := r.URL.Query()["auth"][0]
	if authKey != "test123" {
		log.Println("Error: Authentication not accepted")
		w.WriteHeader(http.StatusUnauthorized)
		response := OKResponse{
			Code:    401,
			Message: "Authenticcation not accepted!",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

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
			_, message, err := ws.ReadMessage()
			if err != nil {
				log.Println("There was an error on the socket!")
				copy(clients[index:], clients[index+1:])
				clients[len(clients)-1] = nil
				clients = clients[:len(clients)-1]
				return
			}
			log.Println(message)
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
