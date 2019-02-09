package worldmap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/onorton/cowboysindians/creature"
	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/item"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type TileAttributes struct {
	Icon     icon.Icon
	Passable bool
	Door     bool
}

var terrainData map[string]TileAttributes = fetchTerrainData()

func fetchTerrainData() map[string]TileAttributes {
	data, err := ioutil.ReadFile("data/terrain.json")
	check(err)
	var tD map[string]TileAttributes
	err = json.Unmarshal(data, &tD)
	check(err)
	return tD
}

func newTile(name string, x, y int) Tile {

	terrain := terrainData[name]
	icon := terrain.Icon
	passable := terrain.Passable
	door := terrain.Door
	return Tile{icon, x, y, passable, door, nil, make([]item.Item, 0)}
}

func (t Tile) Serialize() string {
	items := "["
	for _, item := range t.items {
		items += fmt.Sprintf("%s ", item.Serialize())
	}
	items += "]"

	iconJson, err := json.Marshal(t.terrain)
	check(err)

	return fmt.Sprintf("Tile{%s %d %d %v %v %s}", iconJson, t.x, t.y, t.passable, t.door, items)

}

func DeserializeTile(t string) Tile {

	if len(t) == 0 || t[0] != '{' {

		return Tile{}
	}

	b := 0
	e := len(t)
	for i, c := range t {
		if c == '{' && b == 0 {
			b = i
		}
		if c == '}' && e == len(t) {
			e = i
		}
	}

	e++
	tile := Tile{}

	err := json.Unmarshal([]byte(t[b:e]), &(tile.terrain))
	check(err)

	t = t[(e + 1):]
	fields := strings.SplitN(t, " ", 5)

	tile.x, _ = strconv.Atoi(fields[0])
	tile.y, _ = strconv.Atoi(fields[1])
	tile.passable, _ = strconv.ParseBool(fields[2])
	tile.door, _ = strconv.ParseBool(fields[3])
	itemStrings := fields[4][1 : len(fields[4])-2]
	items := regexp.MustCompile("(Ammo)|(Armour)|(Consumable)|(Item)|(Weapon)").Split(itemStrings, -1)
	starter := regexp.MustCompile("(Ammo)|(Armour)|(Consumable)|(Item)|(Weapon)").FindAllString(itemStrings, -1)
	items = items[1:]
	tile.items = make([]item.Item, len(items))

	for i, itemString := range items {
		switch starter[i] {
		case "Item":
			tile.items[i] = item.Deserialize(itemString)
		case "Weapon":
			tile.items[i] = item.DeserializeWeapon(itemString)
		case "Armour":
			tile.items[i] = item.DeserializeArmour(itemString)
		case "Ammo":
			tile.items[i] = item.DeserializeAmmo(itemString)
		case "Consumable":
			tile.items[i] = item.DeserializeConsumable(itemString)
		}
	}
	return tile
}

func (t *Tile) PlaceItem(itm item.Item) {
	t.items = append([]item.Item{itm}, t.items...)
}

func (t Tile) render(x, y int) {
	if t.c != nil {
		t.c.Render(x, y)
	} else if t.door {
		t.terrain.RenderDoor(x, y, t.passable)
	} else if len(t.items) != 0 {
		t.items[0].Render(x, y)
	} else {
		t.terrain.Render(x, y)
	}
}

type Tile struct {
	terrain  icon.Icon
	x        int
	y        int
	passable bool
	door     bool
	c        creature.Creature
	items    []item.Item
}
