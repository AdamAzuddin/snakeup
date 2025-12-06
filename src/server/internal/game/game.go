package game

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"

	"github.com/AdamAzuddin/snakeup/server/internal/player"
	"github.com/AdamAzuddin/snakeup/server/internal/spatial_hash_grid"
	"github.com/AdamAzuddin/snakeup/server/internal/wall"
)

type GameState int

const (
	Init GameState = iota
	Running
	End
)

type Game struct {
	Id              string
	Players         map[uint64]*player.Player
	Apple           []*player.Apple
	Walls           []*wall.WallChunk
	ChunkSize int
	State           GameState
	Width           int
	Height          int
	WorldGrid       spatial_hash_grid.SpatialHashGrid
	Updates         chan GameState
	Input           chan player.PlayerInput
	StopBroadcast   chan bool
	Broadcast       chan []byte
	Quit            chan struct{}
	idMutex         sync.Mutex
	IdCounter       int64
	colorMux        sync.Mutex
	SnakeColorCount int
	ColorList       []string
}

var startingOffsets = []struct{ xOffset, yOffset int }{
	{0, 1},
	{0, -1},
	{1, 0},
	{-1, 0},
}

func (g *Game) GeneratePlayerId() int64 {
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

func (g *Game) InitWalls() {
    chunkSize := g.ChunkSize

    startX := -g.Width
    startY := -g.Height
    endX := g.Width
    endY := g.Height

    chunks := []*wall.WallChunk{}
    id := uint64(1)

    gridX := 0
    for x := startX; x+chunkSize <= endX; x += chunkSize {
        gridY := 0
        for y := startY; y+chunkSize <= endY; y += chunkSize {

            // ✅ alternating pattern
            if (gridX+gridY)%2 != 0 {
                gridY++
                continue
            }

            chunk := &wall.WallChunk{
                Id:       id,
                Position: player.Position{X: x, Y: y},
                Width:    chunkSize,
                Height:   chunkSize,
            }

            chunk.GenerateMaze()
            g.WorldGrid.InsertWallChunk(chunk)

            chunks = append(chunks, chunk)
            id++
            gridY++
        }
        gridX++
    }

    g.Walls = chunks
    println("Wall chunks generated:", len(chunks))
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

		if collision {
			continue
		}

		nearby := g.WorldGrid.FindNearPosition(pos)
		for wallPos := range nearby.Walls {
			if wallPos.X == pos.X && wallPos.Y == pos.Y {
				collision = true
				break
			}
		}

		if collision {
			continue
		}

		return pos
	}
}

func (g *Game) getColor() string {
	g.colorMux.Lock()
	defer g.colorMux.Unlock()

	color := g.ColorList[g.SnakeColorCount%len(g.ColorList)]
	g.SnakeColorCount++
	return color
}

func (g *Game) BroadcastPlayersData(p *player.Player) {
	g.WorldGrid.UpdateClient(p)
	nearby := g.WorldGrid.FindNear(p)

	var bodyPositions []map[string]int
	for e := p.Snake.Body.Front(); e != nil; e = e.Next() {
		pos := e.Value.(player.Position)
		bodyPositions = append(bodyPositions, map[string]int{
			"x": pos.X,
			"y": pos.Y,
		})
	}

	playersData := []map[string]interface{}{
		{
			"playerId":   p.Id,
			"snakeColor": p.SnakeColor,
			"body":       bodyPositions,
			"length":     p.Snake.Body.Len(),
		},
	}

	applesData := []map[string]interface{}{}

	for other := range nearby.Players {
		var bodyPositions []map[string]int
		for e := other.Snake.Body.Front(); e != nil; e = e.Next() {
			pos := e.Value.(player.Position)
			bodyPositions = append(bodyPositions, map[string]int{
				"x": pos.X,
				"y": pos.Y,
			})
		}

		playersData = append(playersData, map[string]interface{}{
			"playerId":   other.Id,
			"snakeColor": other.SnakeColor,
			"body":       bodyPositions,
			"length":     other.Snake.Body.Len(),
		})
	}

	for visibleApples := range nearby.Apples {
		applesData = append(applesData, map[string]interface{}{
			"appleId":    visibleApples.Id,
			"appleColor": visibleApples.Color,
			"pos":        visibleApples.Position,
		})
	}

	// Include nearby walls
	wallsData := []map[string]int{}
	for wallPos := range nearby.Walls {
		wallsData = append(wallsData, map[string]int{
			"x": wallPos.X,
			"y": wallPos.Y,
		})
	}

	msg := map[string]interface{}{
		"type":    "players_update",
		"gameId":  g.Id,
		"players": playersData,
		"apples":  applesData,
		"walls":   wallsData, // added walls here
	}

	data, _ := json.Marshal(msg)
	p.Send <- data
}

