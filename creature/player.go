package creature

import (
	"github.com/onorton/cowboysindians/icon"
)

type Creature interface {
}

func NewPlayer() Player {
	return Player{Creature{}, 0, 0, icon.CreatePlayerIcon}
}

type Player struct {
	creature Creature
	x        int
	y        int
	icon.Icon
}
