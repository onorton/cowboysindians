package player

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"

	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/event"
	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/npc"
	"github.com/onorton/cowboysindians/ui"
	"github.com/onorton/cowboysindians/worldmap"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func NewPlayer(world *worldmap.Map) *Player {
	attributes := map[string]*worldmap.Attribute{
		"hp":          worldmap.NewAttribute(10, 10),
		"hunger":      worldmap.NewAttribute(0, 100),
		"thirst":      worldmap.NewAttribute(0, 100),
		"ac":          worldmap.NewAttribute(15, 15),
		"str":         worldmap.NewAttribute(12, 12),
		"dex":         worldmap.NewAttribute(10, 10),
		"encumbrance": worldmap.NewAttribute(100, 100)}

	attributes["hunger"].AddEffect(item.NewOngoingEffect(1))
	attributes["thirst"].AddEffect(item.NewOngoingEffect(1))

	player := &Player{worldmap.Coordinates{0, 0}, icon.CreatePlayerIcon(), 1, attributes, false, 1000, nil, nil, make(map[rune]([]*item.Item)), "", nil, world}
	player.AddItem(item.NewWeapon("shotgun"))
	player.AddItem(item.NewWeapon("sawn-off shotgun"))
	player.AddItem(item.NewWeapon("baseball bat"))
	player.AddItem(item.NewArmour("leather jacket"))
	player.AddItem(item.NewAmmo("shotgun shell"))
	player.AddItem(item.NewConsumable("standard ration"))
	player.AddItem(item.NewConsumable("beer"))
	event.Subscribe(player)
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

	keys := []string{"Location", "Icon", "Initiative", "Attributes", "Crouching", "Money", "Weapon", "Armour", "Inventory", "MountID"}

	mountID := ""
	if p.mount != nil {
		mountID = p.mount.GetID()
	}

	playerValues := map[string]interface{}{
		"Location":   p.location,
		"Icon":       p.icon,
		"Initiative": p.initiative,
		"Attributes": p.attributes,
		"Money":      p.money,
		"Weapon":     p.weapon,
		"Armour":     p.armour,
		"Crouching":  p.crouching,
		"MountID":    mountID,
	}

	var inventory []*item.Item
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
		Location   worldmap.Coordinates
		Icon       icon.Icon
		Initiative int
		Attributes map[string]*worldmap.Attribute
		Crouching  bool
		Money      int
		Weapon     *item.Item
		Armour     *item.Item
		Inventory  []*item.Item
		MountID    string
	}
	v := playerJson{}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	p.location = v.Location
	p.icon = v.Icon
	p.initiative = v.Initiative
	p.attributes = v.Attributes
	p.crouching = v.Crouching
	p.money = v.Money
	p.weapon = v.Weapon
	p.armour = v.Armour
	p.mountID = v.MountID
	p.inventory = make(map[rune][]*item.Item)

	for _, itm := range v.Inventory {
		p.AddItem(itm)
	}

	return nil
}

func (p *Player) GetCoordinates() (int, int) {
	return p.location.X, p.location.Y
}

func (p *Player) SetCoordinates(x int, y int) {
	p.location.X = x
	p.location.Y = y
	if p.mount != nil {
		p.mount.SetCoordinates(x, y)
	}
}

func (p *Player) GetInitiative() int {
	return p.initiative
}

func (p *Player) Weapon() item.WeaponComponent {
	return p.weapon.Component("weapon").(item.WeaponComponent)
}

func (p *Player) attack(c worldmap.Creature, hitBonus, damageBonus int) {
	if c.AttackHits(rand.Intn(20) + hitBonus + 1) {
		message.Enqueue(fmt.Sprintf("You hit %s.", c.GetName().WithDefinite()))
		if p.weapon != nil {
			c.TakeDamage(p.Weapon().GetDamage() + damageBonus)
		} else {
			c.TakeDamage(damageBonus)
		}
	} else {
		message.Enqueue(fmt.Sprintf("You miss %s.", c.GetName().WithDefinite()))
	}
	if c.IsDead() {
		message.Enqueue(fmt.Sprintf("%s died.", c.GetName().WithDefinite()))

		// If non-enemy dead, send murder event
		if c.GetAlignment() == worldmap.Neutral {
			event.Emit(event.NewMurder(p, c, p.location))
		}
	}
}

func (p *Player) MeleeAttack(c worldmap.Creature) {
	p.attack(c, worldmap.GetBonus(p.attributes["str"].Value()), worldmap.GetBonus(p.attributes["str"].Value()))
}

func (p *Player) TakeDamage(damage int) {
	p.attributes["hp"].AddEffect(item.NewInstantEffect(-damage))
}

