package npc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"

	"github.com/onorton/cowboysindians/event"
	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/ui"
	"github.com/onorton/cowboysindians/worldmap"
	"github.com/rs/xid"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type NpcAttributes struct {
	Icon          icon.Icon
	Initiative    int
	Hp            int
	Ac            int
	Str           int
	Dex           int
	Encumbrance   int
	Money         int
	Unarmed       item.WeaponComponent
	Inventory     [][]item.ItemChoice
	ShopInventory map[string]int
	DialogueType  *dialogueType
	AiType        string
	Mount         map[string]float64
	Protector     map[string]float64
	Probability   float64
	Human         bool
}

var npcData map[string]NpcAttributes = fetchNpcData()

func fetchNpcData() map[string]NpcAttributes {
	data, err := ioutil.ReadFile("data/npc.json")
	check(err)
	var eD map[string]NpcAttributes
	err = json.Unmarshal(data, &eD)
	check(err)
	return eD
}

func generateName(npcType string, human bool) ui.Name {
	if !human {
		return &ui.PlainName{npcType}
	}
	firstName := Names.FirstNames[rand.Intn(len(Names.FirstNames))]
	lastName := Names.LastNames[rand.Intn(len(Names.LastNames))]
	switch npcType {
	case "sheriff":
		return &npcName{fmt.Sprintf("Sheriff %s", lastName), npcType, false}
	case "deputy":
		return &npcName{fmt.Sprintf("Deputy %s", lastName), npcType, false}
	}
	return &npcName{firstName + " " + lastName, npcType, false}
}

func RandomNpcType() string {
	probabilities := map[string]float64{}
	for npcType, npcInfo := range npcData {
		probabilities[npcType] = npcInfo.Probability
	}
	max := 0.0

	for _, probability := range probabilities {
		if probability > 0 {
			inverse := 1.0 / probability
			if inverse > max {
				max = inverse
			}
		}
	}
	possibleNpcs := make([]string, 0)

	for name, probability := range probabilities {
		count := int(probability * max)
		for i := 0; i < count; i++ {
			possibleNpcs = append(possibleNpcs, name)
		}
	}

	n := rand.Intn(len(possibleNpcs))
	return possibleNpcs[n]
}

func NewNpc(npcType string, x, y int, world *worldmap.Map, protectee *string) *Npc {
	n := npcData[npcType]
	id := xid.New().String()
	dialogue := newDialogue(n.DialogueType, world, nil, nil)
	location := worldmap.Coordinates{x, y}
	ai := newAi(n.AiType, world, location, nil, nil, dialogue, protectee)

	attributes := map[string]*worldmap.Attribute{
		"hp":          worldmap.NewAttribute(n.Hp, n.Hp),
		"ac":          worldmap.NewAttribute(n.Ac, n.Ac),
		"str":         worldmap.NewAttribute(n.Str, n.Str),
		"dex":         worldmap.NewAttribute(n.Dex, n.Dex),
		"encumbrance": worldmap.NewAttribute(n.Encumbrance, n.Encumbrance)}
	npc := &Npc{generateName(npcType, n.Human), id, worldmap.Coordinates{x, y}, n.Icon, n.Initiative, attributes, false, n.Money, n.Unarmed, nil, nil, make([]*item.Item, 0), "", generateMount(n.Mount, x, y), world, ai, dialogue, n.Human}
	for _, itm := range generateInventory(n.Inventory) {
		npc.PickupItem(itm)
	}
	event.Subscribe(npc)
	return npc
}

func AddProtector(npcType string, x, y int, protectee string) *Npc {
	probabilities := npcData[npcType].Protector
	if probabilities == nil {
		return nil
	}

	protectorType := "None"

	max := 0.0
	for _, probability := range probabilities {
		if probability > 0 {
			inverse := 1.0 / probability
			if inverse > max {
				max = inverse
			}
		}
	}

	protectors := make([]string, 0)
	for name, probability := range probabilities {
		count := int(probability * max)
		for i := 0; i < count; i++ {
			protectors = append(protectors, name)
		}
	}

	protectorType = protectors[rand.Intn(len(protectors))]
	if protectorType != "None" {
		return NewNpc(protectorType, x, y, nil, &protectee)
	}
	return nil
}

