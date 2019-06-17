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

type EnemyAttributes struct {
	Icon         icon.Icon
	Initiative   int
	Hp           int
	Ac           int
	Str          int
	Dex          int
	Encumbrance  int
	Money        int
	Unarmed      item.WeaponComponent
	DialogueType *dialogueType
	AiType       string
	Inventory    [][]item.ItemChoice
	Mount        map[string]float64
	Probability  float64
	Human        bool
}

var enemyData map[string]EnemyAttributes = fetchEnemyData()

func fetchEnemyData() map[string]EnemyAttributes {
	data, err := ioutil.ReadFile("data/enemy.json")
	check(err)
	var eD map[string]EnemyAttributes
	err = json.Unmarshal(data, &eD)
	check(err)
	return eD
}

func RandomEnemyType() string {
	probabilities := map[string]float64{}
	for enemyType, enemyInfo := range enemyData {
		probabilities[enemyType] = enemyInfo.Probability
	}

	return chooseType(probabilities)
}

func NewEnemy(enemyType string, x, y int, world *worldmap.Map) *Enemy {
	enemy := enemyData[enemyType]
	id := xid.New().String()
	dialogue := newDialogue(enemy.DialogueType, world, nil, nil)
	ai := newAi(enemy.AiType, world, worldmap.Coordinates{x, y}, nil, nil, dialogue, nil)
	attributes := map[string]*worldmap.Attribute{
		"hp":          worldmap.NewAttribute(enemy.Hp, enemy.Hp),
		"ac":          worldmap.NewAttribute(enemy.Ac, enemy.Ac),
		"str":         worldmap.NewAttribute(enemy.Str, enemy.Str),
		"dex":         worldmap.NewAttribute(enemy.Dex, enemy.Dex),
		"encumbrance": worldmap.NewAttribute(enemy.Encumbrance, enemy.Encumbrance)}
	name := generateName(enemyType, enemy.Human)
	e := &Enemy{name, id, worldmap.Coordinates{x, y}, enemy.Icon, enemy.Initiative, attributes, worldmap.Enemy, false, enemy.Money, enemy.Unarmed, nil, nil, make([]*item.Item, 0), "", generateMount(enemy.Mount, x, y), world, ai, dialogue, enemy.Human}
	for _, itm := range generateInventory(enemy.Inventory) {
		e.PickupItem(itm)
	}
	return e
}

func (e *Enemy) Render() ui.Element {
	if e.mount != nil {
		return icon.MergeIcons(e.icon, e.mount.GetIcon())
	}
	return e.icon.Render()
}

