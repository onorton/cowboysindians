package npc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
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
	id := xid.New()
	location := worldmap.Coordinates{x, y}
	m := &Mount{&ui.PlainName{name}, id.String(), location, mount.Icon, mount.Initiative, mount.Hp, mount.Hp, mount.Ac, mount.Str, mount.Dex, mount.Encumbrance, nil, world, false, worldmap.NewRandomWaypoint(world, location)}
	return m
}
func (m *Mount) Render() ui.Element {
	return m.icon.Render()
}

func (m *Mount) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	keys := []string{"Name", "Id", "Location", "Icon", "Initiative", "Hp", "MaxHp", "AC", "Str", "Dex", "Encumbrance", "WaypointSystem"}

	mountValues := map[string]interface{}{
		"Name":           m.name,
		"Id":             m.id,
		"Location":       m.location,
		"Icon":           m.icon,
		"Initiative":     m.initiative,
		"Hp":             m.hp,
		"MaxHp":          m.maxHp,
		"AC":             m.ac,
		"Str":            m.str,
		"Dex":            m.dex,
		"Encumbrance":    m.encumbrance,
		"WaypointSystem": m.waypoint,
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
		Name           *ui.PlainName
		Id             string
		Location       worldmap.Coordinates
		Icon           icon.Icon
		Initiative     int
		Hp             int
		MaxHp          int
		AC             int
		Str            int
		Dex            int
		Encumbrance    int
		WaypointSystem map[string]interface{}
	}
	var v mountJson

	json.Unmarshal(data, &v)

	m.name = v.Name
	m.id = v.Id
	m.location = v.Location
	m.icon = v.Icon
	m.initiative = v.Initiative
	m.hp = v.Hp
	m.maxHp = v.MaxHp
	m.ac = v.AC
	m.str = v.Str
	m.dex = v.Dex
	m.encumbrance = v.Encumbrance
	m.waypoint = worldmap.UnmarshalWaypointSystem(v.WaypointSystem)
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
	m.attack(c, worldmap.GetBonus(m.str), worldmap.GetBonus(m.str))
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
	return roll > m.ac
}
func (m *Mount) TakeDamage(damage int) {
	m.hp -= damage

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
	return m.hp <= 0
}

func (m *Mount) Update() (int, int) {
	if m.rider != nil {
		if m.rider.IsDead() {
			m.RemoveRider()
		} else {
			m.location.X, m.location.Y = m.rider.GetCoordinates()
		}
	} else {
		waypoint := m.waypoint.NextWaypoint(m.location)
		aiMap := getWaypointMap(waypoint, m.world, m.location, m.GetVisionDistance())
		current := aiMap[m.GetVisionDistance()][m.GetVisionDistance()]
		possibleLocations := make([]worldmap.Coordinates, 0)
		// Find adjacent locations closer to the goal
		for i := -1; i <= 1; i++ {
			for j := -1; j <= 1; j++ {
				nX := m.location.X + i
				nY := m.location.Y + j
				if aiMap[nY-m.location.Y+m.GetVisionDistance()][nX-m.location.X+m.GetVisionDistance()] <= current {
					// Add if not occupied
					if m.world.IsValid(nX, nY) && !m.world.IsOccupied(nX, nY) {
						possibleLocations = append(possibleLocations, worldmap.Coordinates{nX, nY})
					}
				}
			}
		}
		if len(possibleLocations) > 0 {
			l := possibleLocations[rand.Intn(len(possibleLocations))]
			return l.X, l.Y
		}

	}
	return m.location.X, m.location.Y
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

func (m *Mount) heal(amount int) {
	originalHp := m.hp
	m.hp = int(math.Min(float64(originalHp+amount), float64(m.maxHp)))
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

func (m *Mount) AddRider(c worldmap.Creature) {
	m.rider = c
}

func (m *Mount) RemoveRider() {
	m.rider = nil
}

func (m *Mount) IsMounted() bool {
	return m.rider != nil
}

func (m *Mount) SetMap(world *worldmap.Map) {
	m.world = world
	if w, ok := m.waypoint.(*worldmap.RandomWaypoint); ok {
		w.SetMap(world)
	}
}

func (m *Mount) GetIcon() icon.Icon {
	return m.icon
}

func (m *Mount) GetEncumbrance() int {
	return m.encumbrance
}

func (m *Mount) GetMount() worldmap.Creature {
	return nil
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
	name        *ui.PlainName
	id          string
	location    worldmap.Coordinates
	icon        icon.Icon
	initiative  int
	hp          int
	maxHp       int
	ac          int
	str         int
	dex         int
	encumbrance int
	rider       worldmap.Creature
	world       *worldmap.Map
	moved       bool
	waypoint    worldmap.WaypointSystem
}
