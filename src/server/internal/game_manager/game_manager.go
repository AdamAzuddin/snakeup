package gamemanager

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/AdamAzuddin/snakeup/server/internal/game"
	"github.com/AdamAzuddin/snakeup/server/internal/player"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const MAX_NO_PLAYERS_IN_A_ROOM = 4

var idCounter int64
var idMutex sync.Mutex

func generatePlayerId() int64 {
	idMutex.Lock()
	defer idMutex.Unlock()
	idCounter++
	return idCounter
}

type GameManager struct {
	games []*game.Game
	mu    sync.Mutex
}

func (gm *GameManager) CreateNewGame(gameId string) *game.Game {
	// spawn a new go routine
	newGame := game.Game{
		Id:            gameId,
		State:         game.Init,
		Players:       make([]*player.Player, 0, MAX_NO_PLAYERS_IN_A_ROOM),
		Width:         178 / 2,
		Height:        100 / 2,
		Updates:       make(chan game.GameState, 100),
		Input:         make(chan player.PlayerInput, 100),
		StopBroadcast: make(chan bool),
		Broadcast:     make(chan []byte, 100),
		Quit:          make(chan struct{}),
	}
	gm.mu.Lock()
	gm.games = append(gm.games, &newGame)
	gm.mu.Unlock()

	go gm.runBroadcaster(&newGame)
	return &newGame
}

func (gm *GameManager) RunGame(g *game.Game) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	tickCount := 0
	fmt.Printf("Running game loop for game id: %s\n", g.Id)
	g.State = game.Running

	// Init players

	var playersData []map[string]interface{}
	for _, pl := range g.Players {
		headPos := pl.Snake.Body.Front().Value.(player.Position) // type assertion
		playersData = append(playersData, map[string]interface{}{
			"playerId":   pl.Id,
			"snakeColor": pl.SnakeColor.String(),
			"x":          headPos.X,
			"y":          headPos.Y})
	}

	// Init apple
	g.Apple = g.GetRandomPosition()

	msg := map[string]interface{}{
		"type":      "game_starting",
		"gameId":    g.Id,
		"players":   playersData,
		"apple":     g.Apple,
		"tickCount": tickCount,
	}
	data, _ := json.Marshal(msg)
	g.Broadcast <- data

	for {
		tickCount++
		select {
		case input := <-g.Input:
			log.Printf("Game %s received input from player %v: %v\n", g.Id, input.PlayerId, input.Movement)
			gm.processPlayerInput(g, &input)

		case <-ticker.C:
			if g.State != game.Running {
				continue
			}

			g.UpdatePlayersPositions()
			winner, loser, isDraw := g.ContainSnakesCollision()
			if winner != nil && loser != nil {
				fmt.Println("Collision detected")
				if !isDraw {
					if winner != loser {
						winner.Score++
					}
					loser.Score = 0
				}
				tickMsg := map[string]interface{}{
					"type":   "game_over",
					"gameId": g.Id,
				}
				data, _ := json.Marshal(tickMsg)
				g.State = game.End
				g.Broadcast <- data
				g.Updates <- game.End
			}

			hasAppleCollision, playerCollidedWithApple := g.ContainAppleCollision()

			if hasAppleCollision {
				playerCollidedWithApple.Score++
				playerCollidedWithApple.Snake.Grow(1)
				g.Apple = g.GetRandomPosition()
				tickMsg := map[string]interface{}{
					"type":     "update_score",
					"playerId": playerCollidedWithApple.Id,
					"newScore": playerCollidedWithApple.Score,
					"apple":    g.Apple,
					"gameId":   g.Id,
				}

				data, _ := json.Marshal(tickMsg)
				g.Broadcast <- data
			}
		case state := <-g.Updates:
			switch state {
			case game.Init:
				fmt.Println("🔄 Game reset detected, resuming loop")
				g.State = game.Running
			case game.End:
				fmt.Println("🛑 Game ended, waiting for reset")
				return
			}
		}

	}
}

func (gm *GameManager) processPlayerInput(g *game.Game, input *player.PlayerInput) {
	for _, pl := range g.Players {
		if pl.Id != input.PlayerId {
			continue
		}

		newX, newY := input.Movement.X, input.Movement.Y

		// Ignore reverse direction: can't move directly opposite current head direction
		currentDir := pl.Snake.Dir
		if (currentDir.X != 0 && newX != 0 && currentDir.X != newX) ||
			(currentDir.Y != 0 && newY != 0 && currentDir.Y != newY) {
			// invalid input: trying to reverse
			return
		}

		// Apply new direction if valid
		if newX != 0 || newY != 0 {
			pl.Snake.Turn(player.Direction{X: newX, Y: newY})
		}

		pl.LastProcessedInputSeq = input.InputSeq
	}
}

func (gm *GameManager) IsGameIdAlreadyExist(gameId string) bool {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	for _, game := range gm.games {
		if game.Id == gameId {
			return true
		}
	}
	return false
}

