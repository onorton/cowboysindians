package creature

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/ui"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func NewPlayer() *Player {
	player := &Player{0, 0, icon.CreatePlayerIcon(), 1, 10, 10, 0, 100, 0, 100, 15, 12, 10, 100, nil, nil, make(map[rune]([]item.Item))}
	player.PickupItem(item.NewWeapon("shotgun"))
	player.PickupItem(item.NewWeapon("sawn-off shotgun"))
	player.PickupItem(item.NewArmour("leather jacket"))
	player.PickupItem(item.NewAmmo("shotgun shell"))
	player.PickupItem(item.NewConsumable("standard ration"))
	player.PickupItem(item.NewConsumable("beer"))
	return player
}

func (p *Player) Render(x, y int) {
	p.icon.Render(x, y)
}

func Deserialize(c string) Creature {
	p := new(Player)
	c = c[strings.Index(c, "{")+1 : len(c)-1]
	restInventory := strings.SplitN(c, "[", 2)
	restWearing := regexp.MustCompile("(Weapon)|(Armour)").Split(restInventory[0], -1)
	wearingTypes := regexp.MustCompile("(Weapon)|(Armour)").FindAllString(restInventory[0], -1)
	rest := strings.Split(restWearing[0], " ")
	inventory := restInventory[1][:len(restInventory[1])-1]

	p.x, _ = strconv.Atoi(rest[0])
	p.y, _ = strconv.Atoi(rest[1])
	p.hp, _ = strconv.Atoi(rest[2])
	p.maxHp, _ = strconv.Atoi(rest[3])
	p.hunger, _ = strconv.Atoi(rest[4])
	p.hunger--
	p.maxHunger, _ = strconv.Atoi(rest[5])
	p.thirst, _ = strconv.Atoi(rest[6])
	p.thirst--
	p.maxThirst, _ = strconv.Atoi(rest[7])
	p.ac, _ = strconv.Atoi(rest[8])
	p.str, _ = strconv.Atoi(rest[9])
	p.dex, _ = strconv.Atoi(rest[10])
	p.encumbrance, _ = strconv.Atoi(rest[11])

	err := json.Unmarshal([]byte(rest[12]), &(p.icon))
	check(err)

	if len(restWearing) > 1 {
		for i := 1; i < len(restWearing); i++ {
			switch wearingTypes[i-1] {
			case "Weapon":
				err := json.Unmarshal([]byte(restWearing[i]), p.weapon)
				check(err)
			case "Armour":
				err := json.Unmarshal([]byte(restWearing[i]), p.armour)
				check(err)
			}
		}
	}
	p.inventory = make(map[rune]([]item.Item))

	items := regexp.MustCompile("(Ammo)|(Armour)|(Consumable)|(Item)|(Weapon)").Split(inventory, -1)
	items = items[1:]
	for _, itemString := range items {
		var itm item.Item
		err := json.Unmarshal([]byte(itemString), &itm)
		check(err)
		p.PickupItem(itm)
	}
	var creature Creature = p
	return creature

}

func (p *Player) Serialize() string {
	if p == nil {
		return ""
	}
	items := "["
	for k, _ := range p.inventory {
		for _, item := range p.inventory[k] {
			itemValue, err := json.Marshal(item)
			check(err)
			items += fmt.Sprintf("%s ", itemValue)

		}
	}

	iconValue, err := json.Marshal(p.icon)
	check(err)

	items += "]"

	weaponValue, err := json.Marshal(p.weapon)
	check(err)

	armourValue, err := json.Marshal(p.armour)
	check(err)

	return fmt.Sprintf("Player{%d %d %d %d %d %d %d %d %d %d %d %d %s %s %s %s}", p.x, p.y, p.hp, p.maxHp, p.hunger, p.maxHunger, p.thirst, p.maxThirst, p.ac, p.str, p.dex, p.encumbrance, iconValue, weaponValue, armourValue, items)
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
		message.Enqueue(fmt.Sprintf("You hit the %s.", c.GetName()))
		if p.weapon != nil {
			c.TakeDamage(p.weapon.GetDamage() + damageBonus)
		} else {
			c.TakeDamage(damageBonus)
		}
	} else {
		message.Enqueue(fmt.Sprintf("You miss the %s.", c.GetName()))
	}
	if c.IsDead() {
		message.Enqueue(fmt.Sprintf("The %s died.", c.GetName()))
	}
}