func (p *Player) GetStats() []string {
	stats := make([]string, 5)
	stats[0] = fmt.Sprintf("HP:%s", p.attributes["hp"].Status())
	stats[1] = fmt.Sprintf("STR:%d(%+d)", p.attributes["str"].Value(), worldmap.GetBonus(p.attributes["str"].Value()))
	stats[2] = fmt.Sprintf("DEX:%d(%+d)", p.attributes["dex"].Value(), worldmap.GetBonus(p.attributes["dex"].Value()))
	stats[3] = fmt.Sprintf("AC:%d", p.attributes["ac"].Value())
	stats[4] = fmt.Sprintf("$%.2f", float64(p.money)/100)
	if p.crouching {
		stats = append(stats, "Crouching")
	}
	if p.attributes["hunger"].Value() > p.attributes["hunger"].Maximum()/2 {
		stats = append(stats, "Hungry")
	}
	if p.attributes["thirst"].Value() > p.attributes["thirst"].Maximum()/2 {
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
		if !p.inventory[k][0].HasComponent("weapon") {
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
		if !p.inventory[k][0].HasComponent("armour") {
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
		if !p.inventory[k][0].HasComponent("consumable") {
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

func (p *Player) PrintReadables() {
	position := 0
	for k, items := range p.inventory {
		if !p.inventory[k][0].HasComponent("readable") {
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
	return p.attributes["hp"].Value() == 0 || p.attributes["hunger"].Value() == p.attributes["hunger"].Maximum() || p.attributes["thirst"].Value() == p.attributes["thirst"].Maximum()
}

func (p *Player) AttackHits(roll int) bool {
	return roll > p.attributes["ac"].Value()
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

	p.Weapon().Fire()
	if target == nil {
		message.Enqueue("You fire your weapon at the ground.")
		return true
	}

	tX, tY := target.GetCoordinates()
	distance := worldmap.Distance(p.location.X, p.location.Y, tX, tY)
	if distance < float64(p.Weapon().Range) {
		coverPenalty := 0
		if p.world.TargetBehindCover(p, target) {
			coverPenalty = 5
		}
		p.attack(target, worldmap.GetBonus(p.attributes["ac"].Value())-coverPenalty, 0)
	} else {
		message.Enqueue("Your target was too far away.")
	}

	return true

}

func (p *Player) getAmmo() *item.Item {
	for k, items := range p.inventory {
		if items[0].HasComponent("ammo") && p.Weapon().AmmoTypeMatches(items[0]) {
			return p.GetItem(k)
		}
	}
	return nil
}

func (p *Player) AddItem(itm *item.Item) {
	existing := p.inventory[itm.GetKey()]
	if existing == nil {
		existing = make([]*item.Item, 0)
	}
	existing = append(existing, itm)
	p.inventory[itm.GetKey()] = existing
	// If item had previous owner, send theft event
	if !itm.Owned(p.GetID()) {
		event.Emit(event.NewTheft(p, itm, p.location))
	}
	itm.TransferOwner(p.GetID())
}

func (p *Player) GetWeaponKeys() string {
	keysSet := make([]bool, 128)
	for k := range p.inventory {

		if p.inventory[k][0].HasComponent("weapon") {
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

		if p.inventory[k][0].HasComponent("armour") {
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

		if p.inventory[k][0].HasComponent("consumable") {
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

func (p *Player) GetReadableKeys() string {
	keysSet := make([]bool, 128)
	for k := range p.inventory {

		if p.inventory[k][0].HasComponent("readable") {
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

func (p *Player) GetItem(key rune) *item.Item {
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
				if itm.HasComponent("weapon") {
					other := p.weapon
					p.weapon = itm
					if other != nil {
						p.AddItem(other)
					}
					message.Enqueue(fmt.Sprintf("You are now wielding a %s.", itm.GetName()))
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
				if itm.HasComponent("armour") {
					other := p.armour
					p.armour = itm
					if other != nil {
						p.AddItem(other)
					}
					message.Enqueue(fmt.Sprintf("You are now wearing a %s.", itm.GetName()))
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
		p.Weapon().Load()
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
				if itm.HasComponent("consumable") {
					message.Enqueue(fmt.Sprintf("You ate a %s.", itm.GetName()))
					p.consume(itm)
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
		return p.Weapon().Range > 0
	}
	return false
}

// Check whether player is carrying a fully loaded weapon
func (p *Player) weaponFullyLoaded() bool {
	return p.Weapon().IsFullyLoaded()
}

// Check whether player has ammo for particular wielded weapon
func (p *Player) hasAmmo() bool {
	for _, items := range p.inventory {
		if items[0].HasComponent("ammo") && p.Weapon().AmmoTypeMatches(items[0]) {
			return true
		}
	}
	return false
}

func (p *Player) weaponLoaded() bool {
	if p.weapon != nil && p.Weapon().NeedsAmmo() {
		return !p.Weapon().IsUnloaded()
	}
	return true

}

func (p *Player) consume(itm *item.Item) {
	originalHp := p.attributes["hp"].Value()
	originalHunger := p.attributes["hunger"].Value()
	originalThirst := p.attributes["thirst"].Value()

	for attr, attribute := range p.attributes {
		for _, effect := range itm.Component("consumable").(item.ConsumableComponent).Effects[attr] {
			attribute.AddEffect(&effect)
		}
	}

	if p.attributes["hp"].Value() > originalHp {
		message.Enqueue(fmt.Sprintf("You healed for %d hit points.", p.attributes["hp"].Value()-originalHp))
	}

	hungerThresh := p.attributes["hunger"].Maximum() / 2
	if originalHunger > hungerThresh && p.attributes["hunger"].Value() <= hungerThresh {
		message.Enqueue("You are no longer hungry.")
	}

	thirstThresh := p.attributes["thirst"].Maximum() / 2
	if originalThirst > hungerThresh && p.attributes["thirst"].Value() <= thirstThresh {
		message.Enqueue("You are no longer thirsty.")
	}
}

func (p *Player) Move(action ui.PlayerAction) (bool, ui.PlayerAction) {
	newX, newY := p.location.X, p.location.Y

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
		newX, newY = p.location.X, p.location.Y
	}

	c := p.world.GetCreature(newX, newY)
	// If occupied by another creature, melee attack
	if c != nil && c != p {
		var m *npc.Mount
		if r, ok := c.(npc.Rider); ok {
			m = r.Mount()
		}

		if m != nil {
			message.PrintMessage(fmt.Sprintf("%s is riding %s. Would you like to target %s instead? [yn]", c.GetName().WithDefinite(), m.GetName().WithIndefinite(), m.GetName().WithDefinite()))

			input := ui.GetInput()
			if input == ui.Confirm {
				p.MeleeAttack(m)
				return true, ui.NoAction
			}
		}

		p.MeleeAttack(c)

		// Will always be NoAction
		return true, ui.NoAction
	}

	if p.mount != nil {

		// If mount has not moved already, player can still do an action
		if !p.mount.Moved() {
			p.world.MovePlayer(p, newX, newY)
			p.world.AdjustViewer()
			p.mount.Move()
			return false, ui.NoAction
		} else {
			return true, action
		}
	}

	p.world.MovePlayer(p, newX, newY)
	p.world.AdjustViewer()
	return true, ui.NoAction
}

func (p *Player) OverEncumbered() bool {
	total := 0.0
	if p.weapon != nil {
		total += p.weapon.GetWeight()
	}
	if p.armour != nil {
		total += p.armour.GetWeight()
	}
	for _, items := range p.inventory {
		for _, item := range items {
			total += item.GetWeight()
		}
	}
	total_encumbrance := p.attributes["encumbrance"].Value()
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

				var m *npc.Mount
				if r, ok := c.(npc.Rider); ok {
					m = r.Mount()
				}

				if m != nil {
					message.PrintMessage(fmt.Sprintf("%s is riding %s. Would you like to target %s instead? [yn]", c.GetName().WithDefinite(), m.GetName().WithIndefinite(), m.GetName().WithDefinite()))

					input := ui.GetInput()
					if input == ui.Confirm {
						return m
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
	x, y := p.location.X, p.location.Y
	itemsOnGround := p.world.GetItems(x, y)
	if itemsOnGround == nil {
		message.PrintMessage("There is no item here.")
		return false
	}

	// find money
	for i, item := range itemsOnGround {
		if item.GetName() == "money" {
			// If item had previous owner, send theft event
			if !item.Owned(p.GetID()) {
				event.Emit(event.NewTheft(p, item, p.location))
			}
			p.money += item.GetValue()
			message.Enqueue(fmt.Sprintf("You pick up $%.2f.", float64(item.GetValue())/100))
			itemsOnGround = append(itemsOnGround[:i], itemsOnGround[i+1:]...)

		}
	}

	items := make(map[rune]([]*item.Item))
	for _, itm := range itemsOnGround {
		existing := items[itm.GetKey()]
		if existing == nil {
			existing = make([]*item.Item, 0)
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
			p.world.ToggleDoor(x, y, open)
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
		if m, ok := c.(*npc.Mount); c != nil && ok {
			m.AddRider(p)
			p.world.DeleteCreature(m)
			p.Move(action)
			p.mount = m
			p.Standup()
			message.Enqueue(fmt.Sprintf("You mount %s.", m.GetName().WithDefinite()))
			return true
		}
		message.PrintMessage("There is no creature to mount here.")

	}

	return false
}

func (p *Player) GetName() ui.Name {
	return &ui.PlainName{"You"}
}

func (p *Player) GetID() string {
	return "Player"
}

func (p *Player) GetAlignment() worldmap.Alignment {
	return worldmap.Player
}

func (p *Player) IsCrouching() bool {
	return p.crouching
}

func (p *Player) Standup() {
	p.crouching = false
}

func (p *Player) Crouch() {
	p.crouching = true
}

func (p *Player) ToggleCrouch() bool {
	p.crouching = !p.crouching
	if p.crouching {
		if p.mount != nil {
			message.PrintMessage("You can't crouch while riding.")
			p.Standup()
			return false
		} else {
			message.Enqueue("You crouch down.")
		}
	} else {
		message.Enqueue("You stand up.")
	}
	return true
}

func (p *Player) Talk() {
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			x := p.location.X + i
			y := p.location.Y + j
			if p.world.IsValid(x, y) {
				creature := p.world.GetCreature(x, y)
				switch c := creature.(type) {
				case *npc.Npc:
					{
						interaction := c.Talk()
						switch interaction {
						case npc.Trade:
							ui.GetInput()
							trade(p, c)
						case npc.Bounty:
							ui.GetInput()
							claimBounties(p, c)
						}
						return
					}
				case *npc.Mount:
					message.PrintMessage(fmt.Sprintf("You try to talk to %s. It doesn't seem to respond.", c.GetName().WithDefinite()))
					return
				case *npc.Enemy:
					message.PrintMessage(fmt.Sprintf("You try to talk to %s. They don't seem amused.", c.GetName().WithDefinite()))
					return
				}
			}
		}
	}
	message.PrintMessage("You talk to yourself.")

}
func (p *Player) Read() {
	items := p.world.GetItems(p.location.X, p.location.Y)

	readables := make([]*item.Item, 0)

	for _, itm := range items {
		if itm.HasComponent("readable") {
			readables = append(readables, itm)
		}
	}

	// Put the items back
	for i := len(items) - 1; i >= 0; i-- {
		p.world.PlaceItem(p.location.X, p.location.Y, items[i])
	}

	if len(readables) > 0 {
		for i, readable := range readables {
			message.PrintMessage(fmt.Sprintf("Would you like to read the %s on the ground? [yn]", readable.GetName()))
			selection := ui.GetInput()
			if selection == ui.Confirm {
				message.PrintMessage(readable.Component("readable").(item.ReadableComponent).Description)
				// if last item don't bother waiting for input
				if i != len(readables)-1 {
					ui.GetInput()
				}
			}
		}
		return
	}

	for {
		message.PrintMessage(fmt.Sprintf("What item do you want to read? [%s or ?*]", p.GetReadableKeys()))
		s, c := ui.GetItemSelection()

		switch s {
		case ui.All:
			p.PrintInventory()
			continue
		case ui.AllRelevant:
			p.PrintReadables()
			continue
		case ui.Cancel:
			message.PrintMessage("Never mind.")
			return
		case ui.SpecificItem:
			itm := p.GetItem(c)
			if itm == nil {
				message.PrintMessage("You don't have that to read.")
				ui.GetInput()
			} else {
				if itm.HasComponent("readable") {
					message.PrintMessage(itm.Component("readable").(item.ReadableComponent).Description)
					return
				} else {
					message.PrintMessage("That is not something that you can read.")
				}
			}

		}
	}
}

func (p *Player) ProcessEvent(e event.Event) {
	switch ev := e.(type) {
	case event.CrimeEvent:
		ev.Witness(p.world, p)
	}
}

func (p *Player) SetMap(world *worldmap.Map) {
	p.world = world
}

func (p *Player) Mount() *npc.Mount {
	return p.mount
}

func (p *Player) AddMount(m *npc.Mount) {
	p.mount = m
}

func (p *Player) GetVisionDistance() int {
	return 20
}

func (p *Player) Update() {

	for _, attribute := range p.attributes {
		attribute.Update()
	}

	// Apply armour AC bonus
	if p.armour != nil {
		p.attributes["ac"].AddEffect(item.NewEffect(p.armour.Component("armour").(item.ArmourComponent).Bonus, 1, true))
		p.attributes["ac"].AddEffect(item.NewEffect(p.armour.Component("armour").(item.ArmourComponent).Bonus, 1, false))
	}

	if p.mount != nil {
		p.mount.ResetMoved()
		p.mount.SetCoordinates(p.location.X, p.location.Y)
		if p.mount.IsDead() {
			p.mount = nil
		}
	}
}

func (p *Player) LoadMount(mounts []*npc.Mount) {
	for _, m := range mounts {
		if p.mountID == m.GetID() {
			m.AddRider(p)
			p.mount = m
		}
	}
}

type Player struct {
	location   worldmap.Coordinates
	icon       icon.Icon
	initiative int
	attributes map[string]*worldmap.Attribute
	crouching  bool
	money      int
	weapon     *item.Item
	armour     *item.Item
	inventory  map[rune]([]*item.Item)
	mountID    string
	mount      *npc.Mount
	world      *worldmap.Map
}
