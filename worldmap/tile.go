package worldmap

import (
	"encoding/json"
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/creature"
	"github.com/onorton/cowboysindians/icon"
	"io/ioutil"
	"strconv"
	"strings"
)

type TileAttributes struct {
	Icon     rune
	Colour   termbox.Attribute
	Passable bool
	Door     bool
}

var terrainData map[string]TileAttributes = fetchTerrainData()

func fetchTerrainData() map[string]TileAttributes {
	data, err := ioutil.ReadFile("data/terrain.json")
	if err != nil {
		panic(err)
	}
	var tD map[string]TileAttributes
	json.Unmarshal(data, &tD)
	return tD
}

func newTile(name string, x, y int) Tile {

	terrain := terrainData[name]
	c := terrain.Icon
	colour := terrain.Colour
	passable := terrain.Passable
	door := terrain.Door
	return Tile{icon.NewIcon(c, colour), x, y, passable, door, nil}
}

func (t Tile) Serialize() string {

	return fmt.Sprintf("Tile{%s %d %d %v %v}", t.terrain.Serialize(), t.x, t.y, t.passable, t.door)

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
	fields := strings.Split(t, " ")

	tile.x, _ = strconv.Atoi(fields[0])
	tile.y, _ = strconv.Atoi(fields[1])
	tile.passable, _ = strconv.ParseBool(fields[2])
	tile.door, _ = strconv.ParseBool(fields[3])

	return tile
}
func (t Tile) render(x, y int) {

	if t.c != nil {
		t.c.Render(x, y)
	} else if t.door {
		t.terrain.RenderDoor(x, y, t.passable)
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
}
