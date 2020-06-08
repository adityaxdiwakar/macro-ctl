package main

import (
	"encoding/json" // unmarshalling the request
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"        // loading environment vars
	"os/signal" // interrupt signal
	"strings"   // url string manipulation
	"time"      // used for ticker (heartbeat)

	"github.com/gorilla/websocket" // import for websocket boilerplate
	"github.com/joho/godotenv"     // import for .env file
)

type CompInstruction struct {
	Type string `json:"type"`
}

func main() {
	// load environment variables
	godotenv.Load()

	// load URL from env variable
	var addr = flag.String("addr", os.Getenv("API_URL"), "http service address")
	flag.Parse()

	// create interrupt channel (from os.Interrupt)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// creation of parametrized URL (including authentication from env)
	u := url.URL{Scheme: os.Getenv("PROTOCOL"), Host: *addr, Path: fmt.Sprintf("/ws/?auth=%s", os.Getenv("AUTH_KEY"))}
	log.Println("Establishing connection with mainstream server, standby...")

	// dial the websocket and handle errors
	c, _, err := websocket.DefaultDialer.Dial(strings.Replace(u.String(), "%3F", "?", 1), nil)
	if err != nil {
		log.Fatal("Dialup error:", err)
	} else {
		log.Printf("Dialup successful with upstream server!")
	}
	defer c.Close() // prevent closure until main() terminates

	done := make(chan struct{})

	// goroutine to handle incoming messages
	go func() {
		defer close(done)
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Fatal("Read error:", err)
				return
			}
			message := string(msg) // cast []byte -> string
			if message != "[]" {
				instruction := CompInstruction{}
				err := json.Unmarshal([]byte(message), &instruction)
				if err != nil {
					log.Println("An error occured decoding an instruction:", err)
					continue
				}
				log.Println("The instruction was:", instruction)
				switch action := instruction.Type; action {
				case "power_on":
					err := SendMagicPacket("d8:cb:8a:9f:e5:f9")
					if err != nil {
						log.Println("An error occured:", err)
					}
				case "power_off":
					log.Println("The power off action was requested!")
				case "restart_pulse":
					log.Println("The restart pulse action was requested!")
				default:
					log.Println("The action requested is not known to the client!")
				}
			}
		}
	}()

	// ticker for heartbeat
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// handler for closures, heartbeat, and interrupts
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
