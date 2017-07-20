package enemy

import (
	"encoding/json"
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/creature"
	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/worldmap"
	"io/ioutil"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

type EnemyAttributes struct {
	Icon       rune
	Colour     termbox.Attribute
	Initiative int
	Hp         int
	Ac         int
	Str        int
	Dex        int
}

var enemyData map[string]EnemyAttributes = fetchEnemyData()

func fetchEnemyData() map[string]EnemyAttributes {
	data, err := ioutil.ReadFile("data/enemy.json")
	if err != nil {
		panic(err)
	}
	var eD map[string]EnemyAttributes
	json.Unmarshal(data, &eD)
	return eD
}

func NewEnemy(name string, x, y int) *Enemy {
	enemy := enemyData[name]
	return &Enemy{x, y, icon.NewIcon(enemy.Icon, enemy.Colour), enemy.Initiative, enemy.Hp, enemy.Ac, enemy.Str, enemy.Dex}
}
func (e *Enemy) Render(x, y int) {
	e.icon.Render(x, y)
}

func Deserialize(e string) creature.Creature {
	enemy := new(Enemy)
	e = e[strings.Index(e, "{")+1 : len(e)-1]
	restIcon := strings.Split(e, "Icon")
	enemy.icon = icon.Deserialize(restIcon[1])

	rest := strings.Split(restIcon[0], " ")
	enemy.x, _ = strconv.Atoi(rest[0])
	enemy.y, _ = strconv.Atoi(rest[1])
	enemy.hp, _ = strconv.Atoi(rest[2])
	enemy.ac, _ = strconv.Atoi(rest[3])
	enemy.str, _ = strconv.Atoi(rest[4])
	enemy.dex, _ = strconv.Atoi(rest[5])
	var c creature.Creature = enemy
	return c

}

func (e *Enemy) Serialize() string {
	return fmt.Sprintf("Enemy{%d %d %d %d %d %d %s}", e.x, e.y, e.hp, e.ac, e.str, e.dex, e.icon.Serialize())
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

	// Initialise Dijkstra map with goals.
	// Max is size of grid.
	for y := 0; y < height; y++ {
		aiMap[y] = make([]int, width)
		for x := 0; x < width; x++ {
			if m.IsVisible(e, x, y) && m.HasPlayer(x, y) {
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
	// While map changes, update
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

					if aiMap[y][x] > min {
						aiMap[y][x] = min + 1
					}
				}

			}
		}
	}
	return aiMap
}

func (e *Enemy) GetInitiative() int {
	return e.initiative
}

func (e *Enemy) MeleeAttack(c creature.Creature) {
	e.attack(c, creature.GetBonus(e.str), creature.GetBonus(e.str))
}
func (e *Enemy) attack(c creature.Creature, hitBonus, damageBonus int) {

	hits := c.AttackHits(rand.Intn(20) + hitBonus + 1)
	if hits {
		c.TakeDamage(1 + damageBonus)
	}
	if _, ok := c.(*creature.Player); ok {
		if hits {
			message.Enqueue("The enemy hit you.")
		} else {
			message.Enqueue("The enemy missed you.")
		}
	}

}

func (e *Enemy) AttackHits(roll int) bool {
	return roll > e.ac
}
func (e *Enemy) TakeDamage(damage int) {
	e.hp -= damage
}

func (e *Enemy) IsDead() bool {
	return e.hp <= 0
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

type Coordinate struct {
	x int
	y int
}

func (e *Enemy) Update(m worldmap.Map) (int, int) {
	aiMap := e.GetAIMap(m)
	current := aiMap[e.y][e.x]
	possibleLocations := make([]Coordinate, 0)
	// Find adjacent locations closer to the goal
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			nX := e.x + i
			nY := e.y + j
			if nX >= 0 && nX < len(aiMap[0]) && nY >= 0 && nY < len(aiMap) && aiMap[nY][nX] < current {
				possibleLocations = append(possibleLocations, Coordinate{nX, nY})
			}
		}
	}
	target := m.GetPlayer()
	tX, tY := target.GetCoordinates()
	// If close enough and can see target, use ranged attack
	if distance := math.Sqrt(math.Pow(float64(e.x-tX), 2) + math.Pow(float64(e.y-tY), 2)); distance < 10 && m.IsVisible(e, tX, tY) {
		e.attack(target, creature.GetBonus(e.dex), 0)
	} else if len(possibleLocations) > 0 {
		l := possibleLocations[rand.Intn(len(possibleLocations))]
		return l.x, l.y
	}

	return e.x, e.y

}

type Enemy struct {
	x          int
	y          int
	icon       icon.Icon
	initiative int
	hp         int
	ac         int
	str        int
	dex        int
}
