package worldmap

import (
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/creature"
	"github.com/onorton/cowboysindians/icon"
)

const padding = 5

func NewTile(c rune, colour termbox.Attribute, x, y int, passable bool) Tile {
	return Tile{icon.NewIcon(c, colour), x, y, passable, nil}
}

func (t Tile) render(x, y int) {

	if t.c != nil {
		t.c.Render(x, y)
	} else {
		t.terrain.Render(x, y)
	}
}

type Tile struct {
	terrain  icon.Icon
	x        int
	y        int
	passable bool
	c        *creature.Player
}

func NewMap(width, height, viewerWidth, viewerHeight int) Map {
	grid := make([][]Tile, height)
	for i := 0; i < height; i++ {
		row := make([]Tile, width)
		for j := 0; j < width; j++ {
			row[j] = NewTile('.', termbox.ColorWhite, j, i, true)

		}
		grid[i] = row
	}

	grid[0][4] = NewTile('|', termbox.ColorWhite, 4, 0, false)
	grid[0][5] = NewTile('-', termbox.ColorWhite, 5, 0, false)
	grid[0][6] = NewTile('|', termbox.ColorWhite, 6, 0, false)
	grid[1][4] = NewTile('|', termbox.ColorWhite, 4, 1, false)
	grid[2][4] = NewTile('|', termbox.ColorWhite, 4, 2, false)
	grid[2][5] = NewTile('-', termbox.ColorWhite, 5, 2, false)
	grid[2][6] = NewTile('|', termbox.ColorWhite, 6, 2, false)
	grid[1][6] = NewTile('|', termbox.ColorWhite, 6, 1, false)

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

func (m Map) MoveCreature(c *creature.Player, x, y int) {

	if !m.grid[y][x].passable {
		return
	}

	for y, row := range m.grid {
		for x, tile := range row {

			if tile.c == c {
				m.grid[y][x].c = nil
			}
		}
	}
	c.Y = y
	c.X = x
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