func (e *Enemy) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	keys := []string{"Name", "Id", "Location", "Icon", "Initiative", "Attributes", "Alignment", "Crouching", "Money", "Unarmed", "Weapon", "Armour", "Inventory", "Ai", "MountID", "Dialogue", "Human"}

	mountID := ""
	if e.mount != nil {
		mountID = e.mount.GetID()
	}

	enemyValues := map[string]interface{}{
		"Name":       e.name,
		"Id":         e.id,
		"Location":   e.location,
		"Icon":       e.icon,
		"Initiative": e.initiative,
		"Attributes": e.attributes,
		"Alignment":  e.alignment,
		"Crouching":  e.crouching,
		"Money":      e.money,
		"Unarmed":    e.unarmed,
		"Weapon":     e.weapon,
		"Armour":     e.armour,
		"Inventory":  e.inventory,
		"Ai":         e.ai,
		"MountID":    mountID,
		"Dialogue":   e.dialogue,
		"Human":      e.human,
	}

	length := len(enemyValues)
	count := 0

	for _, key := range keys {
		jsonValue, err := json.Marshal(enemyValues[key])
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

func (e *Enemy) UnmarshalJSON(data []byte) error {

	type enemyJson struct {
		Name       map[string]interface{}
		Id         string
		Location   worldmap.Coordinates
		Icon       icon.Icon
		Initiative int
		Attributes map[string]*worldmap.Attribute
		Alignment  worldmap.Alignment
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
	var v enemyJson

	json.Unmarshal(data, &v)

	e.name = unmarshalName(v.Name)
	e.id = v.Id
	e.location = v.Location
	e.icon = v.Icon
	e.initiative = v.Initiative
	e.attributes = v.Attributes
	e.alignment = v.Alignment
	e.crouching = v.Crouching
	e.money = v.Money
	e.unarmed = v.Unarmed
	e.weapon = v.Weapon
	e.armour = v.Armour
	e.inventory = v.Inventory
	e.mountID = v.MountID
	e.ai = unmarshalAi(v.Ai)
	e.dialogue = unmarshalDialogue(v.Dialogue)
	e.human = v.Human

	return nil
}

func (e *Enemy) GetCoordinates() (int, int) {
	return e.location.X, e.location.Y
}
func (e *Enemy) SetCoordinates(x int, y int) {
	e.location = worldmap.Coordinates{x, y}
}

func (e *Enemy) GetInitiative() int {
	return e.initiative
}

func (e *Enemy) MeleeAttack(c worldmap.Creature) {
	e.attack(c, worldmap.GetBonus(e.attributes["str"].Value()), worldmap.GetBonus(e.attributes["str"].Value()))
}

func (e *Enemy) rangedAttack(c worldmap.Creature, environmentBonus int) {
	e.attack(c, worldmap.GetBonus(e.attributes["dex"].Value())+environmentBonus, 0)
}

func (e *Enemy) attack(c worldmap.Creature, hitBonus, damageBonus int) {
	event.Emit(event.NewAttack(e, c))

	hits := c.AttackHits(rand.Intn(20) + hitBonus + 1)
	if hits {
		c.TakeDamage(e.Weapon().Damage, e.Weapon().Effects, damageBonus)
		// If non-enemy dead, send murder event
		if c.IsDead() && c.GetAlignment() == worldmap.Neutral {
			event.Emit(event.NewMurder(e, c, e.location))
		}
	}

	if c.GetAlignment() == worldmap.Player {
		if hits {
			message.Enqueue(fmt.Sprintf("%s hit you.", e.name.WithDefinite()))
		} else {
			message.Enqueue(fmt.Sprintf("%s missed you.", e.name.WithDefinite()))
		}
	}

}

func (e *Enemy) AttackHits(roll int) bool {
	return roll > e.attributes["ac"].Value()
}

func (e *Enemy) TakeDamage(damage item.Damage, effects item.Effects, bonus int) {
	total_damage := damage.Damage() + bonus
	e.attributes["hp"].AddEffect(item.NewInstantEffect(-total_damage))
	e.applyEffects(effects)
}

func (e *Enemy) IsDead() bool {
	return e.attributes["hp"].Value() == 0
}

func (e *Enemy) wieldItem() bool {
	changed := false
	for i, itm := range e.inventory {
		if itm.HasComponent("weapon") {
			if e.weapon == nil {
				e.weapon = itm
				e.inventory = append(e.inventory[:i], e.inventory[i+1:]...)
				changed = true

			} else if itm.Component("weapon").(item.WeaponComponent).MaxDamage() > e.Weapon().MaxDamage() {
				e.inventory[i] = e.weapon
				e.weapon = itm
				changed = true
			}

		}

	}
	return changed
}

func (e *Enemy) wearArmour() bool {
	changed := false
	for i, itm := range e.inventory {
		if itm.HasComponent("armour") {
			if e.armour == nil {
				e.armour = itm
				e.inventory = append(e.inventory[:i], e.inventory[i+1:]...)
				changed = true

			} else if itm.Component("armour").(item.ArmourComponent).Bonus > e.armour.Component("armour").(item.ArmourComponent).Bonus {
				e.inventory[i] = e.weapon
				e.armour = itm
				changed = true
			}

		}

	}
	return changed
}

func (e *Enemy) overEncumbered() bool {
	weight := 0.0
	for _, item := range e.inventory {
		weight += item.GetWeight()
	}
	return weight > float64(e.attributes["encumbrance"].Value())
}

func (e *Enemy) maximumLift() float64 {
	return 10 * float64(e.attributes["str"].Value())
}

func (e *Enemy) dropItem(item *item.Item) {
	e.RemoveItem(item)
	e.world.PlaceItem(e.location.X, e.location.Y, item)
	if e.world.IsVisible(e.world.GetPlayer(), e.location.X, e.location.Y) {
		message.Enqueue(fmt.Sprintf("%s dropped a %s.", e.name.WithDefinite(), item.GetName()))
	}

}

func (e *Enemy) Update() {

	for _, attribute := range e.attributes {
		attribute.Update()
	}

	// Apply armour AC bonus
	if e.armour != nil {
		e.attributes["ac"].AddEffect(item.NewEffect(e.armour.Component("armour").(item.ArmourComponent).Bonus, 1, true))
		e.attributes["ac"].AddEffect(item.NewEffect(e.armour.Component("armour").(item.ArmourComponent).Bonus, 1, false))
	}

	if e.IsDead() {
		return
	}

	action := e.ai.update(e, e.world)
	action.execute()
	if _, ok := action.(MountedMoveAction); ok {
		// Gets another action
		e.ai.update(e, e.world).execute()
	}
	if e.mount != nil {
		e.mount.ResetMoved()
		if e.mount.IsDead() {
			e.mount = nil
		}
	}
}

func (e *Enemy) EmptyInventory() {
	// First drop the corpse
	e.world.PlaceItem(e.location.X, e.location.Y, item.NewCorpse("head", e.name.String(), e.name.String(), e.icon))
	e.world.PlaceItem(e.location.X, e.location.Y, item.NewCorpse("body", e.name.String(), e.name.String(), e.icon))

	itemTypes := make(map[string]int)
	for _, item := range e.inventory {
		e.world.PlaceItem(e.location.X, e.location.Y, item)
		itemTypes[item.GetName()]++
	}

	if e.weapon != nil {
		e.world.PlaceItem(e.location.X, e.location.Y, e.weapon)
		itemTypes[e.weapon.GetName()]++
		e.weapon = nil
	}
	if e.armour != nil {
		e.world.PlaceItem(e.location.X, e.location.Y, e.armour)
		itemTypes[e.armour.GetName()]++
		e.armour = nil
	}

	if e.money > 0 {
		e.world.PlaceItem(e.location.X, e.location.Y, item.Money(e.money))
		if e.world.IsVisible(e.world.GetPlayer(), e.location.X, e.location.Y) {
			message.Enqueue(fmt.Sprintf("%s dropped some money.", e.name.WithDefinite()))
		}
	}

	if e.world.IsVisible(e.world.GetPlayer(), e.location.X, e.location.Y) {
		for name, count := range itemTypes {
			if count == 1 {
				message.Enqueue(fmt.Sprintf("%s dropped 1 %s.", e.name.WithDefinite(), name))
			} else {
				message.Enqueue(fmt.Sprintf("%s dropped %d %ss.", e.name.WithDefinite(), count, name))
			}
		}
	}

}

func (e *Enemy) getAmmo() *item.Item {
	for i, itm := range e.inventory {
		if itm.HasComponent("ammo") && e.Weapon().AmmoTypeMatches(itm) {
			e.inventory = append(e.inventory[:i], e.inventory[i+1:]...)
			return itm
		}
	}
	return nil
}

func (e *Enemy) PickupItem(item *item.Item) {
	e.inventory = append(e.inventory, item)
	// If item had previous owner, send theft event
	if !item.Owned(e.id) {
		event.Emit(event.NewTheft(e, item, e.location))
	}
}

func (e *Enemy) RemoveItem(itm *item.Item) {
	for i, item := range e.inventory {
		if itm.GetName() == item.GetName() {
			e.inventory = append(e.inventory[:i], e.inventory[i+1:]...)
			return
		}
	}
}

func (e *Enemy) Inventory() []*item.Item {
	return e.inventory
}

func (e *Enemy) Weapon() item.WeaponComponent {
	if e.weapon != nil {
		return e.weapon.Component("weapon").(item.WeaponComponent)
	}
	return e.unarmed
}

func (e *Enemy) ranged() bool {
	return e.Weapon().Range > 0
}

// Check whether enemy is carrying a fully loaded weapon
func (e *Enemy) weaponFullyLoaded() bool {
	return e.Weapon().IsFullyLoaded()
}

// Check whether enemy has ammo for particular wielded weapon
func (e *Enemy) hasAmmo() bool {
	for _, itm := range e.inventory {
		if itm.HasComponent("ammo") && e.Weapon().AmmoTypeMatches(itm) {
			return true
		}
	}
	return false
}

func (e *Enemy) weaponLoaded() bool {
	if e.weapon != nil && e.Weapon().NeedsAmmo() {
		return !e.Weapon().IsUnloaded()
	}
	return true

}

func (e *Enemy) applyEffects(effects item.Effects) {
	for attr, attribute := range e.attributes {
		for _, effect := range effects[attr] {
			eff := new(item.Effect)
			*eff = effect
			attribute.AddEffect(eff)
		}
	}
}

func (e *Enemy) consume(itm *item.Item) {
	e.applyEffects(itm.Component("consumable").(item.ConsumableComponent).Effects)
}

func (e *Enemy) bloodied() bool {
	return e.attributes["hp"].Value() <= e.attributes["hp"].Maximum()/2
}

func (e *Enemy) GetName() ui.Name {
	return e.name
}

func (e *Enemy) Human() bool {
	return e.human
}

func (e *Enemy) GetID() string {
	return e.id
}

func (e *Enemy) GetAlignment() worldmap.Alignment {
	return e.alignment
}

func (e *Enemy) IsCrouching() bool {
	return e.crouching
}

func (e *Enemy) Standup() {
	e.crouching = false
}

func (e *Enemy) Crouch() {
	e.crouching = true
}

func (e *Enemy) SetMap(world *worldmap.Map) {
	e.world = world

	switch ai := e.ai.(type) {
	case animalAi:
		ai.setMap(world)
	case aggAnimalAi:
		ai.setMap(world)
	case npcAi:
		ai.setMap(world)
	case barPatronAi:
		ai.setMap(world)
	}
}

func (e *Enemy) Mount() *Mount {
	return e.mount
}

func (e *Enemy) AddMount(m *Mount) {
	e.mount = m
	e.Standup()
}

func (e *Enemy) GetVisionDistance() int {
	return 20
}

func (e *Enemy) LoadMount(mounts []*Mount) {
	for _, m := range mounts {
		if e.mountID == m.GetID() {
			m.AddRider(e)
			e.mount = m
		}
	}
}

type Enemy struct {
	name       ui.Name
	id         string
	location   worldmap.Coordinates
	icon       icon.Icon
	initiative int
	attributes map[string]*worldmap.Attribute
	alignment  worldmap.Alignment
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
