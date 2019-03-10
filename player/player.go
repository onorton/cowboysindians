package player

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"

	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/mount"
	"github.com/onorton/cowboysindians/ui"
	"github.com/onorton/cowboysindians/worldmap"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func NewPlayer(world *worldmap.Map) *Player {
	player := &Player{0, 0, icon.CreatePlayerIcon(), 1, 10, 10, 0, 100, 0, 100, 15, 12, 10, 100, false, nil, nil, make(map[rune]([]item.Item)), nil, world}
	player.AddItem(item.NewWeapon("shotgun"))
	player.AddItem(item.NewWeapon("sawn-off shotgun"))
	player.AddItem(item.NewWeapon("baseball bat"))
	player.AddItem(item.NewArmour("leather jacket"))
	player.AddItem(item.NewAmmo("shotgun shell"))
	player.AddItem(item.NewConsumable("standard ration"))
	player.AddItem(item.NewConsumable("beer"))
	return player
}

func (p *Player) Render() ui.Element {
	if p.mount != nil {
		return icon.MergeIcons(p.icon, p.mount.GetIcon())
	}
	return p.icon.Render()
}

func (p *Player) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	keys := []string{"X", "Y", "Icon", "Initiative", "Hp", "MaxHp", "Hunger", "MaxHunger", "Thirst", "MaxThirst", "AC", "Str", "Dex", "Encumbrance", "Crouching", "Weapon", "Armour", "Inventory"}

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
		"Crouching":   p.crouching,
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
		Crouching   bool
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
	p.crouching = v.Crouching
	p.weapon = v.Weapon
	p.armour = v.Armour
	p.inventory = make(map[rune][]item.Item)

	for _, itm := range v.Inventory {
		p.AddItem(itm)
	}

	return nil
}

func (p *Player) GetCoordinates() (int, int) {
	return p.x, p.y
}

func (p *Player) SetCoordinates(x int, y int) {
	p.x = x
	p.y = y
	if p.mount != nil {
		p.mount.SetCoordinates(x, y)
	}
}

func (p *Player) GetInitiative() int {
	return p.initiative
}

func (p *Player) attack(c worldmap.Creature, hitBonus, damageBonus int) {
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

func (p *Player) MeleeAttack(c worldmap.Creature) {
	p.attack(c, worldmap.GetBonus(p.str), worldmap.GetBonus(p.str))
}

func (p *Player) TakeDamage(damage int) {
	p.hp -= damage
}

func (p *Player) GetStats() []string {
	stats := make([]string, 4)
	stats[0] = fmt.Sprintf("HP:%d/%d", p.hp, p.maxHp)
	stats[1] = fmt.Sprintf("STR:%d(%+d)", p.str, worldmap.GetBonus(p.str))
	stats[2] = fmt.Sprintf("DEX:%d(%+d)", p.dex, worldmap.GetBonus(p.dex))
	stats[3] = fmt.Sprintf("AC:%d", p.ac)
	if p.crouching {
		stats = append(stats, "Crouching")
	}
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

func (p *Player) RangedAttack() bool {
	if !p.ranged() {
		message.PrintMessage("You are not wielding a ranged weapon.")
		return false
	}

	if !p.weaponLoaded() {
		message.PrintMessage("The weapon you are carrying is not loaded.")
		return false
	}

	target := p.findTarget()

	p.weapon.Fire()
	if target == nil {
		message.Enqueue("You fire your weapon at the ground.")
		return true
	}

	tX, tY := target.GetCoordinates()
	distance := math.Sqrt(math.Pow(float64(p.x-tX), 2) + math.Pow(float64(p.y-tY), 2))
	if distance < float64(p.weapon.GetRange()) {
		coverPenalty := 0
		if p.world.TargetBehindCover(p, target) {
			coverPenalty = 5
		}
		p.attack(target, worldmap.GetBonus(p.dex)-coverPenalty, 0)
	} else {
		message.Enqueue("Your target was too far away.")
	}

	return true

}

func (p *Player) getAmmo() *item.Ammo {
	for k, items := range p.inventory {
		if a, ok := items[0].(*item.Ammo); ok && p.weapon.AmmoTypeMatches(a) {
			return p.GetItem(k).(*item.Ammo)
		}
	}
	return nil
}

func (p *Player) AddItem(itm item.Item) {
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
						p.AddItem(other)
					}
					message.Enqueue(fmt.Sprintf("You are now wielding a %s.", w.GetName()))
					return true
				} else {
					message.PrintMessage("That is not a weapon.")
					p.AddItem(itm)
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
						p.AddItem(a)
					}
					message.Enqueue(fmt.Sprintf("You are now wearing a %s.", a.GetName()))
					return true
				} else {
					message.PrintMessage("That is not a piece of armour.")
					p.AddItem(itm)
					ui.GetInput()
					return false
				}
			}

		}
	}
}

