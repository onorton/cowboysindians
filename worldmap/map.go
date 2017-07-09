package worldmap

import (
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/creature"
	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/message"
	"strconv"
	"strings"
)

const padding = 5

func NewTile(c rune, colour termbox.Attribute, x, y int, passable bool) Tile {
	return Tile{icon.NewIcon(c, colour), x, y, passable, false, nil}
}

func NewDoor(x, y int, open bool) Tile {
	return Tile{icon.NewIcon('+', termbox.ColorWhite), x, y, open, true, nil}
}

func (t Tile) Serialize() string {
	return fmt.Sprintf("Tile{%s %d %d %v %v %v}", t.terrain.Serialize(), t.x, t.y, t.passable, t.door, (*t.c).Serialize())
}

func DeserializeTile(t string) Tile {

	if len(t) == 0 || t[0] != '{' {

		return Tile{}
	}

	b := 0
	e := len(t)
	for i, c := range t {
		if c == 'I' && b == 0 {
			b = i
		}
		if c == '}' && e == len(t) {
			e = i
		}
	}

	e++
	tile := Tile{}

	tile.terrain = icon.Deserialize(t[b:e])

	t = t[(e + 1):]
	restCreature := strings.Split(t, "Player")
	if len(restCreature) == 2 {
		tile.c = creature.Deserialize(restCreature[1])
	}
	t = restCreature[0]
	fields := strings.Split(t, " ")

	tile.x, _ = strconv.Atoi(fields[0])
	tile.y, _ = strconv.Atoi(fields[1])
	tile.passable, _ = strconv.ParseBool(fields[2])
	tile.door, _ = strconv.ParseBool(fields[3])

	return tile
}
func (t Tile) render(x, y int) {

	if t.c != nil {
		(*t.c).Render(x, y)
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
	c        *creature.Creature
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

func DeserializeViewer(v string) *Viewer {
	v = v[6 : len(v)-1]
	fields := strings.Split(v, " ")
	viewer := new(Viewer)
	viewer.x, _ = strconv.Atoi(fields[0])
	viewer.y, _ = strconv.Atoi(fields[1])
	viewer.width, _ = strconv.Atoi(fields[2])
	viewer.height, _ = strconv.Atoi(fields[3])
	return viewer
}
func (v *Viewer) Serialize() string {
	return fmt.Sprintf("Viewer{%d %d %d %d}", v.x, v.y, v.width, v.height)
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

func DeserializeMap(m string) Map {
	dimensionEntries := strings.Split(m, "\n")
	dimensions := strings.Split(dimensionEntries[0], " ")
	dimensionEntries = dimensionEntries[1:len(dimensionEntries)]
	height, _ := strconv.Atoi(dimensions[1])
	width, _ := strconv.Atoi(dimensions[0])
	grid := make([][]Tile, height)
	for i := 0; i < height; i++ {
		row := make([]Tile, width)
		tiles := strings.Split(dimensionEntries[i], "Tile")
		tiles = tiles[1:len(tiles)]
		for j := 0; j < width; j++ {
			row[j] = DeserializeTile(tiles[j])

		}
		grid[i] = row

	}
	return Map{grid, DeserializeViewer(dimensionEntries[len(dimensionEntries)-1])}
}
func (m Map) Serialize() string {
	result := fmt.Sprintf("%d %d\n", len(m.grid[0]), len(m.grid))
	for _, row := range m.grid {
		for _, tile := range row {
			result += tile.Serialize()
		}
		result += "\n"

	}
	result += m.v.Serialize() + "\n"
	return result
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

func (m Map) MovePlayer(c *creature.Player, x, y int) {
	m.MoveCreature(c, x, y)
	cX, cY := c.GetCoordinates()
	rX := cX - m.v.x
	rY := cY - m.v.y

	if rX < padding && cX >= padding {
		m.v.x--
	}
	if rX > m.v.width-padding && cX <= len(m.grid[0])-padding {
		m.v.x++
	}
	if rY < padding && cY >= padding {
		m.v.y--
	}
	if rY > m.v.height-padding && cY <= len(m.grid)-padding {
		m.v.y++
	}
}
func (m Map) MoveCreature(c creature.Creature, x, y int) {

	if !m.grid[y][x].passable {
		return
	}

	cX, cY := c.GetCoordinates()
	m.grid[cY][cX].c = nil
	cX = x
	cY = y
	c.SetCoordinates(cX, cY)
	m.grid[cY][cX].c = &c

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