func (g *Game) AddPlayer(p *player.Player) {
	fmt.Printf("Adding player with id: %v \n", p.Id)
	p.SnakeColor = g.getColor()
	fmt.Printf("Adding player with color id: %v\n", p.SnakeColor)
	if len(g.Players) < 4 {
		pos := g.GetRandomPosition()
		offset := g.GetRandomStartingOffset()
		p.Snake = player.NewSnake(pos.X, pos.Y, offset)
	} else {
		// fallback if more players somehow
		p.Snake = player.NewSnake(20, 20, player.Direction{X: 1, Y: 0})
	}

	g.Players[p.Id] = p
	g.WorldGrid.NewClient(p)

	// Build a "players_update" message with the full list
	g.BroadcastPlayersData(p)
}

func (g *Game) ToClientSpace(target player.Position, viewer player.Position, view player.Position) player.Position {
	cx := view.X / 2
	cy := view.Y / 2

	return player.Position{
		X: target.X - viewer.X + cx,
		Y: target.Y - viewer.Y + cy,
	}
}

func (g *Game) RemovePlayer(p *player.Player) {
	fmt.Printf("Removing player with id: %v \n", p.Id)
	var playerToRemove []map[string]interface{}

	for _, pl := range g.Players {
		if pl.Id == p.Id {
			// Convert snake body (linked list) into slice of positions
			var bodyPositions []map[string]int
			for e := pl.Snake.Body.Front(); e != nil; e = e.Next() {
				pos := e.Value.(player.Position)
				bodyPositions = append(bodyPositions, map[string]int{
					"x": pos.X,
					"y": pos.Y,
				})
			}
			playerToRemove = append(playerToRemove, map[string]interface{}{
				"playerId":   pl.Id,
				"snakeColor": pl.SnakeColor,
				"body":       bodyPositions,
				"length":     pl.Snake.Body.Len(),
			})
		}

	}

	msg := map[string]interface{}{
		"type":     "player_died",
		"gameId":   g.Id,
		"playerId": p.Id,
	}
	data, _ := json.Marshal(msg)
	p.Send <- data

	msg = map[string]interface{}{
		"type":           "remove_player",
		"gameId":         g.Id,
		"playerToRemove": playerToRemove,
	}

	data, _ = json.Marshal(msg)
	g.Broadcast <- data
	delete(g.Players, p.Id)
}

// Returns the player whose body was collided into (winner) and
// the player whose head collided (loser), or nil, nil if no collision.
func (g *Game) ContainSnakesCollision() (containCollision bool, winner *player.Player, loser *player.Player, isDraw bool) {
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
				return true, pA, pB, true // DRAW
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
			return true, hitPlayer, p, false
		}

		// head-to-head (same tile at same time)
		for other, otherHead := range currHead {
			if other != p && otherHead == head {
				fmt.Println("Head-to-head:", p.Id, "and", other.Id)
				return true, p, other, true // DRAW
			}
		}
	}

	return false, nil, nil, false // no collision
}

func (g *Game) ContainAppleCollision() (bool, *player.Player, *player.Apple) {
	for _, p := range g.Players {
		headPos := p.Snake.Body.Front().Value.(player.Position)

		for _, apple := range g.Apple {
			if headPos.X == apple.Position.X && headPos.Y == apple.Position.Y {
				return true, p, apple
			}
		}
	}
	return false, nil, nil
}

func (g *Game) ContainWallCollision() (bool, *player.Player) {
	for _, p := range g.Players {
		head := p.Snake.Body.Front().Value.(player.Position)
		nearby := g.WorldGrid.FindNear(p)

		for w := range nearby.Walls {
			if w.X == head.X && w.Y == head.Y {
				return true, p
			}
		}
	}
	return false, nil
}


func (g *Game) UpdatePlayersPositions() {
	for i := range g.Players {
		g.Players[i].Snake.Move(g.Width, g.Height)
		g.BroadcastPlayersData(g.Players[i])
	}
}

func (g *Game) GetRandomStartingOffset() player.Direction {
	idx := rand.Intn(len(startingOffsets))
	offset := startingOffsets[idx]
	return player.Direction{X: offset.xOffset, Y: offset.yOffset}
}

func (g *Game) ResetGame() {
	g.State = Init
	for _, p := range g.Players {
		pos := g.GetRandomPosition()
		offset := g.GetRandomStartingOffset()
		p.Snake = player.NewSnake(pos.X, pos.Y, offset)
		p.Score = 0
	}
	var playersData []map[string]interface{}
	for _, pl := range g.Players {
		headPos := pl.Snake.Body.Front().Value.(player.Position)
		playersData = append(playersData, map[string]interface{}{
			"playerId":              pl.Id,
			"snakeColor":            pl.SnakeColor,
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
