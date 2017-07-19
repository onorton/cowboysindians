package creature

import (
	"fmt"
	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/message"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

func NewPlayer() *Player {
	return &Player{0, 0, icon.CreatePlayerIcon(), 1, 10, 15, 12, 10}
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
	p.ac, _ = strconv.Atoi(rest[3])
	var creature Creature = p
	return &creature

}

func (p *Player) Serialize() string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("Player{%d %d %d %d %s}", p.x, p.y, p.hp, p.ac, p.icon.Serialize())
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

func (p *Player) attack(c Creature, hitBonus, damageBonus int) {
	if c.AttackHits(rand.Intn(20) + hitBonus + 1) {
		message.Enqueue("You hit the enemy.")
		c.TakeDamage(1 + damageBonus)
	} else {
		message.Enqueue("You miss the enemy.")
	}
	if c.IsDead() {
		message.Enqueue("The enemy died")
	}
}

func (p *Player) MeleeAttack(c Creature) {
	p.attack(c, (p.str-10)/2, (p.str-10)/2)
}

func (p *Player) TakeDamage(damage int) {
	p.hp -= damage
}
func (p *Player) GetHP() int {
	return p.hp
}

func (p *Player) IsDead() bool {
	return p.hp <= 0
}

func (p *Player) AttackHits(roll int) bool {
	return roll > p.ac
}

func (p *Player) RangedAttack(target Creature) {
	tX, tY := target.GetCoordinates()
	distance := math.Sqrt(math.Pow(float64(p.x-tX), 2) + math.Pow(float64(p.y-tY), 2))
	if distance < 10 {
		p.attack(target, (p.str-10)/2, 0)
	} else {
		message.Enqueue("Your target was too far away.")
	}

}

// Interface shared by Player and Enemy
type Creature interface {
	GetCoordinates() (int, int)
	SetCoordinates(int, int)
	Serialize() string
	Render(int, int)
	GetInitiative() int
	MeleeAttack(Creature)
	TakeDamage(int)
	IsDead() bool
	AttackHits(int) bool
}

type Player struct {
	x          int
	y          int
	icon       icon.Icon
	initiative int
	hp         int
	ac         int
	str        int
	dex        int
}
