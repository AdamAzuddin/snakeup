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
	State         GameState
	Width         int
	Height        int
	Updates       chan GameState
	Input         chan player.PlayerInput
	StopBroadcast chan bool
	Broadcast     chan []byte
	Quit          chan struct{}
}

type Position struct{
	X int
	Y int
}


var startingOffsets = []struct{ xOffset, yOffset int }{
	{0, 1},
	{0, -1},
	{1, 0},
	{-1, 0},
}

var colorMux sync.Mutex
var snakeColorCount int

func (g *Game) getRandomPosition() Position {
    for {
        pos := Position{
            X: rand.Intn(g.Width-4), // added a bit offset so the snake wont spawn off screen
            Y: rand.Intn(g.Height-4),
        }

        collision := false
        for _, pl := range g.Players {
            if pl.X == pos.X && pl.Y == pos.Y {
                collision = true
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
		playersData = append(playersData, map[string]interface{}{
			"playerId":   pl.Id,
			"snakeColor": pl.SnakeColor.String(),
			"x":          pl.X,
			"y":          pl.Y})
	}
	msg := map[string]interface{}{
		"type":    "players_update",
		"gameId":  g.Id,
		"players": playersData,
	}
	data, _ := json.Marshal(msg)
	g.Broadcast <- data
}

func (g *Game) AddPlayer(p *player.Player) {
	fmt.Printf("Adding player with id: %v \n", p.Id)
	p.SnakeColor = player.SnakeColor(getColor())
	fmt.Printf("Adding player with color id: %v\n", p.SnakeColor)
	if len(g.Players) < 4 {
		pos:=g.getRandomPosition()
		p.X = pos.X
		p.Y = pos.Y
		p.StartingXOffset = startingOffsets[p.SnakeColor].xOffset
		p.StartingYOffset = startingOffsets[p.SnakeColor].yOffset
	} else {
		// fallback if more players somehow
		p.X, p.Y = 20, 20
	}

	g.Players = append(g.Players, p)

	// Build a "players_update" message with the full list
	g.BroadcastPlayersData()
}

func (g *Game) ContainCollisions() bool {
	positions := make(map[string]bool)
	for _, p := range g.Players {
		// check if any set of x AND Y is the same for any of the snakes
		key := fmt.Sprintf("%v,%v", p.X, p.Y)
		fmt.Println("Checking position:", key)
		if positions[key] {
			fmt.Println("Collision detected at:", key)
			return true
		}
		positions[key] = true
	}
	return false
}

func (g *Game) UpdatePlayersPositions() {
	// update each player's position based on their starting offsets
	for i := range g.Players {
		g.Players[i].X = (g.Players[i].X + g.Players[i].StartingXOffset + g.Width) % g.Width
		g.Players[i].Y = (g.Players[i].Y + g.Players[i].StartingYOffset + g.Height) % g.Height
	}

	// build tick message with updated positions
	g.BroadcastPlayersData()
}

func (g *Game) ResetGame() {
	g.State = Init
	for _, p := range g.Players {
		pos:=g.getRandomPosition()
		p.X = pos.X
		p.Y = pos.Y
	}
	var playersData []map[string]interface{}
	for _, pl := range g.Players {
		playersData = append(playersData, map[string]interface{}{
			"playerId":              pl.Id,
			"snakeColor":            pl.SnakeColor.String(),
			"x":                     pl.X,
			"y":                     pl.Y,
			"lastProcessedInputSeq": pl.LastProcessedInputSeq,
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
