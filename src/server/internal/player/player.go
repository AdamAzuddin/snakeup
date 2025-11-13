package player

import (
	"container/list"

	"github.com/gorilla/websocket"
)

type Player struct {
	Id                    uint64
	SnakeColor            SnakeColor
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

type SnakeColor int

const (
	Blue SnakeColor = iota
	Orange
	Purple
	Yellow
)

// String returns the hex color value for each snake color
func (sc SnakeColor) String() string {
	switch sc {
	case Blue:
		return "#2196f3"
	case Orange:
		return "#ff9800"
	case Purple:
		return "#9c27b0"
	case Yellow:
		return "#ffc107"
	default:
		return "#ffffff" // fallback white
	}
}

func GetAllColors() []SnakeColor {
	return []SnakeColor{Blue, Orange, Purple, Yellow}
}

// IsValidColor checks if a given int is a valid SnakeColor
func IsValidColor(color int) bool {
	return color >= 0 && color < 4
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
