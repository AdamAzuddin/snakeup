package player

import (
	"container/list"
	"fmt"
	"math/rand"
	"github.com/gorilla/websocket"
)

type Player struct {
	Id                    uint64
	SnakeColor            string
	Score                 int
	GameId                string
	Conn                  *websocket.Conn
	Send                  chan []byte
	LastProcessedInputSeq int
	Snake                 *Snake
}

// ----------------------------
// Snake struct
// ----------------------------
type Position struct {
	X, Y int
}

type Direction struct {
	X, Y int
}

// Pivot represents a turn in the snake's path
type Pivot struct {
	Pos Position
	Dir Direction
}

type Snake struct {
	Body        *list.List // linked list of Positions, head = Front()
	Dir         Direction  // current head direction
	Pivots      []Pivot    // upcoming turns
	GrowPending int        // number of segments to grow
}

// ----------------------------
// Snake methods
// ----------------------------

// Initialize a new snake for a player
func NewSnake(startX, startY int, dir Direction) *Snake {
	s := &Snake{
		Body: list.New(),
		Dir:  dir,
	}
	s.Body.PushFront(Position{X: startX, Y: startY})
	return s
}

// Move advances the snake one step
func (s *Snake) Move(width, height int) {
	head := s.Body.Front().Value.(Position)
	newHead := Position{
		X: (head.X + s.Dir.X + width) % width,
		Y: (head.Y + s.Dir.Y + height) % height,
	}
	s.Body.PushFront(newHead)

	// Process pivot points
	for i := 0; i < len(s.Pivots); i++ {
		pivot := s.Pivots[i]
		tail := s.Body.Back().Value.(Position)
		if tail == pivot.Pos {
			// Tail has reached pivot, remove it
			s.Pivots = append(s.Pivots[:i], s.Pivots[i+1:]...)
			i-- // adjust index
		}
	}

	// Remove tail if not growing
	if s.GrowPending > 0 {
		s.GrowPending--
	} else {
		s.Body.Remove(s.Body.Back())
	}
}

// Turn adds a new pivot when the head changes direction
func (s *Snake) Turn(newDir Direction) {
	head := s.Body.Front().Value.(Position)
	s.Pivots = append(s.Pivots, Pivot{
		Pos: head,
		Dir: newDir,
	})
	s.Dir = newDir
}

// Grow increases snake length by n segments
func (s *Snake) Grow(n int) {
	s.GrowPending += n
}

func GenerateColors(n int) []string {
    colors := make([]string, n)
    for i := 0; i < n; i++ {
        hue := float64(i) * 360.0 / float64(n)
        colors[i] = HSLToHex(hue, 80, 50)
    }

    // Shuffle the colors to avoid similar colors appearing consecutively
    rand.Shuffle(len(colors), func(i, j int) {
        colors[i], colors[j] = colors[j], colors[i]
    })

    return colors
}

func abs(a float64) float64 {
    if a < 0 { return -a }
    return a
}


func HSLToHex(h, s, l float64) string {
	s /= 100
	l /= 100
	c := (1 - abs(2*l-1)) * s
	x := c * (1 - abs(float64(int(h/60)%2)-1))
	m := l - c/2
	var r, g, b float64

	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	r = (r + m) * 255
	g = (g + m) * 255
	b = (b + m) * 255

	return fmt.Sprintf("#%02x%02x%02x", int(r), int(g), int(b))
}

type PlayerInput struct {
	PlayerId uint64
	Movement Direction
	InputSeq int
}

type MoveMessage struct {
	Type     string `json:"type"`
	PlayerID uint64 `json:"playerId"`
	XOffset  int    `json:"xOffset"`
	YOffset  int    `json:"yOffset"`
	InputSeq int    `json:"inputSeq"`
}