func (p *Player) MeleeAttack(c Creature) {
	p.attack(c, GetBonus(p.str), GetBonus(p.str))
}

func (p *Player) TakeDamage(damage int) {
	p.hp -= damage
}
func (p *Player) GetStats() []string {
	stats := make([]string, 4)
	stats[0] = fmt.Sprintf("HP:%d/%d", p.hp, p.maxHp)
	stats[1] = fmt.Sprintf("STR:%d(%+d)", p.str, GetBonus(p.str))
	stats[2] = fmt.Sprintf("DEX:%d(%+d)", p.dex, GetBonus(p.dex))
	stats[3] = fmt.Sprintf("AC:%d", p.ac)
	if p.hunger > p.maxHunger/2 {
		stats = append(stats, "Hungry")
	}
	if p.thirst > p.maxThirst/2 {
		stats = append(stats, "Thirsty")
	}
	return stats
}

func (p *Player) PrintInventory() {
	for i, c := range "Wearing: " {
		termbox.SetCell(i, 0, c, termbox.ColorWhite, termbox.ColorDefault)
	}

	position := 2
	if p.weapon != nil {
		for i, c := range fmt.Sprintf("%s - %s", string(p.weapon.GetKey()), p.weapon.GetName()) {
			termbox.SetCell(i, position, c, termbox.ColorWhite, termbox.ColorDefault)
		}
		position++
	}
	if p.armour != nil {
		for i, c := range fmt.Sprintf("%s - %s", string(p.armour.GetKey()), p.armour.GetName()) {
			termbox.SetCell(i, position, c, termbox.ColorWhite, termbox.ColorDefault)
		}
		position++
	}
	position++
	for i, c := range "Inventory: " {
		termbox.SetCell(i, position, c, termbox.ColorWhite, termbox.ColorDefault)
	}
	position += 2

	for k, items := range p.inventory {
		itemString := fmt.Sprintf("%s - %s", string(k), items[0].GetName())
		if len(items) > 1 {
			itemString += fmt.Sprintf(" x%d", len(items))
		}
		for i, c := range itemString {
			termbox.SetCell(i, position, c, termbox.ColorWhite, termbox.ColorDefault)
		}
		position++
	}

	termbox.Flush()
}

func (p *Player) PrintWeapons() {
	position := 0
	for k, items := range p.inventory {
		if _, ok := p.inventory[k][0].(*item.Weapon); !ok {
			continue
		}
		itemString := fmt.Sprintf("%s - %s", string(k), items[0].GetName())
		if len(items) > 1 {
			itemString += fmt.Sprintf(" x%d", len(items))
		}
		for i, c := range itemString {
			termbox.SetCell(i, position, c, termbox.ColorWhite, termbox.ColorDefault)
		}
		position++
	}
	termbox.Flush()
}

func (p *Player) PrintArmour() {
	position := 0
	for k, items := range p.inventory {
		if _, ok := p.inventory[k][0].(*item.Armour); !ok {
			continue
		}
		itemString := fmt.Sprintf("%s - %s", string(k), items[0].GetName())
		if len(items) > 1 {
			itemString += fmt.Sprintf(" x%d", len(items))
		}
		for i, c := range itemString {
			termbox.SetCell(i, position, c, termbox.ColorWhite, termbox.ColorDefault)
		}
		position++
	}
	termbox.Flush()
}

func (p *Player) PrintConsumables() {
	position := 0
	for k, items := range p.inventory {
		if _, ok := p.inventory[k][0].(*item.Consumable); !ok {
			continue
		}
		itemString := fmt.Sprintf("%s - %s", string(k), items[0].GetName())
		if len(items) > 1 {
			itemString += fmt.Sprintf(" x%d", len(items))
		}
		for i, c := range itemString {
			termbox.SetCell(i, position, c, termbox.ColorWhite, termbox.ColorDefault)
		}
		position++
	}
	termbox.Flush()
}
func (p *Player) IsDead() bool {
	return p.hp <= 0 || p.hunger > p.maxHunger || p.thirst > p.maxThirst
}

