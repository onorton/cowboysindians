package creature

import (
	"fmt"
	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/message"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

func NewPlayer() *Player {
	return &Player{0, 0, icon.CreatePlayerIcon(), 1, 10, 15, 12, 10, make(map[rune]*item.Item)}
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
	p.str, _ = strconv.Atoi(rest[4])
	p.dex, _ = strconv.Atoi(rest[5])
	var creature Creature = p
	return &creature

}

func (p *Player) Serialize() string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("Player{%d %d %d %d %d %d %s}", p.x, p.y, p.hp, p.ac, p.str, p.dex, p.icon.Serialize())
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
	p.attack(c, GetBonus(p.str), GetBonus(p.str))
}

func (p *Player) TakeDamage(damage int) {
	p.hp -= damage
}
func (p *Player) GetStats() []string {
	stats := make([]string, 3)
	stats[0] = fmt.Sprintf("HP:%d", p.hp)
	stats[1] = fmt.Sprintf("STR:%d(%+d)", p.str, GetBonus(p.str))
	stats[2] = fmt.Sprintf("DEX:%d(%+d)", p.dex, GetBonus(p.dex))
	return stats
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
		p.attack(target, GetBonus(p.str), 0)
	} else {
		message.Enqueue("Your target was too far away.")
	}

}

func (p *Player) PickupItem(item *item.Item) {
	p.inventory[item.GetKey()] = item
}

func (p *Player) GetInventoryKeys() string {
	keysSet := make([]bool, 128)
	for k := range p.inventory {
		keysSet[k] = true
	}
	keys := ""
	for i, _ := range keysSet {
		if i < 33 || i == 127 || !keysSet[i] {
			continue
		}

		if keysSet[i-1] && !keysSet[i+1] {
			keys += string(rune(i))
		} else if !keysSet[i-1] {
			keys += string(rune(i))
		} else if keysSet[i-1] && !keysSet[i-2] && keysSet[i+1] {
			keys += "-"
		}
	}
	return keys
}

func (p *Player) GetItem(key rune) *item.Item {
	item := p.inventory[key]
	delete(p.inventory, key)
	return item
}

func (p *Player) GetInventory() map[rune]*item.Item {
	return p.inventory
}
func GetBonus(score int) int {
	return (score - 10) / 2
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
	inventory  map[rune]*item.Item
}
