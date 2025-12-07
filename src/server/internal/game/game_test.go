package game

import (
	"testing"

	"github.com/AdamAzuddin/snakeup/server/internal/player"
	"github.com/AdamAzuddin/snakeup/server/internal/spatial_hash_grid"
)

func TestNewGame(t *testing.T) {
	id := "1"
	shgDimension := player.Position{X: 50, Y: 50}
	worldMax := 1000
	maxPlayer := 50
	chunkSize := 73

	shg := spatial_hash_grid.SpatialHashGrid{
		Bounds: []player.Position{
			{X: -worldMax, Y: -worldMax},
			{X: worldMax, Y: worldMax},
		},
		Dimensions: shgDimension,
		Cells:      make(map[string]*spatial_hash_grid.GridCell),
	}

	g := NewGame(id, worldMax, shg, maxPlayer, chunkSize)

	if g == nil {
		t.Fatal("expected non-nil Game")
	}

	if g.Id != id {
		t.Errorf("expected Id %s, got %s", id, g.Id)
	}

	if g.State != Init {
		t.Errorf("expected State Init, got %v", g.State)
	}

	if g.Width != worldMax || g.Height != worldMax {
		t.Errorf("expected world size %d, got %dx%d", worldMax, g.Width, g.Height)
	}

	if g.WorldGrid.Dimensions != shgDimension {
		t.Errorf("expected world grid dimension (%v,%v), got (%v,%v) ", shgDimension.X, shgDimension.Y, g.WorldGrid.Dimensions.X, g.WorldGrid.Dimensions.Y)
	}
}

func TestInitWalls(t *testing.T) {
	id := "1"
	shgDimension := player.Position{X: 50, Y: 50}
	worldMax := 1000
	maxPlayer := 50
	chunkSize := 73

	shg := spatial_hash_grid.SpatialHashGrid{
		Bounds: []player.Position{
			{X: -worldMax, Y: -worldMax},
			{X: worldMax, Y: worldMax},
		},
		Dimensions: shgDimension,
		Cells:      make(map[string]*spatial_hash_grid.GridCell),
	}

	g := NewGame(id, worldMax, shg, maxPlayer, chunkSize)
	g.InitWalls()

	count := 0
	for x := -g.Width; x+chunkSize <= g.Width; x += chunkSize {
		gridX := (x + g.Width) / chunkSize
		for y := -g.Height; y+chunkSize <= g.Height; y += chunkSize {
			gridY := (y + g.Height) / chunkSize
			if (gridX+gridY)%2 == 0 {
				count++
			}
		}
	}
	expectedChunks := count
	if len(g.Walls) != expectedChunks {
		t.Errorf("expected %v wall chunks, got %v", expectedChunks, len(g.Walls))
	}

	for i, chunk := range g.Walls {
		expectedID := uint64(i + 1)
		if chunk.Id != expectedID {
			t.Errorf("expected chunk ID %d, got %d", expectedID, chunk.Id)
		}

		// At least there's a wall generated for each chunk
		hasTrue := false
		for _, row := range chunk.Grid {
			for _, cell := range row {
				if cell {
					hasTrue = true
					break
				}
			}
			if hasTrue {
				break
			}
		}

		if !hasTrue {
			t.Errorf("expected at least one true value in chunk with id %v", chunk.Id)
		}

	}
}

func TestGeneratePlayerId(t *testing.T) {
	g := NewGame("1", 100, spatial_hash_grid.SpatialHashGrid{}, 10, 10)
	id1 := g.GeneratePlayerId()
	id2 := g.GeneratePlayerId()
	if id2 != id1+1 {
		t.Errorf("expected id2 %d, got %d", id1+1, id2)
	}
}

func TestGetRandomPosition(t *testing.T) {
	g := NewGame("1", 100, spatial_hash_grid.SpatialHashGrid{
		Bounds: []player.Position{{X: -100, Y: -100}, {X: 100, Y: 100}},
		Cells:  make(map[string]*spatial_hash_grid.GridCell),
	}, 10, 10)

	pos := g.GetRandomPosition()
	if pos.X < 0 || pos.X >= g.Width || pos.Y < 0 || pos.Y >= g.Height {
		t.Errorf("position out of bounds: %v", pos)
	}
}