package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/joho/godotenv"
)

// websocket upgrader (http -> ws)
var upgrader = websocket.Upgrader{}

// torrent between home server and global server
var torrent = make(chan CompInstruction)

// list of clients connected (would act on torrent channel)
var clients = []*websocket.Conn{}

func main() {
	// load environment variables from .env file
	godotenv.Load()

	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	go handleAllClients()

	router := mux.NewRouter()
	router.HandleFunc("/", rootPage)
	router.HandleFunc("/ws/", handleNewClients)

	// endpoints
	router.HandleFunc("/instruct/on/", instructOn)
	router.HandleFunc("/instruct/off/", instructOff)
	router.HandleFunc("/instruct/pulse", restartPulse)

	router.Use(loggingMiddleware)
	router.Use(authMiddleware)

	log.Fatal(http.ListenAndServe(":8000", router))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("HTTP Request Received on:", r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["auth"]; !ok {
			log.Println("Error: Authentication not provided!")
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
			log.Println("Error: Authentication not accepted!")
			w.WriteHeader(http.StatusUnauthorized)
			response := OKResponse{
				Code:    401,
				Message: "Authentication not accepted!",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

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
