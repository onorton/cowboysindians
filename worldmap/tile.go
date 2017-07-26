package worldmap

import (
	"encoding/json"
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/creature"
	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/item"

	"io/ioutil"
	"strconv"
	"strings"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type TileAttributes struct {
	Icon     rune
	Colour   termbox.Attribute
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
	c := terrain.Icon
	colour := terrain.Colour
	passable := terrain.Passable
	door := terrain.Door
	return Tile{icon.NewIcon(c, colour), x, y, passable, door, nil, make([]item.Item, 0)}
}

func (t Tile) Serialize() string {
	items := "["
	for _, item := range t.items {
		items += fmt.Sprintf("%s ", item.Serialize())
	}
	items += "]"

	return fmt.Sprintf("Tile{%s %d %d %v %v %s}", t.terrain.Serialize(), t.x, t.y, t.passable, t.door, items)

}

func DeserializeTile(t string) Tile {

	if len(t) == 0 || t[0] != '{' {

		return Tile{}
	}

	b := 0
	e := len(t)
	for i, c := range t {
		if c == 'I' && b == 0 {
			b = i
		}
		if c == '}' && e == len(t) {
			e = i
		}
	}

	e++
	tile := Tile{}

	tile.terrain = icon.Deserialize(t[b:e])

	t = t[(e + 1):]
	fields := strings.SplitN(t, " ", 5)

	tile.x, _ = strconv.Atoi(fields[0])
	tile.y, _ = strconv.Atoi(fields[1])
	tile.passable, _ = strconv.ParseBool(fields[2])
	tile.door, _ = strconv.ParseBool(fields[3])
	tile.items = make([]item.Item, 0)
	items := strings.Split(fields[4][1:len(fields[4])-2], "Item{")
	items = items[1:]
	for _, itemString := range items {
		itemString = fmt.Sprintf("Item{%s", itemString)
		itm := item.Deserialize(itemString)
		tile.items = append(tile.items, itm)
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
