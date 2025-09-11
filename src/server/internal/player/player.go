package player

import "github.com/gorilla/websocket"

type Player struct {
	Id         uint8
	SnakeColor SnakeColor
	Score      int
	GameId     string
	Conn       *websocket.Conn
	Send chan []byte
	X, Y int
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
	PlayerId string
	Movement string
	Data     struct{}
}