func NewShopkeeper(npcType string, x, y int, world *worldmap.Map, t worldmap.Town, b worldmap.Building) *Npc {
	n := npcData[npcType]
	id := xid.New().String()
	dialogue := newDialogue(n.DialogueType, world, &t, &b)
	location := worldmap.Coordinates{x, y}
	ai := newAi(n.AiType, world, location, &t, &b, dialogue, nil)

	attributes := map[string]*worldmap.Attribute{
		"hp":          worldmap.NewAttribute(n.Hp, n.Hp),
		"ac":          worldmap.NewAttribute(n.Ac, n.Ac),
		"str":         worldmap.NewAttribute(n.Str, n.Str),
		"dex":         worldmap.NewAttribute(n.Dex, n.Dex),
		"encumbrance": worldmap.NewAttribute(n.Encumbrance, n.Encumbrance)}

	npc := &Npc{generateName(npcType, n.Human), id, worldmap.Coordinates{x, y}, n.Icon, n.Initiative, attributes, false, n.Money, n.Unarmed, nil, nil, make([]*item.Item, 0), "", generateMount(n.Mount, x, y), world, ai, dialogue, n.Human}
	for c, count := range n.ShopInventory {

		for i := 0; i < count; i++ {
			switch c {
			case "Ammo":
				npc.PickupItem(item.GenerateAmmo())
			case "Armour":
				npc.PickupItem(item.GenerateArmour())
			case "Consumable":
				npc.PickupItem(item.GenerateConsumable())
			case "Item":
				npc.PickupItem(item.GenerateItem())
			case "Weapon":
				npc.PickupItem(item.GenerateWeapon())
			}
		}
	}

	for _, itm := range generateInventory(n.Inventory) {
		npc.PickupItem(itm)
	}
	event.Subscribe(npc)
	return npc
}

func chooseItems(probabilites []float64) int {
	max := 0.0

	for _, probability := range probabilites {
		if probability > 0 {
			inverse := 1.0 / probability
			if inverse > max {
				max = inverse
			}
		}
	}
	choices := make([]int, 0)

	for index, probability := range probabilites {
		count := int(probability * max)
		for i := 0; i < count; i++ {
			choices = append(choices, index)
		}
	}

	n := rand.Intn(len(choices))
	return choices[n]
}

func generateInventory(itemChoices [][]item.ItemChoice) []*item.Item {

	inventory := make([]*item.Item, 0)
	for _, selection := range itemChoices {
		choices := make([]float64, 0)
		for _, choice := range selection {
			choices = append(choices, choice.Probability)
		}

		choice := chooseItems(choices)
		for itemType, count := range selection[choice].Items {
			for i := 0; i < count; i++ {
				inventory = append(inventory, item.NewItem(itemType))
			}
		}
	}
	return inventory
}

func (npc *Npc) Render() ui.Element {
	if npc.mount != nil {
		return icon.MergeIcons(npc.icon, npc.mount.GetIcon())
	}
	return npc.icon.Render()
}