func (gm *GameManager) IsGameRoomAlreadyFull(gameId string) bool {
	g := gm.GetGameWithId(gameId)
	if g == nil {
		return false
	}
	return len(g.Players) >= MAX_NO_PLAYERS_IN_A_ROOM
}

func (gm *GameManager) runBroadcaster(g *game.Game) {
	for {
		select {
		case msg := <-g.Broadcast:
			for _, p := range g.Players {
				if p.Send != nil {
					select {
					case p.Send <- msg: // enqueue to player's writer goroutine
					default:
						log.Printf("⚠️ Player %d send buffer full, dropping message", p.Id)
					}
				}
			}
		case <-g.StopBroadcast:
			log.Printf("Stopping broadcaster for game %s", g.Id)
			return
		}
	}
}

func (gm *GameManager) GetGameWithId(gameId string) *game.Game {
	for _, g := range gm.games {
		if g.Id == gameId {
			return g
		}
	}
	return nil
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins (for dev)
	},
}

// in game_manager/game_manager.go
func (gm *GameManager) HandleRoomCapacity(w http.ResponseWriter, r *http.Request) {
	// ✅ Always set CORS headers, even for OPTIONS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// ✅ Handle preflight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	gameId := vars["gameId"]

	if gm.IsGameRoomAlreadyFull(gameId) {
		http.Error(w, "Room is full", http.StatusForbidden)
		return
	}

	if !gm.IsGameIdAlreadyExist(gameId) {
		gm.CreateNewGame(gameId)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok","message":"Room joined successfully"}`))
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

	p := &player.Player{
		Id:     uint64(generatePlayerId()),
		GameId: gameId,
		Conn:   conn,
		Send:   make(chan []byte, 50),
	}

	go func(pl *player.Player) {
		for msg := range pl.Send {
			if err := pl.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("❌ Failed to write to player %d: %v", p.Id, err)
				return
			}
		}
	}(p)

	defer p.Conn.Close()

	var currentGame *game.Game

	// check if game id already exist
	if !gm.IsGameIdAlreadyExist(gameId) {
		fmt.Printf("Adding new player to game id %v\n", gameId)
		currentGame = gm.CreateNewGame(gameId)
	} else {
		currentGame = gm.GetGameWithId(gameId)
		if !gm.IsGameRoomAlreadyFull(gameId) {
			fmt.Printf("Adding player to an existing game with id %v\n", gameId)
		} else {
			conn.Close()
			return
		}
	}
	currentGame.AddPlayer(p)
	headPos := p.Snake.Body.Front().Value.(player.Position) // type assertion

	msg := map[string]interface{}{
		"type":       "player_init",
		"gameId":     p.GameId,
		"playerId":   p.Id,
		"snakeColor": p.SnakeColor,
		"XPos":       headPos.X,
		"YPos":       headPos.Y,
		"length":     p.Snake.Body.Len(),
	}

	data, _ := json.Marshal(msg)
	p.Send <- data
	gm.handlePlayerConnection(p, currentGame)
}

func (gm *GameManager) handlePlayerConnection(p *player.Player, g *game.Game) {
	if g == nil {
		log.Printf("❌ handlePlayerConnection called with nil game for player %d", p.Id)
		return
	}
	defer p.Conn.Close()

	for {
		var msg map[string]interface{}
		if err := p.Conn.ReadJSON(&msg); err != nil {
			log.Printf("Player %d disconnected from game %s: %v", p.Id, g.Id, err)
			// TODO: Remove player from game
			close(p.Send)
			break
		}

		// Handle incoming messages from this player
		if msgType, ok := msg["type"].(string); ok {
			switch msgType {
			case "move":
				move := player.MoveMessage{
					Type:     "move",
					PlayerID: uint64(msg["playerId"].(float64)),
					XOffset:  int(msg["xOffset"].(float64)),
					YOffset:  int(msg["yOffset"].(float64)),
				}
				if seq, ok := msg["seq"].(float64); ok {
					move.InputSeq = int(seq)
				}

				log.Printf("Received move from player %v (seq=%v): xOff=%v, yOff=%v\n",
					move.PlayerID, move.InputSeq, move.XOffset, move.YOffset)

				input := player.PlayerInput{
					PlayerId: move.PlayerID,
					Movement: player.Direction{X: move.XOffset,
						Y: move.YOffset},
					InputSeq: move.InputSeq,
				}

				g.Input <- input

			case "start":
				room := msg["room"]
				log.Printf("Player %d started the game in room %v (gameId: %s)", p.Id, room, g.Id)
				go gm.RunGame(g)
			case "reset":
				room := msg["room"]
				log.Printf("Player %d reset the game in room %v (gameId: %s)", p.Id, room, g.Id)

				g.ResetGame()
				msg := map[string]interface{}{
					"type":   "game_resetted",
					"gameId": p.GameId,
				}

				data, _ := json.Marshal(msg)
				g.Broadcast <- data
			}
		}
	}
}
