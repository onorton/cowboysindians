package npc

import (
	"encoding/json"
	"io/ioutil"

	"github.com/onorton/cowboysindians/event"
	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/item"
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
	Unarmed     item.WeaponComponent
	Encumbrance int
	AiType      string
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

func NewMount(name string, x, y int, world *worldmap.Map) *Npc {
	mount := mountData[name]
	id := xid.New().String()
	location := worldmap.Coordinates{x, y}
	ai := newAi(mount.AiType, id, world, location, nil, nil, nil, nil)

	attributes := map[string]*worldmap.Attribute{
		"hp":          worldmap.NewAttribute(mount.Hp, mount.Hp),
		"ac":          worldmap.NewAttribute(mount.Ac, mount.Ac),
		"str":         worldmap.NewAttribute(mount.Str, mount.Str),
		"dex":         worldmap.NewAttribute(mount.Dex, mount.Dex),
		"encumbrance": worldmap.NewAttribute(mount.Encumbrance, mount.Encumbrance)}

	npc := &Npc{&ui.PlainName{name}, id, worldmap.Coordinates{x, y}, mount.Icon, mount.Initiative, attributes, worldmap.Neutral, false, 0, mount.Unarmed, nil, nil, make([]*item.Item, 0), &mountableComponent{}, "", nil, world, ai, nil, false}

	event.Subscribe(npc)
	return npc
}

func generateMount(mountProbabilities map[string]float64, x, y int) *Npc {
	if mountProbabilities == nil {
		return nil
	}

	mountType := chooseType(mountProbabilities)
	if mountType != "None" {
		return NewMount(mountType, x, y, nil)
	}
	return nil
}

type mountableComponent struct {
	rider Rider
	Moved bool
}

type Rider interface {
	IsDead() bool
	TakeDamage(item.Damage, item.Effects, int)
	GetAlignment() worldmap.Alignment
	GetCoordinates() (int, int)
	Mount() *Npc
	AddMount(*Npc)
}
