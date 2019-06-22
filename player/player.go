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

	player := &Player{worldmap.Coordinates{0, 0}, icon.CreatePlayerIcon(), 1, attributes, []worldmap.Skill{}, false, 1000, item.WeaponComponent{0, item.NoAmmo, nil, item.NewDamage(2, 1, 0), item.Effects{}}, nil, nil, nil, make(map[rune]([]*item.Item)), "", nil, world}
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

	keys := []string{"Location", "Icon", "Initiative", "Attributes", "Skills", "Crouching", "Money", "Unarmed", "Primary", "Secondary", "Armour", "Inventory", "MountID"}

	mountID := ""
	if p.mount != nil {
		mountID = p.mount.GetID()
	}

	playerValues := map[string]interface{}{
		"Location":   p.location,
		"Icon":       p.icon,
		"Initiative": p.initiative,
		"Attributes": p.attributes,
		"Skills":     p.skills,
		"Money":      p.money,
		"Unarmed":    p.unarmed,
		"Primary":    p.primary,
		"Secondary":  p.secondary,
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
		Skills     []worldmap.Skill
		Crouching  bool
		Money      int
		Unarmed    item.WeaponComponent
		Primary    *item.Item
		Secondary  *item.Item
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
	p.skills = v.Skills
	p.crouching = v.Crouching
	p.money = v.Money
	p.unarmed = v.Unarmed
	p.primary = v.Primary
	p.secondary = v.Secondary
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
	if p.primary != nil {
		return p.primary.Component("weapon").(item.WeaponComponent)
	}

	if p.secondary != nil {
		return p.secondary.Component("weapon").(item.WeaponComponent)
	}

	return p.unarmed
}

func (p *Player) attack(c worldmap.Creature, weapon item.WeaponComponent, hitBonus, damageBonus int) {
	event.Emit(event.NewAttack(p, c))

	if c.AttackHits(rand.Intn(20) + hitBonus + 1) {
		message.Enqueue(fmt.Sprintf("You hit %s.", c.GetName().WithDefinite()))
		c.TakeDamage(weapon.Damage, weapon.Effects, damageBonus)
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
	var weapon item.WeaponComponent
	unarmed := false
	if p.primary != nil && p.secondary != nil {
		primary := p.primary.Component("weapon").(item.WeaponComponent)
		secondary := p.secondary.Component("weapon").(item.WeaponComponent)
		if !primary.Ranged() {
			if !secondary.Ranged() {
				message.PrintMessage(fmt.Sprintf("Which weapon would you like to use? %s[p] or %s[s]", p.primary.GetName(), p.secondary.GetName()))
				selection := ui.NoAction
				for selection == ui.NoAction || selection == ui.CancelAction {
					selection = ui.EquippedSelection()
				}
				switch selection {
				case ui.Primary:
					weapon = primary
				case ui.Secondary:
					weapon = secondary
				}
			} else {
				weapon = primary
			}
		} else {
			// Use unarmed if both weapons are ranged
			unarmed = true
		}
	} else if p.primary != nil && !p.primary.Component("weapon").(item.WeaponComponent).Ranged() {
		weapon = p.primary.Component("weapon").(item.WeaponComponent)
	} else if p.secondary != nil && !p.secondary.Component("weapon").(item.WeaponComponent).Ranged() {
		weapon = p.secondary.Component("weapon").(item.WeaponComponent)
	} else {
		unarmed = true
	}

	if unarmed {
		weapon = p.unarmed
	}

	proficiencyBonus := 0
	if unarmed {
		if p.hasSkill(worldmap.Unarmed) {
			proficiencyBonus = 2
		}
	} else if p.hasWeaponProficiency(weapon) {
		proficiencyBonus = 2
	}

	hitBonus := worldmap.GetBonus(p.attributes["str"].Value()) + proficiencyBonus
	damageBonus := worldmap.GetBonus(p.attributes["str"].Value()) + proficiencyBonus

	p.attack(c, weapon, hitBonus, damageBonus)
}

func (p *Player) RangedAttack() bool {

	weapons := make([]item.WeaponComponent, 0)

	if p.primary != nil {
		weapons = append(weapons, p.primary.Component("weapon").(item.WeaponComponent))
	}

	if p.secondary != nil {
		weapons = append(weapons, p.secondary.Component("weapon").(item.WeaponComponent))
	}

	weaponsRemaining := make([]item.WeaponComponent, 0)
	for _, w := range weapons {
		if w.Ranged() {
			weaponsRemaining = append(weaponsRemaining, w)
		}
	}
	weapons = weaponsRemaining

	if len(weapons) == 0 {
		message.PrintMessage("You are not wielding a ranged weapon.")
		return false
	}

	weaponsRemaining = make([]item.WeaponComponent, 0)
	for _, w := range weapons {
		if !w.IsUnloaded() {
			weaponsRemaining = append(weaponsRemaining, w)
		}
	}
	weapons = weaponsRemaining

	if len(weapons) == 0 {
		message.PrintMessage("You are not wielding a loaded weapon.")
		return false
	}

	choice := 0
	if len(weapons) == 2 {
		message.PrintMessage(fmt.Sprintf("Which weapon would you like to use? %s[p] or %s[s]", p.primary.GetName(), p.secondary.GetName()))

		selection := ui.NoAction
		for selection == ui.NoAction {
			selection = ui.EquippedSelection()
		}
		switch selection {
		case ui.Primary:
			choice = 0
		case ui.Secondary:
			choice = 1
		case ui.CancelAction:
			message.PrintMessage("Never mind.")
			return false
		}
	}

	target := p.findTarget()

	weapons[choice].Fire()
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

		proficiencyBonus := 0
		if p.hasWeaponProficiency(weapons[choice]) {
			proficiencyBonus = 2
		}

		p.attack(target, weapons[choice], worldmap.GetBonus(p.attributes["dex"].Value())+proficiencyBonus-coverPenalty, proficiencyBonus)
		// Attack again if player has double shot, weapon loaded and target not dead
		if p.hasSkill(worldmap.DoubleShot) && !weapons[choice].IsUnloaded() && !target.IsDead() {
			weapons[choice].Fire()
			p.attack(target, weapons[choice], worldmap.GetBonus(p.attributes["dex"].Value())+proficiencyBonus-coverPenalty, proficiencyBonus)
		}
	} else {
		message.Enqueue("Your target was too far away.")
	}

	return true
}

