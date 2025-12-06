package wall

import (
	"math/rand"
	"time"

	"github.com/AdamAzuddin/snakeup/server/internal/player"
)

type WallChunk struct {
	Id       uint64
	Position player.Position
	Width    int
	Height   int
	Grid     [][]bool
}

func (wc *WallChunk) GenerateMaze() {
    // Initialize the grid: all empty
    wc.Grid = make([][]bool, wc.Height)
    for y := 0; y < wc.Height; y++ {
        wc.Grid[y] = make([]bool, wc.Width)
        for x := 0; x < wc.Width; x++ {
            wc.Grid[y][x] = false
        }
    }

    // Parameters: number of blocks and max block size
    numBlocks := (wc.Width * wc.Height) / 64 // adjust density
    maxBlockSize := 3                        // max width/height of a block

    rng := rand.New(rand.NewSource(time.Now().UnixNano()))

    attempts := 0
    placed := 0
    for placed < numBlocks && attempts < numBlocks*10 {
        attempts++

        // Random block size
        w := rng.Intn(maxBlockSize) + 1
        h := rng.Intn(maxBlockSize) + 1

        // Random rotation
        if rng.Intn(2) == 0 {
            w, h = h, w
        }

        // Random position inside chunk
        x := rng.Intn(wc.Width - w + 1)
        y := rng.Intn(wc.Height - h + 1)

        // Check overlap
        overlap := false
        for dy := 0; dy < h && !overlap; dy++ {
            for dx := 0; dx < w; dx++ {
                if wc.Grid[y+dy][x+dx] {
                    overlap = true
                    break
                }
            }
        }

        if overlap {
            continue
        }

        // Place the block
        for dy := 0; dy < h; dy++ {
            for dx := 0; dx < w; dx++ {
                wc.Grid[y+dy][x+dx] = true
            }
        }

        placed++
    }
}


func (wc *WallChunk) PrintGrid() {
	for y := 0; y < len(wc.Grid); y++ {
		row := ""
		for x := 0; x < len(wc.Grid[y]); x++ {
			if wc.Grid[y][x] {
				row += "#"
			} else {
				row += "."
			}
		}
		println(row)
	}
}