func (p *Player) AttackHits(roll int) bool {
	return roll > p.ac
}

func (p *Player) RangedAttack(target Creature) {
	p.getAmmo()
	tX, tY := target.GetCoordinates()
	distance := math.Sqrt(math.Pow(float64(p.x-tX), 2) + math.Pow(float64(p.y-tY), 2))
	if distance < float64(p.weapon.GetRange()) {
		p.attack(target, GetBonus(p.str), 0)
	} else {
		message.Enqueue("Your target was too far away.")
	}

}

func (p *Player) getAmmo() *item.Ammo {
	for k, items := range p.inventory {
		if a, ok := items[0].(*item.Ammo); ok && p.weapon.AmmoTypeMatches(a) {
			return p.GetItem(k).(*item.Ammo)
		}
	}
	return nil
}

func (p *Player) PickupItem(itm item.Item) {
	existing := p.inventory[itm.GetKey()]
	if existing == nil {
		existing = make([]item.Item, 0)
	}
	existing = append(existing, itm)
	p.inventory[itm.GetKey()] = existing
}

func (p *Player) GetWeaponKeys() string {
	keysSet := make([]bool, 128)
	for k := range p.inventory {

		if _, ok := p.inventory[k][0].(*item.Weapon); ok {
			keysSet[k] = true
		}
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

func (p *Player) GetArmourKeys() string {
	keysSet := make([]bool, 128)
	for k := range p.inventory {

		if _, ok := p.inventory[k][0].(*item.Armour); ok {
			keysSet[k] = true
		}
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

func (p *Player) GetConsumableKeys() string {
	keysSet := make([]bool, 128)
	for k := range p.inventory {

		if _, ok := p.inventory[k][0].(*item.Consumable); ok {
			keysSet[k] = true
		}
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

func (p *Player) GetItem(key rune) item.Item {
	items := p.inventory[key]
	if items == nil {
		return nil
	}
	item := items[0]
	items = items[1:]
	if len(items) == 0 {
		delete(p.inventory, key)
	} else {
		p.inventory[key] = items
	}
	return item
}

func (p *Player) WieldItem() bool {
	for {
		message.PrintMessage(fmt.Sprintf("What item do you want to wield? [%s or ?*]", p.GetWeaponKeys()))
		s, c := ui.GetItemSelection()

		switch s {
		case ui.All:
			p.PrintInventory()
			continue
		case ui.AllRelevant:
			p.PrintWeapons()
			continue
		case ui.Cancel:
			message.PrintMessage("Never mind.")
			return false
		case ui.SpecificItem:
			itm := p.GetItem(c)
			if itm == nil {
				message.PrintMessage("You don't have that weapon.")
				ui.GetInput()
			} else {
				if w, ok := itm.(*item.Weapon); ok {
					other := p.weapon
					p.weapon = w
					if other != nil {
						p.PickupItem(w)
					}
					message.Enqueue(fmt.Sprintf("You are now wielding a %s.", w.GetName()))
					return true
				} else {
					message.PrintMessage("That is not a weapon.")
					p.PickupItem(itm)
					ui.GetInput()
					return false
				}
			}

		}

	}
}

func (p *Player) WearArmour() bool {
	for {
		message.PrintMessage(fmt.Sprintf("What item do you want to wear? [%s or ?*]", p.GetArmourKeys()))
		s, c := ui.GetItemSelection()

		switch s {
		case ui.All:
			p.PrintInventory()
			continue
		case ui.AllRelevant:
			p.PrintArmour()
			continue
		case ui.Cancel:
			message.PrintMessage("Never mind.")
			return false
		case ui.SpecificItem:
			itm := p.GetItem(c)
			if itm == nil {
				message.PrintMessage("You don't have that piece of armour.")
				ui.GetInput()
			} else {
				if a, ok := itm.(*item.Armour); ok {
					other := p.armour
					p.armour = a
					p.ac += a.GetACBonus()
					if other != nil {
						p.ac -= other.GetACBonus()
						p.PickupItem(a)
					}
					message.Enqueue(fmt.Sprintf("You are now wearing a %s.", a.GetName()))
					return true
				} else {
					message.PrintMessage("That is not a piece of armour.")
					p.PickupItem(itm)
					ui.GetInput()
					return false
				}
			}

		}
	}
}

func (p *Player) ConsumeItem() bool {

	for {
		message.PrintMessage(fmt.Sprintf("What item do you want to eat? [%s or ?*]", p.GetConsumableKeys()))
		s, c := ui.GetItemSelection()

		switch s {
		case ui.All:
			p.PrintInventory()
			continue
		case ui.AllRelevant:
			p.PrintConsumables()
			continue
		case ui.Cancel:
			message.PrintMessage("Never mind.")
			return false
		case ui.SpecificItem:
			itm := p.GetItem(c)
			if itm == nil {
				message.PrintMessage("You don't have that thing to eat.")
				ui.GetInput()
			} else {
				if c, ok := itm.(*item.Consumable); ok {
					message.Enqueue(fmt.Sprintf("You ate a %s.", c.GetName()))

					if c.GetEffect("hunger") > 0 {
						p.eat(c.GetEffect("hunger"))
					}
					if c.GetEffect("hp") > 0 {
						p.heal(c.GetEffect("hp"))
					}
					if c.GetEffect("thirst") > 0 {
						p.drink(c.GetEffect("thirst"))
					}

					return true
				} else {
					message.PrintMessage("That is not something you can eat.")
					p.PickupItem(itm)
					ui.GetInput()
					return false
				}
			}

		}
	}
}

// Check whether player can carry out a range attack this turn
func (p *Player) Ranged() bool {
	if p.weapon != nil {
		return p.weapon.GetRange() > 0
	}
	return false
}

// Check whether player has ammo for particular wielded weapon
func (p *Player) HasAmmo() bool {
	for _, items := range p.inventory {
		if a, ok := items[0].(*item.Ammo); ok && p.weapon.AmmoTypeMatches(a) {
			return true
		}
	}
	return false
}
func GetBonus(score int) int {
	return (score - 10) / 2
}

func (p *Player) heal(amount int) {
	originalHp := p.hp
	p.hp = int(math.Min(float64(originalHp+amount), float64(p.maxHp)))
	message.Enqueue(fmt.Sprintf("You healed for %d hit points.", p.hp-originalHp))
}

func (p *Player) eat(amount int) {
	originalHunger := p.hunger
	p.hunger = int(math.Max(float64(originalHunger-amount), 0.0))
	if originalHunger > p.maxHunger/2 && p.hp <= p.maxHunger/2 {
		message.Enqueue("You are no longer hungry.")
	}

}

func (p *Player) drink(amount int) {
	originalThirst := p.thirst
	p.thirst = int(math.Max(float64(originalThirst-amount), 0.0))
	if originalThirst > p.maxThirst/2 && p.hp <= p.maxThirst/2 {
		message.Enqueue("You are no longer thirsty.")
	}

}

func (p *Player) OverEncumbered() bool {
	total := 0.0
	if p.weapon != nil {
		total += p.weapon.GetWeight()
	}
	if p.armour != nil {
		total += p.weapon.GetWeight()
	}
	for _, items := range p.inventory {
		for _, item := range items {
			total += item.GetWeight()
		}
	}
	return total > float64(p.encumbrance)
}

func (p *Player) GetName() string {
	return "you"
}

func (p *Player) Update() {
	p.hunger++
	p.thirst++
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
	Ranged() bool
	GetName() string
}

type Player struct {
	x           int
	y           int
	icon        icon.Icon
	initiative  int
	maxHp       int
	hp          int
	hunger      int
	maxHunger   int
	thirst      int
	maxThirst   int
	ac          int
	str         int
	dex         int
	encumbrance int
	weapon      *item.Weapon
	armour      *item.Armour
	inventory   map[rune]([]item.Item)
}
