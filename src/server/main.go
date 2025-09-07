package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Upgrade HTTP requests to WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins (for dev)
	},
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	log.Println("✅ Client connected to /game/1")

	// Start a ticker to send JSON every 50ms
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	// Main loop
	for t := range ticker.C {
		msg := map[string]interface{}{
			"type": "tick",
			"time": t.UnixMilli(),
		}
		data, _ := json.Marshal(msg)

		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Println("❌ Client disconnected:", err)
			return // exit handler, closes conn + stops ticker
		}
	}

}

func main() {
	fmt.Println("Starting server on :42069...")

	http.HandleFunc("/game/1", gameHandler)

	err := http.ListenAndServe(":42069", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
