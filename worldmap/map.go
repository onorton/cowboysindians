package worldmap

import (
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/creature"
	"github.com/onorton/cowboysindians/icon"
)

const padding = 5

func NewTile(c rune, colour termbox.Attribute, x, y int) Tile {
	return Tile{icon.NewIcon(c, colour), x, y, nil}
}

func (t Tile) render(x, y int) {

	if t.c != nil {
		t.c.Render(x, y)
	} else {
		t.terrain.Render(x, y)
	}
}

type Tile struct {
	terrain icon.Icon
	x       int
	y       int
	c       *creature.Player
}

func NewMap(width, height, viewerWidth, viewerHeight int) Map {
	grid := make([][]Tile, height)
	for i := 0; i < height; i++ {
		row := make([]Tile, width)
		for j := 0; j < width; j++ {
			row[j] = NewTile('.', termbox.ColorWhite, j, i)
		}
		grid[i] = row
	}
	viewer := new(Viewer)
	viewer.x = 0
	viewer.y = 0
	viewer.width = viewerWidth
	viewer.height = viewerHeight
	return Map{grid, viewer}
}

type Viewer struct {
	x      int
	y      int
	width  int
	height int
}

type Map struct {
	grid [][]Tile
	v    *Viewer
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
	rX := c.X - m.v.x
	rY := c.Y - m.v.y

	if rX < padding && c.X >= padding {
		m.v.x--
	}
	if rX > m.v.width-padding && c.X <= len(m.grid[0])-padding {
		m.v.x++
	}
	if rY < padding && c.Y >= padding {
		m.v.y--
	}
	if rY > m.v.height-padding && c.Y <= len(m.grid)-padding {
		m.v.y++
	}
}

func (m Map) Render() {
	for y, row := range m.grid {
		for x, tile := range row {
			rX := x - m.v.x
			rY := y - m.v.y
			if rX >= 0 && rX < m.v.width && rY >= 0 && rY < m.v.height {
				tile.render(rX, rY)
			}
		}
	}
	termbox.Flush()
}
