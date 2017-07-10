package enemy

import (
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/creature"
	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/worldmap"
	"strconv"
	"strings"
)

func NewEnemy(x, y int, c rune, i termbox.Attribute) *Enemy {
	return &Enemy{x, y, true, icon.NewIcon(c, i)}
}
func (e *Enemy) Render(x, y int) {
	e.icon.Render(x, y)
}

func Deserialize(e string) *creature.Creature {
	enemy := new(Enemy)
	e = e[strings.Index(e, "{")+1 : len(e)-1]
	coordinatesIcon := strings.Split(e, "Icon")
	enemy.icon = icon.Deserialize(coordinatesIcon[1])

	coordinates := strings.Split(coordinatesIcon[0], " ")
	enemy.x, _ = strconv.Atoi(coordinates[0])
	enemy.y, _ = strconv.Atoi(coordinates[1])
	var c creature.Creature = enemy
	return &c

}

func (e *Enemy) Serialize() string {
	return fmt.Sprintf("Enemy{%d %d %s}", e.x, e.y, e.icon.Serialize())
}

func (e *Enemy) GetCoordinates() (int, int) {
	return e.x, e.y
}

func (e *Enemy) SetCoordinates(x int, y int) {

	e.x = x
	e.y = y
}

func (e *Enemy) Update(m worldmap.Map) (int, int) {
	y := e.y
	if e.direction {
		y++
	} else {
		y--
	}
	if y >= m.GetHeight() || y < 0 || m.IsOccupied(e.x, y) || !m.IsPassable(e.x, y) {
		e.direction = !e.direction
		y = e.y
	}

	return e.x, y
}

type Enemy struct {
	x         int
	y         int
	direction bool
	icon      icon.Icon
}
