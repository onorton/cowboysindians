package worldmap

import (
	"bytes"
	"encoding/json"

	"github.com/onorton/cowboysindians/item"
)

type World [][]*Grid

func NewWorld() World {
	numBlocksX := WorldConf.Width / chunkSize
	numBlocksY := WorldConf.Height / chunkSize
	world := make(World, numBlocksY)
	for row := range world {
		world[row] = make([]*Grid, numBlocksX)
		for chunk := range world[row] {
			world[row][chunk] = NewGrid(chunkSize, chunkSize)
		}
	}
	return world
}

func (world World) PlaceItem(x, y int, itm *item.Item) {
	chunk, cX, cY := world.globalToChunkAndLocal(x, y)
	chunk.items[cY][cX] = append([]*item.Item{itm}, chunk.items[cY][cX]...)
}

func (world World) IsValid(x, y int) bool {
	return x >= 0 && x < world.Width() && y >= 0 && y < world.Height()
}

func (world World) IsPassable(x, y int) bool {
	chunk, cX, cY := world.globalToChunkAndLocal(x, y)
	return chunk.passable[cY][cX]
}

func (world World) IsOccupied(x, y int) bool {
	chunk, cX, cY := world.globalToChunkAndLocal(x, y)
	return chunk.c[cY][cX] != nil
}

func (world World) Place(c Creature) {
	x, y := c.GetCoordinates()
	chunk, cX, cY := world.globalToChunkAndLocal(x, y)
	chunk.c[cY][cX] = c
}

func (world World) NewTile(tileType string, x, y int) {
	chunk, cX, cY := world.globalToChunkAndLocal(x, y)
	chunk.newTile(tileType, cX, cY)
}

func (world World) Door(x, y int) *doorComponent {
	chunk, cX, cY := world.globalToChunkAndLocal(x, y)
	return chunk.door[cY][cX]
}

func (world World) Width() int {
	return len(world[0]) * chunkSize
}

func (world World) Height() int {
	return len(world) * chunkSize
}

func (world World) globalToChunkAndLocal(x, y int) (*Grid, int, int) {
	chunkCoordinates := globalToChunkCoordinates(x, y)
	return world[chunkCoordinates.ChunkY][chunkCoordinates.ChunkX], chunkCoordinates.Local.X, chunkCoordinates.Local.Y
}

func (world World) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("[\n")

	for _, row := range world {
		for _, chunk := range row {
			chunkJson, err := json.Marshal(chunk)
			check(err)
			buffer.Write(chunkJson)
			buffer.WriteString(",\n")
		}
	}
	buffer = bytes.NewBuffer(bytes.TrimRight(buffer.Bytes(), ",\n"))
	buffer.WriteRune('\n')

	buffer.WriteString("]\n")
	return buffer.Bytes(), nil
}
