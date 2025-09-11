package game

import (
	"github.com/AdamAzuddin/snakeup/server/internal/player"
)

type GameState int

const (
	Init GameState = iota
	Running
	End
)

type Game struct {
	Id      string
	Players []*player.Player
	State   GameState
	Updates chan GameState
	Input  chan player.PlayerInput
	Done    chan bool
	Broadcast chan []byte
}



func (*Game) loop() error {
	return nil
}

func (*Game) AddPlayer() error {
	return nil
}

func (*Game) RemovePlayer() error {
	return nil
}
