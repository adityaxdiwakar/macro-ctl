package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"

	"flag"
	"net/url"
	"os/signal"
	"strings"

	"github.com/gorilla/websocket"
	"time"
)

func main() {
	// load environment variables
	godotenv.Load()

	var addr = flag.String("addr", os.Getenv("API_URL"), "http service address")
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: os.Getenv("PROTOCOL"), Host: *addr, Path: fmt.Sprintf("/ws/?auth=%s", os.Getenv("AUTH_KEY"))}
	log.Println("Establishing connection with mainstream server, standby...")

	c, _, err := websocket.DefaultDialer.Dial(strings.Replace(u.String(), "%3F", "?", 1), nil)
	if err != nil {
		log.Fatal("Dialup error:", err)
	} else {
		log.Printf("Dialup successful with upstream server!")
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Fatal("Read error:", err)
				return
			}
			message := string(msg) // cast []byte -> string
			log.Println("Message was received:", message)

		}
	}()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			log.Fatal("Channel closed")
			return
		case _ = <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte("[]"))
			if err != nil {
				log.Fatal("Write error:", err)
				return
			}
			log.Println("Message was sent: []")
		case <-interrupt:
			log.Println("Socket interrupted")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Fatal("Error with closing the socket", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
