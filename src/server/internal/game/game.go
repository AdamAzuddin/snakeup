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
}

var startingOffsets = []struct{ xOffset, yOffset int }{
	{0, 1},
	{0, -1},
	{1, 0},
	{-1, 0},
}

var colorMux sync.Mutex
var snakeColorCount int

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

func getColor() player.SnakeColor {
	colorMux.Lock()
	defer colorMux.Unlock()

	// Assign color based on count, cycling through the available ones
	color := player.SnakeColor(snakeColorCount % 4)
	snakeColorCount++
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
	p.SnakeColor = player.SnakeColor(getColor())
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

// Returns the player whose body was collided into (winner) and
// the player whose head collided (loser), or nil, nil if no collision.
func (g *Game) ContainSnakesCollision() (winner *player.Player, loser *player.Player, isDraw bool) {
    // Map to track all body positions (excluding heads)
    bodyPositions := make(map[string]*player.Player)
    headPositions := make(map[string]*player.Player)

    for _, p := range g.Players {
        // Track head positions separately
        head := p.Snake.Body.Front().Value.(player.Position)
        headKey := fmt.Sprintf("%v,%v", head.X, head.Y)
        headPositions[headKey] = p

        // Skip head, store body positions
        e := p.Snake.Body.Front().Next()
        for ; e != nil; e = e.Next() {
            pos := e.Value.(player.Position)
            key := fmt.Sprintf("%v,%v", pos.X, pos.Y)
            bodyPositions[key] = p
        }
    }

    // Check head collisions
    for _, p := range g.Players {
        head := p.Snake.Body.Front().Value.(player.Position)
        key := fmt.Sprintf("%v,%v", head.X, head.Y)

        // Check head-to-body collision
        if hitPlayer, exists := bodyPositions[key]; exists {
            fmt.Println("Collision detected! Head of player", p.Id, "hit body of player", hitPlayer.Id)
            return hitPlayer, p, false
        }

        // Check head-to-head collision
        if otherPlayer, exists := headPositions[key]; exists && otherPlayer.Id != p.Id {
            fmt.Println("Collision detected! Head of player", p.Id, "hit head of player", otherPlayer.Id)
            return otherPlayer, p, true // arbitrarily treat otherPlayer as winner
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

func (*Game) RemovePlayer() error {
	return nil
}
