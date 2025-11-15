package game

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"

	"github.com/AdamAzuddin/snakeup/server/internal/player"
)

type GameState int

const (
	Init GameState = iota
	Running
	End
)

type Game struct {
	Id            string
	Players       []*player.Player
	Apple         player.Position
	State         GameState
	Width         int
	Height        int
	Updates       chan GameState
	Input         chan player.PlayerInput
	StopBroadcast chan bool
	Broadcast     chan []byte
	Quit          chan struct{}
	idMutex sync.Mutex
	IdCounter int64
	colorMux sync.Mutex
	SnakeColorCount int
}

var startingOffsets = []struct{ xOffset, yOffset int }{
	{0, 1},
	{0, -1},
	{1, 0},
	{-1, 0},
}

func (g *Game)GeneratePlayerId() int64 {
	g.idMutex.Lock()
	defer g.idMutex.Unlock()
	g.IdCounter++
	return g.IdCounter
}


func (g *Game) Shutdown() {
	g.Players = nil
    select {
    case <-g.Quit:
        // already closed / signalled
    default:
        close(g.Quit)
    }

}

func (g *Game) GetRandomPosition() player.Position {
	for {
		pos := player.Position{
			X: rand.Intn(g.Width - 4), // optional offset
			Y: rand.Intn(g.Height - 4),
		}

		collision := false
		for _, pl := range g.Players {
			if pl.Snake == nil || pl.Snake.Body == nil {
				continue
			}

			// check all snake segments
			for e := pl.Snake.Body.Front(); e != nil; e = e.Next() {
				seg := e.Value.(player.Position)
				if seg.X == pos.X && seg.Y == pos.Y {
					collision = true
					break
				}
			}

			if collision {
				break
			}
		}

		if !collision {
			return pos
		}
	}
}

func (g *Game)getColor() player.SnakeColor {
	g.colorMux.Lock()
	defer g.colorMux.Unlock()
	color := player.SnakeColor(g.SnakeColorCount % 4)
	g.SnakeColorCount++
	return color
}

func (g *Game) BroadcastPlayersData() {
	var playersData []map[string]interface{}

	for _, pl := range g.Players {
		// Convert snake body (linked list) into slice of positions
		var bodyPositions []map[string]int
		for e := pl.Snake.Body.Front(); e != nil; e = e.Next() {
			pos := e.Value.(player.Position)
			bodyPositions = append(bodyPositions, map[string]int{
				"x": pos.X,
				"y": pos.Y,
			})
		}

		playersData = append(playersData, map[string]interface{}{
			"playerId":   pl.Id,
			"snakeColor": pl.SnakeColor.String(),
			"body":       bodyPositions, // send the whole body
			"length":     pl.Snake.Body.Len(),
		})
	}

	msg := map[string]interface{}{
		"type":    "players_update",
		"gameId":  g.Id,
		"players": playersData,
		"apple":   g.Apple,
	}

	data, _ := json.Marshal(msg)
	g.Broadcast <- data
}

func (g *Game) AddPlayer(p *player.Player) {
	fmt.Printf("Adding player with id: %v \n", p.Id)
	p.SnakeColor = player.SnakeColor(g.getColor())
	fmt.Printf("Adding player with color id: %v\n", p.SnakeColor)
	if len(g.Players) < 4 {
		pos := g.GetRandomPosition()
		p.Snake = player.NewSnake(pos.X, pos.Y, player.Direction{X: startingOffsets[p.SnakeColor].xOffset, Y: startingOffsets[p.SnakeColor].yOffset})
	} else {
		// fallback if more players somehow
		p.Snake = player.NewSnake(20, 20, player.Direction{X: 1, Y: 0})
	}

	g.Players = append(g.Players, p)

	// Build a "players_update" message with the full list
	g.BroadcastPlayersData()
}

func (g *Game) RemovePlayer(p *player.Player) {
    fmt.Printf("Removing player with id: %v \n", p.Id)
    for i, pl := range g.Players {
        if pl.Id == p.Id {
            g.Players = append(g.Players[:i], g.Players[i+1:]...)
            break
        }
    }
    g.BroadcastPlayersData()
}