func (p *Player) LoadWeapon() bool {
	if !p.ranged() {
		message.PrintMessage("You are not wielding a ranged weapon.")
		return false
	}

	if p.weaponFullyLoaded() {
		message.PrintMessage("The weapon you are wielding is already fully loaded.")
		return false
	}

	if !p.hasAmmo() {
		message.PrintMessage("You don't have ammo for the weapon you are wielding.")
		return false
	}

	for !p.weaponFullyLoaded() && p.hasAmmo() {
		p.getAmmo()
		p.weapon.Load()
	}

	if p.weaponFullyLoaded() {
		message.Enqueue("You have fully loaded your weapon.")
	} else {
		message.Enqueue("You have loaded your weapon.")
	}
	return true
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
					p.AddItem(itm)
					ui.GetInput()
					return false
				}
			}

		}
	}
}

// Check whether player can carry out a range attack this turn
func (p *Player) ranged() bool {
	if p.weapon != nil {
		return p.weapon.GetRange() > 0
	}
	return false
}

// Check whether player is carrying a fully loaded weapon
func (p *Player) weaponFullyLoaded() bool {
	return p.weapon.IsFullyLoaded()
}

// Check whether player has ammo for particular wielded weapon
func (p *Player) hasAmmo() bool {
	for _, items := range p.inventory {
		if a, ok := items[0].(*item.Ammo); ok && p.weapon.AmmoTypeMatches(a) {
			return true
		}
	}
	return false
}

