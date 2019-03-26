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
	Icon         icon.Icon
	Passable     bool
	BlocksVision bool
	Door         bool
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

func NewTile(name string) Tile {

	terrain := terrainData[name]
	icon := terrain.Icon
	passable := terrain.Passable
	blocksV := terrain.BlocksVision
	door := terrain.Door
	if door {
		return &Door{icon, passable, blocksV, blocksV, nil, make([]item.Item, 0)}
	} else {
		return &NormalTile{icon, passable, blocksV, nil, make([]item.Item, 0)}
	}

}

type NormalTile struct {
	terrain  icon.Icon
	passable bool
	blocksV  bool
	c        Creature
	items    []item.Item
}

func (t *NormalTile) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	keys := []string{"Terrain", "Passable", "BlocksVision", "Items"}

	tileValues := map[string]interface{}{
		"Terrain":      t.terrain,
		"Passable":     t.passable,
		"BlocksVision": t.blocksV,
		"Items":        t.items,
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

func (t *NormalTile) UnmarshalJSON(data []byte) error {

	type tileJson struct {
		Terrain      icon.Icon
		Passable     bool
		BlocksVision bool
		Items        item.ItemList
	}
	v := tileJson{}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	t.terrain = v.Terrain
	t.passable = v.Passable
	t.blocksV = v.BlocksVision
	t.items = v.Items

	return nil

}

func (t *NormalTile) placeItem(itm item.Item) {
	t.items = append([]item.Item{itm}, t.items...)
}

func (t *NormalTile) getItems() []item.Item {
	items := t.items
	t.items = make([]item.Item, 0)
	return items
}

func (t *NormalTile) givesCover() bool {
	cover := !t.passable

	for _, item := range t.items {
		cover = cover || item.GivesCover()
	}
	return cover
}

func (t *NormalTile) blocksVision() bool {
	return t.blocksV
}

func (t *NormalTile) isPassable() bool {
	return t.passable
}

func (t *NormalTile) isOccupied() bool {
	return t.c != nil
}

func (t *NormalTile) getCreature() Creature {
	return t.c
}

func (t *NormalTile) setCreature(c Creature) {
	t.c = c
}

func (t *NormalTile) render() ui.Element {
	if t.c != nil {
		return t.c.Render()
	} else if len(t.items) != 0 {
		// pick an item that gives cover if it exists
		for _, item := range t.items {
			if item.GivesCover() {
				return item.Render()
			}
		}
		return t.items[0].Render()
	} else {
		return t.terrain.Render()
	}
}

type Door struct {
	terrain       icon.Icon
	passable      bool
	blocksV       bool
	blocksVClosed bool
	c             Creature
	items         []item.Item
}

func (d *Door) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	keys := []string{"Terrain", "Passable", "BlocksVision", "BlocksVisionClosed", "Items"}

	tileValues := map[string]interface{}{
		"Terrain":            d.terrain,
		"Passable":           d.passable,
		"BlocksVision":       d.blocksV,
		"BlocksVisionClosed": d.blocksVClosed,
		"Items":              d.items,
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

func (d *Door) UnmarshalJSON(data []byte) error {

	type doorJson struct {
		Terrain            icon.Icon
		Passable           bool
		BlocksVision       bool
		BlocksVisionClosed bool
		Items              item.ItemList
	}
	v := doorJson{}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	d.terrain = v.Terrain
	d.passable = v.Passable
	d.blocksV = v.BlocksVision
	d.blocksVClosed = v.BlocksVisionClosed
	d.items = v.Items

	return nil

}

func (d *Door) givesCover() bool {
	return !d.passable
}

func (d *Door) blocksVision() bool {
	return d.blocksV
}

func (d *Door) isPassable() bool {
	return d.passable
}

func (d *Door) isOccupied() bool {
	return d.c != nil
}

func (d *Door) getCreature() Creature {
	return d.c
}

func (d *Door) setCreature(c Creature) {
	d.c = c
}

func (d *Door) getItems() []item.Item {
	items := d.items
	d.items = make([]item.Item, 0)
	return items
}

func (d *Door) placeItem(itm item.Item) {
	d.items = append([]item.Item{itm}, d.items...)
}

func (d *Door) render() ui.Element {
	if d.c != nil {
		return d.c.Render()
	} else if d.passable {

		if len(d.items) != 0 {
			// pick an item that gives cover if it exists
			for _, item := range d.items {
				if item.GivesCover() {
					return item.Render()
				}
			}

			return d.items[0].Render()
		}

		return terrainData["ground"].Icon.Render()
	} else {
		return d.terrain.Render()
	}
}

func unmarshalTile(tile map[string]interface{}) Tile {
	tileJson, err := json.Marshal(tile)
	check(err)

	if _, ok := tile["BlocksVisionClosed"]; ok {
		var d Door
		err = json.Unmarshal(tileJson, &d)
		check(err)
		return &d
	}

	var t NormalTile
	err = json.Unmarshal(tileJson, &t)
	check(err)
	return &t
}

type Tile interface {
	render() ui.Element
	givesCover() bool
	blocksVision() bool
	isPassable() bool
	isOccupied() bool
	getCreature() Creature
	setCreature(Creature)
	getItems() []item.Item
	placeItem(item.Item)
}
