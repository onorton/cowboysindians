package creature

import (
	"fmt"
	"github.com/onorton/cowboysindians/icon"
	"strconv"
	"strings"
)

func NewPlayer() *Player {
	return &Player{0, 0, icon.CreatePlayerIcon()}
}

func (p *Player) Render(x, y int) {
	p.icon.Render(x, y)
}

func Deserialize(c string) *Player {
	p := new(Player)
	c = c[strings.Index(c, "{")+1 : len(c)-1]
	coordinatesIcon := strings.Split(c, "Icon")
	p.icon = icon.Deserialize(coordinatesIcon[1])

	coordinates := strings.Split(coordinatesIcon[0], " ")
	p.X, _ = strconv.Atoi(coordinates[0])
	p.Y, _ = strconv.Atoi(coordinates[1])
	return p

}

func (p *Player) Serialize() string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("Player{%d %d %s}", p.X, p.Y, p.icon.Serialize())
}

type Player struct {
	X    int
	Y    int
	icon icon.Icon
}