func (p *Player) weaponLoaded() bool {
	if p.weapon != nil && p.weapon.NeedsAmmo() {
		return !p.weapon.IsUnloaded()
	}
	return true

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

func (p *Player) Move(action ui.PlayerAction) (bool, ui.PlayerAction) {
	newX, newY := p.x, p.y

	switch action {
	case ui.MoveWest:
		newX--
	case ui.MoveEast:
		newX++
	case ui.MoveNorth:
		newY--
	case ui.MoveSouth:
		newY++
	case ui.MoveSouthWest:
		newX--
		newY++
	case ui.MoveSouthEast:
		newX++
		newY++
	case ui.MoveNorthWest:
		newX--
		newY--
	case ui.MoveNorthEast:
		newY--
		newX++
	}

	// If out of bounds, reset to original position
	if newX < 0 || newY < 0 || newX >= p.world.GetWidth() || newY >= p.world.GetHeight() {
		newX, newY = p.x, p.y
	}

	c := p.world.GetCreature(newX, newY)
	// If occupied by another creature, melee attack
	if c != nil && c != p {
		mount := c.GetMount()
		if mount != nil {
			message.Enqueue(fmt.Sprintf("The %s is riding a %s. Would you like to target the %s instead? [y/n]", c.GetName(), mount.GetName(), mount.GetName()))
			input := ui.GetInput()
			if input == ui.Confirm {
				p.MeleeAttack(mount)
			}
		} else {
			p.MeleeAttack(c)
		}
		// Will always be NoAction
		return true, ui.NoAction
	}

	if p.mount != nil {

		// If mount has not moved already, player can still do an action
		if !p.mount.Moved() {
			p.world.Move(p, newX, newY)
			p.world.AdjustViewer()
			p.mount.Move()
			return false, ui.NoAction
		} else {
			return true, action
		}
	}

	p.world.Move(p, newX, newY)
	p.world.AdjustViewer()
	return true, ui.NoAction
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
	total_encumbrance := p.encumbrance
	if p.mount != nil {
		total_encumbrance += p.mount.GetEncumbrance()
	}

	return total > float64(total_encumbrance)
}

func (p *Player) findTarget() worldmap.Creature {
	x, y := p.GetCoordinates()
	// In terms of viewer space rather than world space
	rX, rY := x-p.world.GetViewerX(), y-p.world.GetViewerY()
	width, height := p.world.GetWidth(), p.world.GetHeight()
	vWidth, vHeight := p.world.GetViewerWidth(), p.world.GetViewerHeight()
	for {
		message.PrintMessage("Select target")
		ui.DrawElement(rX, rY, ui.NewElement('X', termbox.ColorYellow))
		x, y = p.world.GetViewerX()+rX, p.world.GetViewerY()+rY
		oX, oY := rX, rY

		action := ui.GetInput()
		if action.IsMovementAction() {
			switch action {
			case ui.MoveWest:
				if rX != 0 && x != 0 {
					rX--
				}
			case ui.MoveEast:
				if rX < vWidth-1 && x < width-1 {
					rX++
				}
			case ui.MoveNorth:
				if rY != 0 && y != 0 {
					rY--
				}
			case ui.MoveSouth:
				if rY < vHeight-1 && y < height-1 {
					rY++
				}
			case ui.MoveSouthWest:
				if rX != 0 && rY < vHeight-1 && x != 0 && y < height-1 {
					rX--
					rY++
				}
			case ui.MoveSouthEast:
				if rX < vWidth-1 && rY < vHeight-1 && x < width-1 && y < height-1 {
					rX++
					rY++
				}
			case ui.MoveNorthWest:
				if rX != 0 && rY != 0 && x != 0 && y != 0 {
					rX--
					rY--
				}
			case ui.MoveNorthEast:
				if rY != 0 && rX < vWidth-1 && y != 0 && x < width-1 {
					rY--
					rX++
				}
			}
		} else if action == ui.CancelAction { // Counter intuitive at the moment
			if p.world.IsOccupied(x, y) {
				// If a creature is there, return it.
				c := p.world.GetCreature(x, y)
				m := c.GetMount()
				if m != nil {
					message.Enqueue(fmt.Sprintf("The %s is riding a %s. Would you like to target the %s instead? [y/n]", c.GetName(), m.GetName(), m.GetName()))
					input := ui.GetInput()
					if input == ui.Confirm {
						return c
					}
				}
				return c
			} else {
				message.PrintMessage("Never mind...")
				return nil
			}
		}

		// overwrite
		ui.DrawElement(oX, oY, p.world.RenderTile(x, y))
	}
}

func (p *Player) PickupItem() bool {
	x, y := p.x, p.y
	itemsOnGround := p.world.GetItems(x, y)
	if itemsOnGround == nil {
		message.PrintMessage("There is no item here.")
		return false
	}

	items := make(map[rune]([]item.Item))
	for _, itm := range itemsOnGround {

		existing := items[itm.GetKey()]
		if existing == nil {
			existing = make([]item.Item, 0)
		}
		existing = append(existing, itm)
		items[itm.GetKey()] = existing
	}

	for k := range items {
		for _, item := range items[k] {
			p.AddItem(item)
		}
		if len(items[k]) == 1 {
			message.Enqueue(fmt.Sprintf("You pick up 1 %s.", items[k][0].GetName()))
		} else {
			message.Enqueue(fmt.Sprintf("You pick up %d %ss.", len(items[k]), items[k][0].GetName()))
		}

	}
	return true
}

func (p *Player) DropItem() bool {
	x, y := p.GetCoordinates()
	for {
		message.PrintMessage(fmt.Sprintf("What do you want to drop? [%s or *]", p.GetInventoryKeys()))
		s, c := ui.GetItemSelection()

		switch s {
		case ui.All:
			p.PrintInventory()
			continue
		case ui.Cancel:
			message.PrintMessage("Never mind.")
			return false
		case ui.SpecificItem:
			item := p.GetItem(c)
			if item == nil {
				message.PrintMessage("You don't have that item.")
				ui.GetInput()
			} else {
				p.world.PlaceItem(x, y, item)
				message.Enqueue(fmt.Sprintf("You dropped a %s.", item.GetName()))
				return true
			}
		// Not selectable but still need to consider it
		case ui.AllRelevant:
			return false
		}
	}
}

func (p *Player) ToggleDoor(open bool) bool {
	message.PrintMessage("Which direction?")
	height := p.world.GetHeight()
	width := p.world.GetWidth()
	x, y := p.GetCoordinates()

	// Select direction
	for {
		validMove := true
		action := ui.GetInput()

		if action.IsMovementAction() {
			switch action {
			case ui.MoveWest:
				if x != 0 {
					x--
				}
			case ui.MoveEast:
				if x < width-1 {
					x++
				}
			case ui.MoveNorth:
				if y != 0 {
					y--
				}
			case ui.MoveSouth:
				if y < height-1 {
					y++
				}
			case ui.MoveSouthWest:
				if x != 0 && y < height-1 {
					x--
					y++
				}

			case ui.MoveSouthEast:
				if x < width-1 && y < height-1 {
					x++
					y++
				}
			case ui.MoveNorthWest:
				if x != 0 && y != 0 {
					x--
					y--
				}
			case ui.MoveNorthEast:
				if y != 0 && x < width-1 {
					y--
					x++
				}
			}
		} else if action == ui.CancelAction {
			message.PrintMessage("Never mind...")
			return false
		} else {
			message.PrintMessage("Invalid direction.")
			validMove = false
		}

		if validMove {
			break
		}
	}
	// If there is a door, toggle its position if it's not already there
	if p.world.IsDoor(x, y) {
		if p.world.IsPassable(x, y) != open {
			p.world.SetPassable(x, y, open)
			p.world.SetBlocksVision(x, y, !open)
			if open {
				message.Enqueue("The door opens.")
			} else {
				message.Enqueue("The door closes.")
			}
			return true
		} else {
			if open {
				message.PrintMessage("The door is already open.")
			} else {
				message.PrintMessage("The door is already closed.")
			}
		}
	} else {
		message.PrintMessage("You see no door there.")
	}
	return false
}

func (p *Player) ToggleMount() bool {
	message.PrintMessage("Which direction?")
	height := p.world.GetHeight()
	width := p.world.GetWidth()
	x, y := p.GetCoordinates()

	mounted := p.mount != nil
	action := ui.Wait
	// Select direction
	for {
		validMove := true
		action = ui.GetInput()

		if action.IsMovementAction() {
			switch action {
			case ui.MoveWest:
				if x != 0 {
					x--
				}
			case ui.MoveEast:
				if x < width-1 {
					x++
				}
			case ui.MoveNorth:
				if y != 0 {
					y--
				}
			case ui.MoveSouth:
				if y < height-1 {
					y++
				}
			case ui.MoveSouthWest:
				if x != 0 && y < height-1 {
					x--
					y++
				}

			case ui.MoveSouthEast:
				if x < width-1 && y < height-1 {
					x++
					y++
				}
			case ui.MoveNorthWest:
				if x != 0 && y != 0 {
					x--
					y--
				}
			case ui.MoveNorthEast:
				if y != 0 && x < width-1 {
					y--
					x++
				}
			}
		} else if action == ui.CancelAction {
			message.PrintMessage("Never mind...")
			return false
		} else {
			message.PrintMessage("Invalid direction.")
			validMove = false
		}

		if validMove {
			break
		}
	}

	if mounted {
		if p.world.IsPassable(x, y) {
			p.mount.RemoveRider()
			p.mount = nil
			p.Move(action)
			message.Enqueue("You dismount.")
			return true
		} else {
			message.PrintMessage("You cannot dismount here.")
		}
	} else {
		c := p.world.GetCreature(x, y)
		if m, ok := c.(*mount.Mount); c != nil && ok {
			m.AddRider(p)
			p.world.DeleteCreature(m)
			p.Move(action)
			p.mount = m
			message.Enqueue(fmt.Sprintf("You mount the %s.", m.GetName()))
			return true
		}
		message.PrintMessage("There is no creature to mount here.")

	}

	return false
}

func (p *Player) GetName() string {
	return "you"
}

func (p *Player) GetAlignment() worldmap.Alignment {
	return worldmap.Player
}

func (p *Player) IsCrouching() bool {
	return p.crouching
}

func (p *Player) ToggleCrouch() {
	p.crouching = !p.crouching
	if p.crouching {
		message.Enqueue("You crouch down.")
	} else {
		message.Enqueue("You stand up.")
	}
}

func (p *Player) SetMap(world *worldmap.Map) {
	p.world = world
}

func (p *Player) GetMount() worldmap.Creature {
	return p.mount
}

func (p *Player) Update() {
	p.hunger++
	p.thirst++
	if p.mount != nil {
		p.mount.ResetMoved()
	}
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
	crouching   bool
	weapon      *item.Weapon
	armour      *item.Armour
	inventory   map[rune]([]item.Item)
	mount       *mount.Mount
	world       *worldmap.Map
}
