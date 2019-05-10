package worldmap

import "github.com/onorton/cowboysindians/item"

type World [][]*Grid

func NewWorld(width, height int) World {
	numBlocksX := width / chunkSize
	numBlocksY := height / chunkSize
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

func (world World) IsPassable(x, y int) bool {
	chunk, cX, cY := world.globalToChunkAndLocal(x, y)
	return chunk.passable[cY][cX]
}

func (world World) IsOccupied(x, y int) bool {
	chunk, cX, cY := world.globalToChunkAndLocal(x, y)
	return chunk.c[cY][cX] != nil
}

func (world World) Place(c Creature, x, y int) {
	chunk, cX, cY := world.globalToChunkAndLocal(x, y)
	chunk.c[cY][cX] = c
}

func (world World) NewTile(tileType string, x, y int) {
	chunk, cX, cY := world.globalToChunkAndLocal(x, y)
	chunk.newTile(tileType, cX, cY)
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
