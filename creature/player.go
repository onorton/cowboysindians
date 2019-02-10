package creature

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"

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

func (p *Player) Render() ui.Element {
	return p.icon.Render()
}

func (p *Player) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	keys := []string{"X", "Y", "Icon", "Initiative", "Hp", "MaxHp", "Hunger", "MaxHunger", "Thirst", "MaxThirst", "AC", "Str", "Dex", "Encumbrance", "Weapon", "Armour", "Inventory"}

	playerValues := map[string]interface{}{
		"X":           p.x,
		"Y":           p.y,
		"Icon":        p.icon,
		"Initiative":  p.initiative,
		"Hp":          p.hp,
		"MaxHp":       p.maxHp,
		"Hunger":      p.hunger,
		"MaxHunger":   p.maxHunger,
		"Thirst":      p.thirst,
		"MaxThirst":   p.maxThirst,
		"AC":          p.ac,
		"Str":         p.str,
		"Dex":         p.dex,
		"Encumbrance": p.encumbrance,
		"Weapon":      p.weapon,
		"Armour":      p.armour,
	}

	var inventory []item.Item
	for _, setItems := range p.inventory {
		for _, item := range setItems {
			inventory = append(inventory, item)
		}
	}
	playerValues["Inventory"] = inventory

	length := len(playerValues)
	count := 0

	for _, key := range keys {
		jsonValue, err := json.Marshal(playerValues[key])
		if err != nil {
			return nil, err
		}
		buffer.WriteString(fmt.Sprintf("\"%s\":%s", key, jsonValue))
		count++
		if count < length {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (p *Player) UnmarshalJSON(data []byte) error {

	type playerJson struct {
		X           int
		Y           int
		Icon        icon.Icon
		Initiative  int
		Hp          int
		MaxHp       int
		Hunger      int
		MaxHunger   int
		Thirst      int
		MaxThirst   int
		AC          int
		Str         int
		Dex         int
		Encumbrance int
		Weapon      *item.Weapon
		Armour      *item.Armour
		Inventory   item.ItemList
	}
	v := playerJson{}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	p.x = v.X
	p.y = v.Y
	p.icon = v.Icon
	p.initiative = v.Initiative
	p.hp = v.Hp
	p.maxHp = v.MaxHp
	p.hunger = v.Hunger
	p.maxHunger = v.MaxHunger
	p.thirst = v.Thirst
	p.maxThirst = v.MaxThirst
	p.ac = v.AC
	p.str = v.Str
	p.dex = v.Dex
	p.encumbrance = v.Encumbrance
	p.weapon = v.Weapon
	p.armour = v.Armour
	p.inventory = make(map[rune][]item.Item)

	for _, itm := range v.Inventory {
		p.PickupItem(itm)
	}

	return nil
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
	ui.WriteText(0, 0, "Wearing: ")

	position := 2
	if p.weapon != nil {
		equippedWeaponText := fmt.Sprintf("%s - %s", string(p.weapon.GetKey()), p.weapon.GetName())
		ui.WriteText(0, position, equippedWeaponText)
		position++
	}
	if p.armour != nil {
		equippedArmourText := fmt.Sprintf("%s - %s", string(p.armour.GetKey()), p.armour.GetName())
		ui.WriteText(0, 2, equippedArmourText)
		position++
	}
	position++
	ui.WriteText(0, position, "Inventory: ")
	position += 2

	for k, items := range p.inventory {
		itemString := fmt.Sprintf("%s - %s", string(k), items[0].GetName())
		if len(items) > 1 {
			itemString += fmt.Sprintf(" x%d", len(items))
		}
		ui.WriteText(0, position, itemString)
		position++
	}
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
		ui.WriteText(0, position, itemString)
		position++
	}
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
		ui.WriteText(0, position, itemString)
		position++
	}
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
		ui.WriteText(0, position, itemString)

		position++
	}
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
					log.Panic(itm.(*item.Weapon))
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
	Render() ui.Element
	GetInitiative() int
	MeleeAttack(Creature)
	TakeDamage(int)
	IsDead() bool
	AttackHits(int) bool
	Ranged() bool
	GetName() string
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
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