// Returns the player whose body was collided into (winner) and
// the player whose head collided (loser), or nil, nil if no collision.
func (g *Game) ContainSnakesCollision() (winner *player.Player, loser *player.Player, isDraw bool) {
	// Track previous and current head positions
	prevHead := make(map[*player.Player]player.Position)
	currHead := make(map[*player.Player]player.Position)

	// Also track body positions (excluding heads)
	bodyPositions := make(map[string]*player.Player)

	// --- STEP 1: scan body and store prev heads ---
	for _, p := range g.Players {
		head := p.Snake.Body.Front().Value.(player.Position)
		prevHead[p] = head

		// store body (skip head)
		e := p.Snake.Body.Front().Next()
		for ; e != nil; e = e.Next() {
			pos := e.Value.(player.Position)
			key := fmt.Sprintf("%v,%v", pos.X, pos.Y)
			bodyPositions[key] = p
		}
	}

	// --- STEP 2: after movement, get new head positions ---
	for _, p := range g.Players {
		head := p.Snake.Body.Front().Value.(player.Position)
		currHead[p] = head
	}

	// --- STEP 3: detect head-head swap ---
	for pA, prevA := range prevHead {
		currA := currHead[pA]
		for pB, prevB := range prevHead {
			if pA == pB {
				continue
			}

			currB := currHead[pB]

			// Did they swap heads?
			if currA == prevB && currB == prevA {
				fmt.Println("Head-swap detected between:", pA.Id, "and", pB.Id)
				return nil, nil, true // DRAW
			}
		}
	}

	// --- STEP 4: detect head-to-body & head-to-head ---
	for _, p := range g.Players {
		head := currHead[p]
		key := fmt.Sprintf("%v,%v", head.X, head.Y)

		// head-to-body
		if hitPlayer, exists := bodyPositions[key]; exists {
			fmt.Println("Collision! Head of", p.Id, "hit body of", hitPlayer.Id)
			return hitPlayer, p, false
		}

		// head-to-head (same tile at same time)
		for other, otherHead := range currHead {
			if other != p && otherHead == head {
				fmt.Println("Head-to-head:", p.Id, "and", other.Id)
				return nil, nil, true // DRAW
			}
		}
	}

	return nil, nil, false // no collision
}

func (g *Game) ContainAppleCollision() (bool, *player.Player) {
	for _, p := range g.Players {
		headPos := p.Snake.Body.Front().Value.(player.Position)
		if headPos.X == g.Apple.X && headPos.Y == g.Apple.Y {
			return true, p
		}
	}
	return false, nil
}

func (g *Game) UpdatePlayersPositions() {
	// update each player's position based on their starting offsets
	for i := range g.Players {
		g.Players[i].Snake.Move(g.Width, g.Height)
	}

	// build tick message with updated positions
	g.BroadcastPlayersData()
}

func (g *Game) ResetGame() {
	g.State = Init
	for _, p := range g.Players {
		pos := g.GetRandomPosition()
		p.Snake = player.NewSnake(pos.X, pos.Y, player.Direction{X: startingOffsets[p.SnakeColor].xOffset, Y: startingOffsets[p.SnakeColor].yOffset})
		p.Score = 0
	}
	var playersData []map[string]interface{}
	for _, pl := range g.Players {
		headPos := pl.Snake.Body.Front().Value.(player.Position)
		playersData = append(playersData, map[string]interface{}{
			"playerId":              pl.Id,
			"snakeColor":            pl.SnakeColor.String(),
			"x":                     headPos.X,
			"y":                     headPos.Y,
			"lastProcessedInputSeq": pl.LastProcessedInputSeq,
			"length":                pl.Snake.Body.Len(),
		})
	}
	msg := map[string]interface{}{
		"type":    "reset_game",
		"gameId":  g.Id,
		"players": playersData,
	}
	data, _ := json.Marshal(msg)
	g.Broadcast <- data
	fmt.Printf("Reset message sent")
}