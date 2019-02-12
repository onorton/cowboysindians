package worldmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/ui"
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

func (t *Tile) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	keys := []string{"Terrain", "X", "Y", "Passable", "Door", "Items"}

	tileValues := map[string]interface{}{
		"Terrain":  t.terrain,
		"X":        t.x,
		"Y":        t.y,
		"Passable": t.passable,
		"Door":     t.door,
		"Items":    t.items,
	}

	length := len(tileValues)
	count := 0

	for _, key := range keys {
		jsonValue, err := json.Marshal(tileValues[key])
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

func (t *Tile) UnmarshalJSON(data []byte) error {

	type tileJson struct {
		Terrain  icon.Icon
		X        int
		Y        int
		Passable bool
		Door     bool
		Items    item.ItemList
	}
	v := tileJson{}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	t.terrain = v.Terrain
	t.x = v.X
	t.y = v.Y
	t.passable = v.Passable
	t.door = v.Door
	t.items = v.Items

	return nil

}

func (t *Tile) PlaceItem(itm item.Item) {
	t.items = append([]item.Item{itm}, t.items...)
}

func (t Tile) render() ui.Element {
	if t.c != nil {
		return t.c.Render()
	} else if t.door {
		return t.terrain.RenderDoor(t.passable)
	} else if len(t.items) != 0 {
		return t.items[0].Render()
	} else {
		return t.terrain.Render()
	}
}

type Tile struct {
	terrain  icon.Icon
	x        int
	y        int
	passable bool
	door     bool
	c        Creature
	items    []item.Item
}