func (p *Player) TakeDamage(damage item.Damage, effects item.Effects, bonus int) {
	total_damage := damage.Damage() + bonus
	p.attributes["hp"].AddEffect(item.NewInstantEffect(-total_damage))
	p.applyEffects(effects)
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
	if p.primary != nil {
		equippedWeaponText := fmt.Sprintf("%s - %s", string(p.primary.GetKey()), p.primary.GetName())
		ui.WriteText(0, position, equippedWeaponText)
		position++
	}

	if p.secondary != nil {
		equippedWeaponText := fmt.Sprintf("%s - %s", string(p.secondary.GetKey()), p.secondary.GetName())
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

func (p *Player) PrintItemsByType(itemType string) {
	position := 0
	for k, items := range p.inventory {
		if !p.inventory[k][0].HasComponent(itemType) {
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

func (p *Player) getAmmo(weapon item.WeaponComponent) *item.Item {
	for k, items := range p.inventory {
		if items[0].HasComponent("ammo") && weapon.AmmoTypeMatches(items[0]) {
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

func (p *Player) KeysByType(itemType string) string {
	keysSet := make([]bool, 128)
	for k := range p.inventory {
		if p.inventory[k][0].HasComponent(itemType) {
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
	money := item.Money(p.money)
	if key == money.GetKey() {
		return money
	}

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
		message.PrintMessage(fmt.Sprintf("What item do you want to wield? [%s or ?*]", p.KeysByType("weapon")))
		s, c := ui.GetItemSelection()

		switch s {
		case ui.All:
			p.PrintInventory()
			continue
		case ui.AllRelevant:
			p.PrintItemsByType("weapon")
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
					var other *item.Item
					if p.primary != nil && p.secondary != nil {
						message.PrintMessage(fmt.Sprintf("Which weapon would you like to unwield? %s[p] or %s[s]", p.primary.GetName(), p.secondary.GetName()))
						selection := ui.NoAction
						for selection == ui.NoAction {
							selection = ui.EquippedSelection()
						}
						switch selection {
						case ui.Primary:
							other = p.primary
							p.primary = itm
						case ui.Secondary:
							other = p.secondary
							p.secondary = itm
						case ui.CancelAction:
							message.PrintMessage("Never mind.")
							return false
						}

					} else if p.primary != nil {
						other = p.secondary
						p.secondary = itm
					} else {
						other = p.primary
						p.primary = itm
					}

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
		message.PrintMessage(fmt.Sprintf("What item do you want to wear? [%s or ?*]", p.KeysByType("armour")))
		s, c := ui.GetItemSelection()

		switch s {
		case ui.All:
			p.PrintInventory()
			continue
		case ui.AllRelevant:
			p.PrintItemsByType("armour")
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
	weapons := make([]item.WeaponComponent, 0)

	if p.primary != nil {
		weapons = append(weapons, p.primary.Component("weapon").(item.WeaponComponent))
	}

	if p.secondary != nil {
		weapons = append(weapons, p.secondary.Component("weapon").(item.WeaponComponent))
	}

	weaponsRemaining := make([]item.WeaponComponent, 0)
	for _, w := range weapons {
		if w.Ranged() {
			weaponsRemaining = append(weaponsRemaining, w)
		}
	}
	weapons = weaponsRemaining

	if len(weapons) == 0 {
		message.PrintMessage("You are not wielding a ranged weapon.")
		return false
	}

	weaponsRemaining = make([]item.WeaponComponent, 0)
	for _, w := range weapons {
		if !w.IsFullyLoaded() {
			weaponsRemaining = append(weaponsRemaining, w)
		}
	}
	weapons = weaponsRemaining

	if len(weapons) == 0 {
		message.PrintMessage("The weapons that you are wielding are already fully loaded.")
		return false
	}

	weaponsRemaining = make([]item.WeaponComponent, 0)
	for _, w := range weapons {
		if p.hasAmmo(w) {
			weaponsRemaining = append(weaponsRemaining, w)
		}
	}
	weapons = weaponsRemaining

	if len(weapons) == 0 {
		message.PrintMessage("You don't have ammo for the weapons you are wielding.")
		return false
	}

	choice := 0
	if len(weapons) == 2 {
		message.PrintMessage(fmt.Sprintf("Which weapon would you like to load? %s[p] or %s[s]", p.primary.GetName(), p.secondary.GetName()))

		selection := ui.NoAction
		for selection == ui.NoAction {
			selection = ui.EquippedSelection()
		}
		switch selection {
		case ui.Primary:
			choice = 0
		case ui.Secondary:
			choice = 1
		case ui.CancelAction:
			message.PrintMessage("Never mind.")
			return false
		}
	}

	for !weapons[choice].IsFullyLoaded() && p.hasAmmo(weapons[choice]) {
		p.getAmmo(weapons[choice])
		weapons[choice].Load()
	}

	if weapons[choice].IsFullyLoaded() {
		message.Enqueue("You have fully loaded your weapon.")
	} else {
		message.Enqueue("You have loaded your weapon.")
	}
	return true
}

func (p *Player) ConsumeItem() bool {

	for {
		message.PrintMessage(fmt.Sprintf("What item do you want to eat? [%s or ?*]", p.KeysByType("consumable")))
		s, c := ui.GetItemSelection()

		switch s {
		case ui.All:
			p.PrintInventory()
			continue
		case ui.AllRelevant:
			p.PrintItemsByType("consumable")
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

// Check whether player has ammo for particular weapon
func (p *Player) hasAmmo(weapon item.WeaponComponent) bool {
	for _, items := range p.inventory {
		if items[0].HasComponent("ammo") && weapon.AmmoTypeMatches(items[0]) {
			return true
		}
	}
	return false
}

func (p *Player) applyEffects(effects item.Effects) {
	for attr, attribute := range p.attributes {
		for _, effect := range effects[attr] {
			eff := new(item.Effect)
			*eff = effect
			attribute.AddEffect(eff)
		}
	}
}

func (p *Player) consume(itm *item.Item) {
	originalHp := p.attributes["hp"].Value()
	originalHunger := p.attributes["hunger"].Value()
	originalThirst := p.attributes["thirst"].Value()

	p.applyEffects(itm.Component("consumable").(item.ConsumableComponent).Effects)

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
	if p.primary != nil {
		total += p.primary.GetWeight()
	}

	if p.secondary != nil {
		total += p.secondary.GetWeight()
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
		tooHeavy := items[k][0].GetWeight() > 10*float64(p.attributes["str"].Value())
		for _, item := range items[k] {
			if tooHeavy {
				message.Enqueue(fmt.Sprintf("You try to lift the %s but it is too heavy.", item.GetName()))
				p.world.PlaceItem(x, y, item)
			} else {
				p.AddItem(item)
			}
		}

		if tooHeavy {
			continue
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

	x, y, _ := p.SelectDirection()
	if p.location == (worldmap.Coordinates{x, y}) {
		return false
	}

	// If there is a door, toggle its position if it's not already there
	if p.world.IsDoor(x, y) {
		if p.world.Door(x, y).Open() != open {
			p.world.ToggleDoor(x, y, open)
			if open {
				if p.world.Door(x, y).Locked() {
					message.Enqueue("The door is locked.")
				} else {
					message.Enqueue("The door opens.")
				}
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

func (p *Player) SelectDirection() (int, int, ui.PlayerAction) {
	message.PrintMessage("Which direction?")
	height := p.world.GetHeight()
	width := p.world.GetWidth()
	x, y := p.GetCoordinates()
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
			return x, y, ui.Wait
		} else {
			message.PrintMessage("Invalid direction.")
			validMove = false
		}

		if validMove {
			break
		}
	}
	return x, y, action
}

func (p *Player) ToggleMount() bool {
	mounted := p.mount != nil

	x, y, action := p.SelectDirection()
	if p.location == (worldmap.Coordinates{x, y}) {
		return false
	}

	if mounted {
		if p.world.IsPassable(x, y) {
			p.mount.RemoveRider()
			p.mount = nil
			p.Move(action)
			message.Enqueue("You dismount.")
			p.location = worldmap.Coordinates{x, y}
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
						case npc.DoesNotSpeak:
							message.PrintMessage(fmt.Sprintf("You try to talk to %s. It doesn't seem to respond.", c.GetName().WithDefinite()))
						}
						return
					}
				case *npc.Mount:
					message.PrintMessage(fmt.Sprintf("You try to talk to %s. It doesn't seem to respond.", c.GetName().WithDefinite()))
					return
				}
			}
		}
	}
	message.PrintMessage("You talk to yourself.")
}

func (p *Player) Pickpocket() bool {
	x, y, _ := p.SelectDirection()
	if p.location == (worldmap.Coordinates{x, y}) {
		return false
	}

	c := p.world.GetCreature(x, y)
	if c == nil {
		message.Enqueue("There is no one there to pickpocket.")
		return true
	}
	if n, ok := c.(*npc.Npc); ok && n.Human() {
		pickpocket(p, n)
	} else {
		message.Enqueue("You can't pickpocket the creature there")
	}
	return true
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
		message.PrintMessage(fmt.Sprintf("What item do you want to read? [%s or ?*]", p.KeysByType("readable")))
		s, c := ui.GetItemSelection()

		switch s {
		case ui.All:
			p.PrintInventory()
			continue
		case ui.AllRelevant:
			p.PrintItemsByType("readable")
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
					p.AddItem(itm)
					return
				} else {
					message.PrintMessage("That is not something that you can read.")
					p.AddItem(itm)
				}
			}

		}
	}
}

func (p *Player) Use() bool {
	for {
		message.PrintMessage(fmt.Sprintf("What do you want to use or apply? [%s or ?*]", p.KeysByType("usable")))
		s, c := ui.GetItemSelection()

		switch s {
		case ui.All:
			p.PrintInventory()
			continue
		case ui.AllRelevant:
			p.PrintItemsByType("usable")
			continue
		case ui.Cancel:
			message.PrintMessage("Never mind.")
			return false
		case ui.SpecificItem:
			itm := p.GetItem(c)
			if itm == nil {
				message.PrintMessage("You don't have that.")
				ui.GetInput()
			} else {
				if itm.HasComponent("usable") {
					if itm.HasComponent("key") {
						lockpickingBonus := 0.0
						if p.hasSkill(worldmap.Lockpicking) {
							lockpickingBonus = 0.2
						}

						// Keys are multiple use
						p.AddItem(itm)
						x, y, _ := p.SelectDirection()
						if p.location == (worldmap.Coordinates{x, y}) {
							return false
						}

						if !p.world.IsDoor(x, y) {
							message.Enqueue("You see no door here.")
						} else {
							door := p.world.Door(x, y)
							// Can have multiple keys that unlock different doors
							allKeys := p.inventory[c]
							anyFit := false
							for _, key := range allKeys {
								anyFit = anyFit || door.KeyFits(key)
							}

							if anyFit {
								if itm.Component("key").(item.KeyComponent).Works(lockpickingBonus) {
									door.ToggleLocked()
									if door.Locked() {
										message.Enqueue("You lock the door.")
									} else {
										message.Enqueue("You unlock the door.")
									}
								} else {
									message.Enqueue(fmt.Sprintf("The %s didn't work.", itm.GetName()))
								}
							} else {
								message.Enqueue("This does not work for this door.")
							}
						}
						itm = p.GetItem(c)
					}
					name := itm.GetName()
					if itm.TryBreaking() {
						message.Enqueue(fmt.Sprintf("The %s broke.", name))
					}
					p.AddItem(itm)

					return true
				} else {
					message.PrintMessage("You can't use that.")
					p.AddItem(itm)
					ui.GetInput()
					return false
				}
			}

		}

	}
}

func (p *Player) hasSkill(skill worldmap.Skill) bool {
	for _, s := range p.skills {
		if s == skill {
			return true
		}
	}
	return false
}

func (p *Player) hasWeaponProficiency(weapon item.WeaponComponent) bool {

	var skill worldmap.Skill
	switch p.Weapon().Type {
	case item.NoAmmo:
		skill = worldmap.Melee
	case item.Bow:
		skill = worldmap.Archery
	case item.Shotgun:
		skill = worldmap.Shotguns
	case item.Rifle:
		skill = worldmap.Rifles
	case item.Pistol:
		skill = worldmap.Pistols
	}
	return p.hasSkill(skill)
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
	skills     []worldmap.Skill
	crouching  bool
	money      int
	unarmed    item.WeaponComponent
	primary    *item.Item
	secondary  *item.Item
	armour     *item.Item
	inventory  map[rune]([]*item.Item)
	mountID    string
	mount      *npc.Mount
	world      *worldmap.Map
}
