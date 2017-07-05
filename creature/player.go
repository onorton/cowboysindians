package creature

import (
	"github.com/onorton/cowboysindians/icon"
)

func NewPlayer() Player {
	return Player{0, 0, icon.CreatePlayerIcon()}
}

func (p *Player) Render(x, y int) {
	p.icon.Render(x, y)
}

type Player struct {
	X    int
	Y    int
	icon icon.Icon
}
