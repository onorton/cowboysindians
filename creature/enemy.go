package creature

import (
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/icon"
	"strconv"
	"strings"
)

func NewEnemy(x, y int, c rune, i termbox.Attribute) *Enemy {
	return &Enemy{x, y, icon.NewIcon(c, i)}
}
func (e *Enemy) Render(x, y int) {
	e.icon.Render(x, y)
}

func DeserializeEnemy(c string) *Creature {
	e := new(Enemy)
	c = c[strings.Index(c, "{")+1 : len(c)-1]
	coordinatesIcon := strings.Split(c, "Icon")
	e.icon = icon.Deserialize(coordinatesIcon[1])

	coordinates := strings.Split(coordinatesIcon[0], " ")
	e.x, _ = strconv.Atoi(coordinates[0])
	e.y, _ = strconv.Atoi(coordinates[1])
	var creature Creature = e
	return &creature

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

type Enemy struct {
	x    int
	y    int
	icon icon.Icon
}
