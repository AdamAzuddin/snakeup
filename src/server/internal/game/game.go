package game

type GameState int

const (
	Init GameState = iota
	Running
	End
)

type Player struct {
	Id     uint8
	Color  string
	Score  int
	GameId string
}

type Game struct {
	Id      string
	Players []Player
	State   GameState
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
