package creature

import (
	"fmt"
	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/message"
	"strconv"
	"strings"
)

func NewPlayer() *Player {
	return &Player{0, 0, icon.CreatePlayerIcon(), 1, 10}
}

func (p *Player) Render(x, y int) {
	p.icon.Render(x, y)
}

func Deserialize(c string) *Creature {
	p := new(Player)
	c = c[strings.Index(c, "{")+1 : len(c)-1]
	restIcon := strings.Split(c, "Icon")
	p.icon = icon.Deserialize(restIcon[1])

	rest := strings.Split(restIcon[0], " ")
	p.x, _ = strconv.Atoi(rest[0])
	p.y, _ = strconv.Atoi(rest[1])
	p.hp, _ = strconv.Atoi(rest[2])
	var creature Creature = p
	return &creature

}

func (p *Player) Serialize() string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("Player{%d %d %d %s}", p.x, p.y, p.hp, p.icon.Serialize())
}

func (p *Player) GetCoordinates() (int, int) {

	return p.x, p.y
}

func (p *Player) SetCoordinates(x int, y int) {
	p.x = x
	p.y = y
}

func (p *Player) GetInitiative() int {
	return p.initiative
}

func (p *Player) Attack(c Creature) {
	c.TakeDamage(1)
	message.Enqueue("You hit the enemy.")
	if c.IsDead() {
		message.Enqueue("The enemy died")
	}
}

func (p *Player) TakeDamage(damage int) {
	p.hp -= damage
	message.Enqueue("You took damage.")
}
func (p *Player) GetHP() int {
	return p.hp
}

func (p *Player) IsDead() bool {
	return p.hp <= 0
}

type Creature interface {
	GetCoordinates() (int, int)
	SetCoordinates(int, int)
	Serialize() string
	Render(int, int)
	GetInitiative() int
	Attack(Creature)
	TakeDamage(int)
	IsDead() bool
}

type Player struct {
	x          int
	y          int
	icon       icon.Icon
	initiative int
	hp         int
}
