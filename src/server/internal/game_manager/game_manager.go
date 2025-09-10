package gamemanager

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/AdamAzuddin/snakeup/server/internal/game"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)


type GameManager struct {
	games []game.Game
}


func (*GameManager) AddNewGame(){

}

func (*GameManager) GetNewOrExistingGame(){

}

// Upgrade HTTP requests to WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins (for dev)
	},
}

func (gm *GameManager) GameHandler(w http.ResponseWriter, r *http.Request) {
	// Extract game ID from URL
	vars := mux.Vars(r)
	gameId := vars["gameId"]

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	log.Printf("✅ Client connected to game ID: %s", gameId)
	
	// Now you can use gameId to manage the specific game
	// todo: assign client to specific game based on the game id
	// if game id doesnt exist, create one. if it does, check if this client can still enter it

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for t := range ticker.C {
		msg := map[string]interface{}{
			"type":   "tick",
			"time":   t.UnixMilli(),
			"gameId": gameId, // Include gameId in response
		}
		data, _ := json.Marshal(msg)

		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("❌ Client disconnected from game %s: %v", gameId, err)
			return
		}
	}
}