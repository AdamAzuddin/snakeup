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

const MAX_NO_PLAYERS_IN_A_ROOM = 4

type GameManager struct {
	games []game.Game
}

func (gm *GameManager) CreateNewGame(gameId string) (*game.Game){
	// spawn a new go routine
	newGame := game.Game{
		Id: gameId,
	}
	gm.games = append(gm.games, newGame)
	return &newGame
}

func (*GameManager) RunGame(game *game.Game){

}

func (gm *GameManager) IsGameIdAlreadyExist(gameId string) bool {
	for _, game := range gm.games {
		if game.Id == gameId {
			return true
		}
	}
	return false
}

func (gm *GameManager) IsGameRoomAlreadyFull(id string) bool {
	for _, game := range gm.games {
		if len(game.Players) < MAX_NO_PLAYERS_IN_A_ROOM {
			return true
		}
	}
	return false
}

func (*GameManager) AddPlayerToExistingGame(gameId string, player game.Player) {
	// choose a different color for the snake
}

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

	log.Printf("A client connected to game ID: %s", gameId)

	player := game.Player{
		GameId: gameId,
	}

	// check if game id already exist
	if gm.IsGameIdAlreadyExist(gameId) {
		if gm.IsGameRoomAlreadyFull(gameId) {
			{
				// return error to client
			}
			gm.AddPlayerToExistingGame(gameId, player)
		} else {
			// spawn a new go routine for this specific game

			//gameChan := make([]byte,1024)
			go func() {
				gm.RunGame(gm.CreateNewGame(gameId))
			}()
			gm.AddPlayerToExistingGame(gameId, player)
		}
	}
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
