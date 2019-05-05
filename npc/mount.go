package npc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"

	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/ui"
	"github.com/onorton/cowboysindians/worldmap"
	"github.com/rs/xid"
)

type MountAttributes struct {
	Icon        icon.Icon
	Initiative  int
	Hp          int
	Ac          int
	Str         int
	Dex         int
	Encumbrance int
}

var mountData map[string]MountAttributes = fetchMountData()

func fetchMountData() map[string]MountAttributes {
	data, err := ioutil.ReadFile("data/mount.json")
	check(err)
	var eD map[string]MountAttributes
	err = json.Unmarshal(data, &eD)
	check(err)
	return eD
}

func NewMount(name string, x, y int, world *worldmap.Map) *Mount {
	mount := mountData[name]
	id := xid.New().String()
	location := worldmap.Coordinates{x, y}
	ai := mountAi{worldmap.NewRandomWaypoint(world, location)}
	attributes := map[string]*worldmap.Attribute{
		"hp":          worldmap.NewAttribute(mount.Hp, mount.Hp),
		"ac":          worldmap.NewAttribute(mount.Ac, mount.Ac),
		"str":         worldmap.NewAttribute(mount.Str, mount.Str),
		"dex":         worldmap.NewAttribute(mount.Dex, mount.Dex),
		"encumbrance": worldmap.NewAttribute(mount.Encumbrance, mount.Encumbrance),
	}
	m := &Mount{&ui.PlainName{name}, id, location, mount.Icon, mount.Initiative, attributes, nil, world, false, ai}
	return m
}
func (m *Mount) Render() ui.Element {
	return m.icon.Render()
}

func (m *Mount) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	keys := []string{"Name", "Id", "Location", "Icon", "Initiative", "Attributes", "Ai"}

	mountValues := map[string]interface{}{
		"Name":       m.name,
		"Id":         m.id,
		"Location":   m.location,
		"Icon":       m.icon,
		"Initiative": m.initiative,
		"Attributes": m.attributes,
		"Ai":         m.ai,
	}

	length := len(mountValues)
	count := 0

	for _, key := range keys {
		jsonValue, err := json.Marshal(mountValues[key])
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

func (m *Mount) UnmarshalJSON(data []byte) error {

	type mountJson struct {
		Name       *ui.PlainName
		Id         string
		Location   worldmap.Coordinates
		Icon       icon.Icon
		Initiative int
		Attributes map[string]*worldmap.Attribute
		Ai         map[string]interface{}
	}
	var v mountJson

	json.Unmarshal(data, &v)

	m.name = v.Name
	m.id = v.Id
	m.location = v.Location
	m.icon = v.Icon
	m.initiative = v.Initiative
	m.attributes = v.Attributes
	m.ai = unmarshalAi(v.Ai)
	return nil
}

func (m *Mount) GetCoordinates() (int, int) {
	return m.location.X, m.location.Y
}

func (m *Mount) SetCoordinates(x int, y int) {
	m.location = worldmap.Coordinates{x, y}
}

func (m *Mount) GetInitiative() int {
	return m.initiative
}

func (m *Mount) MeleeAttack(c worldmap.Creature) {
	m.attack(c, worldmap.GetBonus(m.attributes["str"].Value()), worldmap.GetBonus(m.attributes["str"].Value()))
}
func (m *Mount) attack(c worldmap.Creature, hitBonus, damageBonus int) {

	hits := c.AttackHits(rand.Intn(20) + hitBonus + 1)
	if hits {
		c.TakeDamage(damageBonus)
	}
	if c.GetAlignment() == worldmap.Player {
		if hits {
			message.Enqueue(fmt.Sprintf("%s hit you.", m.name.WithDefinite()))
		} else {
			message.Enqueue(fmt.Sprintf("%s missed you.", m.name.WithDefinite()))
		}
	}

}

func (m *Mount) AttackHits(roll int) bool {
	return roll > m.attributes["ac"].Value()
}
func (m *Mount) TakeDamage(damage int) {
	m.attributes["hp"].AddEffect(item.NewInstantEffect(-damage))
	// Rider takes falling damage if mount dies
	if m.rider != nil && m.IsDead() {
		m.rider.TakeDamage(rand.Intn(4) + 1)
		if m.rider.GetAlignment() == worldmap.Player {
			message.Enqueue(fmt.Sprintf("Your %s died and you fell.", m.name))
		}
		m.RemoveRider()
	}
}

func (m *Mount) IsDead() bool {
	return m.attributes["hp"].Value() == 0
}

func (m *Mount) Update() {
	for _, attribute := range m.attributes {
		attribute.Update()
	}
	if m.IsDead() {
		return
	}

	if m.rider != nil {
		if m.rider.IsDead() {
			m.RemoveRider()
		} else {
			rX, rY := m.rider.GetCoordinates()
			m.location = worldmap.Coordinates{rX, rY}
			return
		}
	}
	action := m.ai.update(m, m.world)
	action.execute()
}

func (m *Mount) ResetMoved() {
	m.moved = false
}

func (m *Mount) Move() {
	m.moved = true
}

func (m *Mount) Moved() bool {
	return m.moved
}

func (m *Mount) consume(itm *item.Item) {
	for attr, attribute := range m.attributes {
		for _, effect := range itm.Component("consumable").(item.ConsumableComponent).Effects[attr] {
			attribute.AddEffect(&effect)
		}
	}
}

func (m *Mount) bloodied() bool {
	return m.attributes["hp"].Value() <= m.attributes["hp"].Maximum()/2
}

func (m *Mount) GetName() ui.Name {
	return m.name
}

func (m *Mount) GetAlignment() worldmap.Alignment {
	return worldmap.Neutral
}

func (m *Mount) IsCrouching() bool {
	return false
}

func (m *Mount) Crouch() {
}

func (m *Mount) Standup() {
}

func (m *Mount) AddRider(r Rider) {
	m.rider = r
}

func (m *Mount) RemoveRider() {
	m.rider = nil
}

func (m *Mount) IsMounted() bool {
	return m.rider != nil
}

func (m *Mount) SetMap(world *worldmap.Map) {
	m.world = world

	switch ai := m.ai.(type) {
	case mountAi:
		ai.setMap(world)
	case npcAi:
		ai.setMap(world)
	}
}

func (m *Mount) GetIcon() icon.Icon {
	return m.icon
}

func (m *Mount) GetEncumbrance() int {
	return m.attributes["encumbrance"].Value()
}

func (m *Mount) GetVisionDistance() int {
	return 20
}

func (m *Mount) DropCorpse() {
	m.world.PlaceItem(m.location.X, m.location.Y, item.NewCorpse("head", m.id, m.name.String(), m.icon))
	m.world.PlaceItem(m.location.X, m.location.Y, item.NewCorpse("body", m.id, m.name.String(), m.icon))
}

func (m *Mount) GetID() string {
	return m.id
}

type Mount struct {
	name       *ui.PlainName
	id         string
	location   worldmap.Coordinates
	icon       icon.Icon
	initiative int
	attributes map[string]*worldmap.Attribute
	rider      Rider
	world      *worldmap.Map
	moved      bool
	ai         ai
}

type Rider interface {
	IsDead() bool
	TakeDamage(int)
	GetAlignment() worldmap.Alignment
	GetCoordinates() (int, int)
	Mount() *Mount
	AddMount(*Mount)
}
