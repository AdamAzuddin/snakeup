package game

import (
	"encoding/json"
	"fmt"
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
	Id        string
	Players   []*player.Player
	State     GameState
	Updates   chan GameState
	Input     chan player.PlayerInput
	Done      chan bool
	Broadcast chan []byte
}

var spawnPoints = []struct{ X, Y int }{
	{5, 5},   // top-left
	{25, 5},  // top-right
	{5, 35},  // bottom-left
	{25, 25}, // bottom-right
}

var startingOffsets = []struct{xOffset, yOffset int}{
	{0,1},
	{0,-1},
	{1,0},
	{-1,0},
}

var colorMux sync.Mutex
var snakeColorCount int

func getColor() player.SnakeColor {
	colorMux.Lock()
	defer colorMux.Unlock()

	// Assign color based on count, cycling through the available ones
	color := player.SnakeColor(snakeColorCount % 4)
	snakeColorCount++
	return color
}

func (g *Game) AddPlayer(p *player.Player) {
	fmt.Printf("Adding player with id: %v \n", p.Id)
	p.SnakeColor = player.SnakeColor(getColor())
	fmt.Printf("Adding player with color id: %v\n", p.SnakeColor)
	if len(g.Players) < len(spawnPoints) {
		p.X = spawnPoints[p.SnakeColor].X
		p.Y = spawnPoints[p.SnakeColor].Y
		p.StartingXOffset = startingOffsets[p.SnakeColor].xOffset
		p.StartingYOffset = startingOffsets[p.SnakeColor].yOffset
	} else {
		// fallback if more players somehow
		p.X, p.Y = 20, 20
	}

	g.Players = append(g.Players, p)

	// Build a "players_update" message with the full list
	var playersData []map[string]interface{}
	for _, pl := range g.Players {
		playersData = append(playersData, map[string]interface{}{
			"playerId":   pl.Id,
			"snakeColor": pl.SnakeColor.String(),
			"x":          pl.X,
			"y":          pl.Y,
		})
	}

	msg := map[string]interface{}{
		"type":    "players_update",
		"gameId":  g.Id,
		"players": playersData,
	}

	data, _ := json.Marshal(msg)

	// Send the full roster to everyone
	g.Broadcast <- data
}

func (g *Game) UpdatePlayersPositions() []byte {
	// update each player's position based on their starting offsets
	for i := range g.Players {
		g.Players[i].X += g.Players[i].StartingXOffset
		g.Players[i].Y += g.Players[i].StartingYOffset
	}

	// build tick message with updated positions
	playersData := make([]map[string]interface{}, len(g.Players))
	for i, p := range g.Players {
		playersData[i] = map[string]interface{}{
			"playerId": p.Id,
			"snakeColor": p.SnakeColor.String(),
			"x":  p.X,
			"y":  p.Y,
		}
	}

	tickMsg := map[string]interface{}{
		"type":    "players_update",
		"gameId":  g.Id,
		"players": playersData,
	}

	data, _ := json.Marshal(tickMsg)
	return data
}

func (*Game) RemovePlayer() error {
	return nil
}