func (npc *Npc) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	keys := []string{"Name", "Id", "Location", "Icon", "Initiative", "Attributes", "Crouching", "Money", "Unarmed", "Weapon", "Armour", "Inventory", "MountID", "Ai", "Dialogue", "Human"}

	mountID := ""
	if npc.mount != nil {
		mountID = npc.mount.GetID()
	}

	npcValues := map[string]interface{}{
		"Name":       npc.name,
		"Id":         npc.id,
		"Location":   npc.location,
		"Icon":       npc.icon,
		"Initiative": npc.initiative,
		"Attributes": npc.attributes,
		"Crouching":  npc.crouching,
		"Money":      npc.money,
		"Unarmed":    npc.unarmed,
		"Weapon":     npc.weapon,
		"Armour":     npc.armour,
		"Inventory":  npc.inventory,
		"MountID":    mountID,
		"Ai":         npc.ai,
		"Dialogue":   npc.dialogue,
		"Human":      npc.human,
	}

	length := len(npcValues)
	count := 0

	for _, key := range keys {
		jsonValue, err := json.Marshal(npcValues[key])

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

func (npc *Npc) Talk() interaction {
	if npc.dialogue == nil {
		return DoesNotSpeak
	}
	npc.name.PlayerKnows()
	return npc.dialogue.interact()
}

func (npc *Npc) UnmarshalJSON(data []byte) error {

	type npcJson struct {
		Name       map[string]interface{}
		Id         string
		Location   worldmap.Coordinates
		Icon       icon.Icon
		Initiative int
		Attributes map[string]*worldmap.Attribute
		Crouching  bool
		Money      int
		Unarmed    item.WeaponComponent
		Weapon     *item.Item
		Armour     *item.Item
		Inventory  []*item.Item
		MountID    string
		Ai         map[string]interface{}
		Dialogue   map[string]interface{}
		Human      bool
	}
	var v npcJson

	json.Unmarshal(data, &v)

	npc.name = unmarshalName(v.Name)
	npc.id = v.Id
	npc.location = v.Location
	npc.icon = v.Icon
	npc.initiative = v.Initiative
	npc.attributes = v.Attributes
	npc.crouching = v.Crouching
	npc.money = v.Money
	npc.unarmed = v.Unarmed
	npc.weapon = v.Weapon
	npc.armour = v.Armour
	npc.inventory = v.Inventory
	npc.mountID = v.MountID
	npc.ai = unmarshalAi(v.Ai)
	npc.dialogue = unmarshalDialogue(v.Dialogue)
	npc.human = v.Human

	event.Subscribe(npc)

	return nil
}

func (npc *Npc) GetCoordinates() (int, int) {
	return npc.location.X, npc.location.Y
}
func (npc *Npc) SetCoordinates(x int, y int) {
	npc.location = worldmap.Coordinates{x, y}
}

func (npc *Npc) GetInitiative() int {
	return npc.initiative
}

func (npc *Npc) MeleeAttack(c worldmap.Creature) {
	npc.attack(c, worldmap.GetBonus(npc.attributes["str"].Value()), worldmap.GetBonus(npc.attributes["str"].Value()))
}

func (npc *Npc) rangedAttack(c worldmap.Creature, environmentBonus int) {
	npc.attack(c, worldmap.GetBonus(npc.attributes["dex"].Value())+environmentBonus, 0)
}

func (npc *Npc) attack(c worldmap.Creature, hitBonus, damageBonus int) {
	event.Emit(event.NewAttack(npc, c))

	hits := c.AttackHits(rand.Intn(20) + hitBonus + 1)
	if hits {
		c.TakeDamage(npc.Weapon().Damage, npc.Weapon().Effects, damageBonus)
		// If non-enemy dead, send murder event
		if c.IsDead() && c.GetAlignment() == worldmap.Neutral {
			event.Emit(event.NewMurder(npc, c, npc.location))
		}

	}
	if c.GetAlignment() == worldmap.Player {
		if hits {
			message.Enqueue(fmt.Sprintf("%s hit you.", npc.name.WithDefinite()))
		} else {
			message.Enqueue(fmt.Sprintf("%s missed you.", npc.name.WithDefinite()))
		}
	}

}

func (npc *Npc) AttackHits(roll int) bool {
	return roll > npc.attributes["ac"].Value()
}

func (npc *Npc) TakeDamage(damage item.Damage, effects item.Effects, bonus int) {
	total_damage := damage.Damage() + bonus
	npc.attributes["hp"].AddEffect(item.NewInstantEffect(-total_damage))
	npc.applyEffects(effects)
}

func (npc *Npc) IsDead() bool {
	return npc.attributes["hp"].Value() == 0
}

func (npc *Npc) wieldItem() bool {
	changed := false
	for i, itm := range npc.inventory {
		if itm.HasComponent("weapon") {
			if npc.weapon == nil {
				npc.weapon = itm
				npc.inventory = append(npc.inventory[:i], npc.inventory[i+1:]...)
				changed = true

			} else if itm.Component("weapon").(item.WeaponComponent).MaxDamage() > npc.Weapon().MaxDamage() {
				npc.inventory[i] = npc.weapon
				npc.weapon = itm
				changed = true
			}

		}

	}
	return changed
}

func (npc *Npc) wearArmour() bool {
	changed := false
	for i, itm := range npc.inventory {
		if itm.HasComponent("armour") {
			if npc.armour == nil {
				npc.armour = itm
				npc.inventory = append(npc.inventory[:i], npc.inventory[i+1:]...)
				changed = true

			} else if itm.Component("armour").(item.ArmourComponent).Bonus > npc.armour.Component("armour").(item.ArmourComponent).Bonus {
				npc.inventory[i] = npc.weapon
				npc.armour = itm
				changed = true
			}

		}

	}
	return changed
}

func (npc *Npc) overEncumbered() bool {
	weight := 0.0
	for _, item := range npc.inventory {
		weight += item.GetWeight()
	}
	return weight > float64(npc.attributes["encumbrance"].Value())
}
func (npc *Npc) dropItem(item *item.Item) {
	npc.RemoveItem(item)
	npc.world.PlaceItem(npc.location.X, npc.location.Y, item)
	if npc.world.IsVisible(npc.world.GetPlayer(), npc.location.X, npc.location.Y) {
		message.Enqueue(fmt.Sprintf("%s dropped a %s.", npc.name.WithDefinite(), item.GetName()))
	}

}

func (npc *Npc) Update() {
	for _, attribute := range npc.attributes {
		attribute.Update()
	}

	// Apply armour AC bonus
	if npc.armour != nil {
		npc.attributes["ac"].AddEffect(item.NewEffect(npc.armour.Component("armour").(item.ArmourComponent).Bonus, 1, true))
		npc.attributes["ac"].AddEffect(item.NewEffect(npc.armour.Component("armour").(item.ArmourComponent).Bonus, 1, false))
	}

	if npc.IsDead() {
		return
	}

	p := npc.world.GetPlayer()
	pX, pY := p.GetCoordinates()
	if npc.world.InConversationRange(npc, p) && npc.dialogue != nil {
		npc.dialogue.initialGreeting()
	} else if npc.world.IsVisible(npc, pX, pY) && npc.dialogue != nil {
		npc.dialogue.resetSeen()
	}
	action := npc.ai.update(npc, npc.world)
	action.execute()

	if _, ok := action.(MountedMoveAction); ok {
		// Gets another action
		npc.ai.update(npc, npc.world).execute()
	}

	if npc.mount != nil {
		npc.mount.ResetMoved()
		if npc.mount.IsDead() {
			npc.mount = nil
		}
	}
}

func (npc *Npc) EmptyInventory() {
	// First drop the corpse
	npc.world.PlaceItem(npc.location.X, npc.location.Y, item.NewCorpse("head", npc.id, npc.name.String(), npc.icon))
	npc.world.PlaceItem(npc.location.X, npc.location.Y, item.NewCorpse("body", npc.id, npc.name.String(), npc.icon))

	itemTypes := make(map[string]int)
	for _, item := range npc.inventory {
		npc.world.PlaceItem(npc.location.X, npc.location.Y, item)
		itemTypes[item.GetName()]++
	}

	if npc.weapon != nil {
		npc.world.PlaceItem(npc.location.X, npc.location.Y, npc.weapon)
		itemTypes[npc.weapon.GetName()]++
		npc.weapon = nil
	}
	if npc.armour != nil {
		npc.world.PlaceItem(npc.location.X, npc.location.Y, npc.armour)
		itemTypes[npc.armour.GetName()]++
		npc.armour = nil
	}

	if npc.money > 0 {
		npc.world.PlaceItem(npc.location.X, npc.location.Y, item.Money(npc.money))
		if npc.world.IsVisible(npc.world.GetPlayer(), npc.location.X, npc.location.Y) {
			message.Enqueue(fmt.Sprintf("%s dropped some money.", npc.name.WithDefinite()))
		}
	}

	if npc.world.IsVisible(npc.world.GetPlayer(), npc.location.X, npc.location.Y) {
		for name, count := range itemTypes {
			if count == 1 {
				message.Enqueue(fmt.Sprintf("%s dropped 1 %s.", npc.name.WithDefinite(), name))
			} else {
				message.Enqueue(fmt.Sprintf("%s dropped %d %ss.", npc.name.WithDefinite(), count, name))
			}
		}
	}

}

func (npc *Npc) getAmmo() *item.Item {
	for i, itm := range npc.inventory {
		if itm.HasComponent("ammo") && npc.Weapon().AmmoTypeMatches(itm) {
			npc.inventory = append(npc.inventory[:i], npc.inventory[i+1:]...)
			return itm
		}
	}
	return nil
}

func (npc *Npc) PickupItem(item *item.Item) {
	npc.inventory = append(npc.inventory, item)
	// If item had previous owner, send theft event
	if !item.Owned(npc.id) {
		event.Emit(event.NewTheft(npc, item, npc.location))
	}
	item.TransferOwner(npc.id)
}

func (npc *Npc) Inventory() []*item.Item {
	return npc.inventory
}

func (npc *Npc) ranged() bool {
	return npc.Weapon().Range > 0
}

func (npc *Npc) Weapon() item.WeaponComponent {
	if npc.weapon != nil {
		return npc.weapon.Component("weapon").(item.WeaponComponent)
	}
	return npc.unarmed
}

// Check whether npc is carrying a fully loaded weapon
func (npc *Npc) weaponFullyLoaded() bool {
	return npc.Weapon().IsFullyLoaded()
}

// Check whether npc has ammo for particular wielded weapon
func (npc *Npc) hasAmmo() bool {
	for _, itm := range npc.inventory {
		if itm.HasComponent("ammo") && npc.Weapon().AmmoTypeMatches(itm) {
			return true
		}
	}
	return false
}

func (npc *Npc) weaponLoaded() bool {
	if npc.weapon != nil && npc.Weapon().NeedsAmmo() {
		return !npc.Weapon().IsUnloaded()
	}
	return true

}

func (npc *Npc) applyEffects(effects item.Effects) {
	for attr, attribute := range npc.attributes {
		for _, effect := range effects[attr] {
			attribute.AddEffect(&effect)
		}
	}
}

func (npc *Npc) consume(itm *item.Item) {
	npc.applyEffects(itm.Component("consumable").(item.ConsumableComponent).Effects)
}

func (npc *Npc) bloodied() bool {
	return npc.attributes["hp"].Value() <= npc.attributes["hp"].Maximum()/2
}

func (npc *Npc) GetName() ui.Name {
	return npc.name
}

func (npc *Npc) GetAlignment() worldmap.Alignment {
	return worldmap.Neutral
}

func (npc *Npc) IsCrouching() bool {
	return npc.crouching
}

func (npc *Npc) Standup() {
	npc.crouching = false
}

func (npc *Npc) Crouch() {
	npc.crouching = true
}

func (npc *Npc) SetMap(world *worldmap.Map) {
	npc.world = world

	switch ai := npc.ai.(type) {
	case animalAi:
		ai.setMap(world)
	case aggAnimalAi:
		ai.setMap(world)
	case npcAi:
		ai.setMap(world)
	case barPatronAi:
		ai.setMap(world)
	}

	switch d := npc.dialogue.(type) {
	case *shopkeeperDialogue:
		d.setMap(world)
	case *sheriffDialogue:
		d.setMap(world)
	}

}

func (npc *Npc) Mount() *Mount {
	return npc.mount
}

func (npc *Npc) AddMount(m *Mount) {
	npc.mount = m
	npc.Standup()
}

func (npc *Npc) GetVisionDistance() int {
	return 20
}

func (npc *Npc) GetItems(addMoney bool) map[rune]([]*item.Item) {
	items := make(map[rune]([]*item.Item))
	for _, itm := range npc.inventory {
		existing := items[itm.GetKey()]
		if existing == nil {
			existing = make([]*item.Item, 0)
		}
		existing = append(existing, itm)
		items[itm.GetKey()] = existing
	}
	if addMoney && npc.money > 0 {
		money := item.Money(npc.money)
		items[money.GetKey()] = []*item.Item{money}
	}

	return items
}

func (npc *Npc) RemoveItem(itm *item.Item) {
	for i, item := range npc.inventory {
		if itm.GetName() == item.GetName() {
			npc.inventory = append(npc.inventory[:i], npc.inventory[i+1:]...)
			return
		}
	}
}

func (npc *Npc) LoadMount(mounts []*Mount) {
	for _, m := range mounts {
		if npc.mountID == m.GetID() {
			m.AddRider(npc)
			npc.mount = m
		}
	}
}

func (npc Npc) CanBuy(itm *item.Item) bool {
	return itm.GetValue() <= npc.money
}

func (npc *Npc) AddMoney(amount int) {
	npc.money += amount
}

func (npc *Npc) RemoveMoney(amount int) {
	npc.money -= amount
}

func (npc *Npc) Human() bool {
	return npc.human
}
func (npc *Npc) GetID() string {
	return npc.id
}

func (npc *Npc) GetBounties() *Bounties {
	if ai, ok := npc.ai.(*sheriffAi); ok {
		return ai.bounties
	}
	return &Bounties{}
}

func (npc *Npc) ProcessEvent(e event.Event) {
	if ev, ok := e.(event.CrimeEvent); ok && npc.Human() {
		ev.Witness(npc.world, npc)
	}
}

type Npc struct {
	name       ui.Name
	id         string
	location   worldmap.Coordinates
	icon       icon.Icon
	initiative int
	attributes map[string]*worldmap.Attribute
	crouching  bool
	money      int
	unarmed    item.WeaponComponent
	weapon     *item.Item
	armour     *item.Item
	inventory  []*item.Item
	mountID    string
	mount      *Mount
	world      *worldmap.Map
	ai         ai
	dialogue   dialogue
	human      bool
}
