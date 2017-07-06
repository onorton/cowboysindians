package worldmap

import (
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/creature"
	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/message"
)

const padding = 5

func NewTile(c rune, colour termbox.Attribute, x, y int, passable bool) Tile {
	return Tile{icon.NewIcon(c, colour), x, y, passable, false, nil}
}

func NewDoor(x, y int, open bool) Tile {
	return Tile{icon.NewIcon('+', termbox.ColorWhite), x, y, open, true, nil}
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
	grid[1][4] = NewDoor(4, 1, false)
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

func (m Map) ToggleDoor(x, y int, open bool) bool {
	message.PrintMessage("Which direction?")
	height := len(m.grid)
	width := len(m.grid[0])
	for {
		validMove := true
		e := termbox.PollEvent()
		if e.Type == termbox.EventKey {
			switch e.Key {
			case termbox.KeyArrowLeft:
				if x != 0 {
					x--
				}
			case termbox.KeyArrowRight:
				if x < width-1 {
					x++
				}
			case termbox.KeyArrowUp:
				if y != 0 {
					y--
				}
			case termbox.KeyArrowDown:
				if y < height-1 {
					y++
				}
			case termbox.KeyEnter:
				message.PrintMessage("Never mind...")
				return false
			default:
				{

					switch e.Ch {
					case '1':
						if x != 0 && y < height-1 {
							x--
							y++
						}
					case '2':
						if y < height-1 {
							y++
						}
					case '3':
						if x < width-1 && y < height-1 {
							x++
							y++
						}

					case '4':
						if x != 0 {
							x--
						}
					case '6':
						if x < width-1 {
							x++
						}
					case '7':
						if x != 0 && x != 0 {
							x--
							y--
						}
					case '8':
						if y != 0 {
							y--
						}
					case '9':
						if y != 0 && x < width-1 {
							y--
							x++
						}
					default:
						message.PrintMessage("Invalid direction.")
						validMove = false

					}
				}

			}

		}

		if validMove {
			break
		}
	}
	if m.grid[y][x].door {
		if m.grid[y][x].passable != open {
			m.grid[y][x].passable = open
			return true
		} else {
			if open {
				message.PrintMessage("The door is already open.")
			} else {
				message.PrintMessage("The door is already closed.")
			}
		}
	} else {
		message.PrintMessage("You see no door there.")
	}
	return false

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
