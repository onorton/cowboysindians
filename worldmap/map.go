package worldmap

import (
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/creature"
	"github.com/onorton/cowboysindians/icon"
)

func NewTile(c rune, colour termbox.Attribute, x, y int) Tile {
	return Tile{icon.NewIcon(c, colour), x, y, nil}
}

func (t Tile) render() {

	if t.c != nil {
		t.c.Render()
	} else {
		t.terrain.Render(t.x, t.y)
	}
}

type Tile struct {
	terrain icon.Icon
	x       int
	y       int
	c       *creature.Player
}

func NewMap(width, height int) Map {
	grid := make([][]Tile, height)
	for i := 0; i < height; i++ {
		row := make([]Tile, width)
		for j := 0; j < width; j++ {
			row[j] = NewTile('.', termbox.ColorWhite, j, i)
		}
		grid[i] = row
	}
	return Map{grid}
}

type Map struct {
	grid [][]Tile
}

func (m Map) MoveCreature(c *creature.Player) {

	for y, row := range m.grid {
		for x, tile := range row {

			if tile.c == c {
				m.grid[y][x].c = nil
			}
		}
	}
	m.grid[c.Y][c.X].c = c
}

func (m Map) Render() {
	for _, row := range m.grid {
		for _, tile := range row {
			tile.render()
		}
	}
}
