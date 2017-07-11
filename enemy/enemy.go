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

func (e *Enemy) GetAIMap(m worldmap.Map) [][]int {
	height, width := m.GetHeight(), m.GetWidth()
	aiMap := make([][]int, height)
	for y := 0; y < height; y++ {
		aiMap[y] = make([]int, width)
		for x := 0; x < width; x++ {
			if m.HasPlayer(x, y) {
				aiMap[y][x] = 0
			} else {
				aiMap[y][x] = height * width
			}
		}
	}
	prev := make([][]int, height)
	for i, _ := range prev {
		prev[i] = make([]int, width)
	}
	for !compareMaps(aiMap, prev) {
		prev = aiMap
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				if !m.IsPassable(x, y) {
					continue
				}
				min := 100
				for i := -1; i <= 1; i++ {
					for j := -1; j <= 1; j++ {
						nX := x + i
						nY := y + j
						if nX >= 0 && nX < width && nY >= 0 && nY < height && aiMap[nY][nX] < min {
							min = aiMap[nY][nX]
						}
					}
					if aiMap[y][x] > min+1 {
						aiMap[y][x] = min + 1
					}
				}

			}
		}
	}
	return aiMap
}

func compareMaps(m, o [][]int) bool {
	for i := 0; i < len(m); i++ {
		for j := 0; j < len(m[0]); j++ {
			if m[i][j] != o[i][j] {
				return false
			}
		}
	}
	return true

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
