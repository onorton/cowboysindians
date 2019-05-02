package worldmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/item"
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

type Grid struct {
	terrain       [][]icon.Icon
	passable      [][]bool
	door          [][]bool
	blocksVision  [][]bool
	blocksVClosed [][]bool
	c             [][]Creature
	items         [][][]*item.Item
}

func NewGrid(width int, height int) *Grid {
	terrain := make([][]icon.Icon, height)
	passable := make([][]bool, height)
	door := make([][]bool, height)
	blocksVision := make([][]bool, height)
	blocksVClosed := make([][]bool, height)
	c := make([][]Creature, height)
	items := make([][][]*item.Item, height)

	grid := &Grid{}

	for y := 0; y < height; y++ {
		terrain[y] = make([]icon.Icon, width)
		passable[y] = make([]bool, width)
		door[y] = make([]bool, width)
		blocksVision[y] = make([]bool, width)
		blocksVClosed[y] = make([]bool, width)
		c[y] = make([]Creature, width)
		items[y] = make([][]*item.Item, width)
	}

	grid.terrain = terrain
	grid.passable = passable
	grid.door = door
	grid.blocksVision = blocksVision
	grid.blocksVClosed = blocksVClosed
	grid.c = c
	grid.items = items

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			grid.NewTile("ground", x, y)
		}
	}

	return grid

}

func (grid *Grid) Width() int {
	return len(grid.terrain[0])
}

func (grid *Grid) Height() int {
	return len(grid.terrain)
}

func (grid *Grid) NewTile(tileType string, x, y int) {
	terrain := terrainData[tileType]
	grid.terrain[y][x] = terrain.Icon
	grid.passable[y][x] = terrain.Passable
	grid.door[y][x] = terrain.Door
	grid.blocksVision[y][x] = terrain.BlocksVision
	grid.blocksVClosed[y][x] = terrain.BlocksVision
}

func (grid *Grid) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	keys := []string{"Terrain", "Passable", "Door", "BlocksVision", "BlocksVisionClosed", "Items"}

	gridValues := map[string]interface{}{
		"Terrain":            grid.terrain,
		"Passable":           grid.passable,
		"Door":               grid.door,
		"BlocksVision":       grid.blocksVision,
		"BlocksVisionClosed": grid.blocksVClosed,
		"Items":              grid.items,
	}

	length := len(gridValues)
	count := 0

	for _, key := range keys {
		jsonValue, err := json.Marshal(gridValues[key])
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

func (grid *Grid) UnmarshalJSON(data []byte) error {

	type gridJson struct {
		Terrain            [][]icon.Icon
		Passable           [][]bool
		Door               [][]bool
		BlocksVision       [][]bool
		BlocksVisionClosed [][]bool
		Items              [][][]*item.Item
	}
	v := gridJson{}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	grid.terrain = v.Terrain
	grid.passable = v.Passable
	grid.door = v.Door
	grid.blocksVision = v.BlocksVision
	grid.blocksVClosed = v.BlocksVisionClosed
	grid.items = v.Items
	grid.c = make([][]Creature, grid.Height())
	for y := 0; y < grid.Height(); y++ {
		grid.c[y] = make([]Creature, grid.Width())
	}
	return nil

}
