package worldmap

import (
	termbox "github.com/nsf/termbox-go"
)

func (i Icon) render(x, y int) {
	termbox.SetCell(x, y, i.icon, i.colour, termbox.ColorDefault)
}

type Icon struct {
	icon   rune
	colour termbox.Attribute
}

func NewTile(icon rune, colour termbox.Attribute, x, y int) Tile {
	return Tile{Icon{icon, colour}, x, y}
}

func (t Tile) render() {
	t.terrain.render(t.x, t.y)
}

type Tile struct {
	terrain Icon
	x       int
	y       int
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

func (m Map) Render() {
	for _, row := range m.grid {
		for _, tile := range row {
			tile.render()
		}
	}
}
