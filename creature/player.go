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
	p.x, _ = strconv.Atoi(coordinates[0])
	p.y, _ = strconv.Atoi(coordinates[1])
	return p

}

func (p *Player) Serialize() string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("Player{%d %d %s}", p.x, p.y, p.icon.Serialize())
}

func (p *Player) GetCoordinates() (int, int) {

	return p.x, p.y
}

func (p *Player) SetCoordinates(x int, y int) {
	p.x = x
	p.y = y
}

type Player struct {
	x    int
	y    int
	icon icon.Icon
}
