package spatial_hash_grid

import (
	"fmt"
	"math"

	"github.com/AdamAzuddin/snakeup/server/internal/player"
	"github.com/AdamAzuddin/snakeup/server/internal/wall"
)

func Clamp01(x float64) float64 {
	return math.Min(math.Max(x, 0.0), 1.0)
}

type Set[T comparable] map[T]struct{}

func NewSet[T comparable]() Set[T] {
	return make(Set[T])
}

func (s Set[T]) Add(v T) {
	s[v] = struct{}{}
}

func (s Set[T]) Has(v T) bool {
	_, ok := s[v]
	return ok
}

func (s Set[T]) Delete(v T) {
	delete(s, v)
}

type GridCell struct {
	Players Set[*player.Player]
	Apples  Set[*player.Apple]
	Walls   Set[player.Position]
}

type NearbyObjects struct {
	Players Set[*player.Player]
	Apples  Set[*player.Apple]
	Walls   Set[player.Position]
}

type SpatialHashGrid struct {
	Bounds     []player.Position
	Dimensions player.Position
	Cells      map[string]*GridCell
}

func NewGridCell() *GridCell {
	return &GridCell{
		Players: NewSet[*player.Player](),
		Apples:  NewSet[*player.Apple](),
		Walls:   NewSet[player.Position](),
	}
}

func (shg *SpatialHashGrid) _GetCellIndex(position player.Position) player.Position {
	x := Clamp01((float64(position.X) - float64(shg.Bounds[0].X)) / (float64(shg.Bounds[1].X) - float64(shg.Bounds[0].X)))
	y := Clamp01((float64(position.Y) - float64(shg.Bounds[0].Y)) / (float64(shg.Bounds[1].Y) - float64(shg.Bounds[0].Y)))

	xIndex := x * (float64(shg.Dimensions.X) - 1)
	yIndex := y * (float64(shg.Dimensions.Y) - 1)

	return player.Position{X: int(xIndex), Y: int(yIndex)}
}

func _Key(x int, y int) string {
	return fmt.Sprintf("%d.%d", x, y)
}

func (shg *SpatialHashGrid) InsertWallChunk(chunk *wall.WallChunk) {
    for y := 0; y < chunk.Height; y++ {
        for x := 0; x < chunk.Width; x++ {
            if chunk.Grid[y][x] {
                worldX := chunk.Position.X + x
                worldY := chunk.Position.Y + y

                idx := shg._GetCellIndex(player.Position{X: worldX, Y: worldY})
                key := _Key(idx.X, idx.Y)

                cell, exists := shg.Cells[key]
                if !exists {
                    cell = NewGridCell()
                    shg.Cells[key] = cell
                }

                pos := &player.Position{X: worldX, Y: worldY}
                cell.Walls.Add(*pos)
            }
        }
    }
}

func (shg *SpatialHashGrid) InsertClient(client *player.Player) {
	x := client.Snake.Body.Front().Value.(player.Position).X
	y := client.Snake.Body.Front().Value.(player.Position).Y

	w := client.Bounds.X
	h := client.Bounds.Y

	i1 := shg._GetCellIndex(player.Position{X: x - w/2, Y: y - h/2})
	i2 := shg._GetCellIndex(player.Position{X: x + w/2, Y: y + h/2})

	client.Indices = append(client.Indices, i1, i2)

	for x := i1.X; x <= i2.X; x++ {
		for y := i1.Y; y <= i2.Y; y++ {
			key := _Key(x, y)

			cell, exists := shg.Cells[key]
			if !exists {
				cell = NewGridCell()
				shg.Cells[key] = cell
			}

			cell.Players.Add(client)
		}
	}

}

func (shg *SpatialHashGrid) InsertApple(apple *player.Apple) {
	idx := shg._GetCellIndex(apple.Position)
	key := _Key(idx.X, idx.Y)

	// Create cell if missing
	cell, exists := shg.Cells[key]
	if !exists {
		cell = NewGridCell()
		shg.Cells[key] = cell
	}

	cell.Apples.Add(apple)
}

func (shg *SpatialHashGrid) NewClient(player *player.Player) {

	player.Bounds = shg.Dimensions
	player.Indices = nil

	shg.InsertClient(player)
}

func (shg *SpatialHashGrid) FindNear(client *player.Player) NearbyObjects {
	result := NearbyObjects{
		Players: make(Set[*player.Player]),
		Apples:  make(Set[*player.Apple]),
		Walls: make(Set[player.Position]),
	}

	headPos := client.Snake.Body.Front().Value.(player.Position)
	bounds := client.Bounds

	x := headPos.X
	y := headPos.Y

	w := bounds.X
	h := bounds.Y

	i1 := shg._GetCellIndex(player.Position{X: x - w/2, Y: y - h/2})
	i2 := shg._GetCellIndex(player.Position{X: x + w/2, Y: y + h/2})

	for cx := i1.X; cx <= i2.X; cx++ {
		for cy := i1.Y; cy <= i2.Y; cy++ {

			key := _Key(cx, cy)
			cell, exists := shg.Cells[key]
			if !exists {
				continue
			}

			// Add nearby players
			for p := range cell.Players {
				result.Players.Add(p)
			}

			// Add nearby apples
			for a := range cell.Apples {
				result.Apples.Add(a)
			}

			for w := range cell.Walls{
				result.Walls.Add(w)
			}
		}
	}

	return result
}

func (shg *SpatialHashGrid) FindNearPosition(pos player.Position) NearbyObjects {
	result := NearbyObjects{
		Players: make(Set[*player.Player]),
		Apples:  make(Set[*player.Apple]),
		Walls:   make(Set[player.Position]),
	}

	idx := shg._GetCellIndex(pos)
	for cx := idx.X - 1; cx <= idx.X+1; cx++ {
		for cy := idx.Y - 1; cy <= idx.Y+1; cy++ {
			key := _Key(cx, cy)
			cell, exists := shg.Cells[key]
			if !exists {
				continue
			}

			for p := range cell.Players {
				result.Players.Add(p)
			}
			for a := range cell.Apples {
				result.Apples.Add(a)
			}
			for w := range cell.Walls {
				result.Walls.Add(w)
			}
		}
	}

	return result
}


func (shg *SpatialHashGrid) UpdateClient(client *player.Player) {
	shg.RemoveClient(client)
	shg.InsertClient(client)
}

func (shg *SpatialHashGrid) RemoveClient(client *player.Player) {
	x := client.Snake.Body.Front().Value.(player.Position).X
	y := client.Snake.Body.Front().Value.(player.Position).Y

	w := client.Bounds.X
	h := client.Bounds.Y

	i1 := shg._GetCellIndex(player.Position{X: x - w/2, Y: y - h/2})
	i2 := shg._GetCellIndex(player.Position{X: x + w/2, Y: y + h/2})

	client.Indices = append(client.Indices, i1)
	client.Indices = append(client.Indices, i2)

	for x := i1.X; x <= i2.X; x++ {
		for y := i1.Y; y <= i2.Y; y++ {
			key := _Key(x, y)
			if cell, exists := shg.Cells[key]; exists {
				cell.Players.Delete(client)
			}
		}
	}

	client.Indices = nil
}
