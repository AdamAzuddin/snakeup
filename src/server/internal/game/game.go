package game


type GameState int
const(
	Init GameState=iota
	Running
	End
)

type Player struct{
	id uint8
	color string
	score int
}

type Game struct{
	id uint16
	players []Player
	state GameState
}

func (*Game) loop()(error){
	return nil
}