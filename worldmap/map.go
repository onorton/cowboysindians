package worldmap

import (
	"fmt"
	"strconv"
	"strings"

	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/creature"
	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/ui"
)

const padding = 5

func NewMap(width, height, viewerWidth, viewerHeight int) Map {
	grid := make([][]Tile, height)
	for i := 0; i < height; i++ {
		row := make([]Tile, width)
		for j := 0; j < width; j++ {
			row[j] = newTile("ground", j, i)

		}
		grid[i] = row
	}

	grid[0][4] = newTile("wall", 4, 0)
	grid[0][5] = newTile("wall", 5, 0)
	grid[0][6] = newTile("wall", 6, 0)
	grid[1][4] = newTile("door", 4, 1)
	grid[2][4] = newTile("wall", 4, 2)
	grid[2][5] = newTile("wall", 5, 2)
	grid[2][6] = newTile("wall", 6, 2)
	grid[1][6] = newTile("wall", 6, 1)
	grid[2][2].PlaceItem(item.NewItem("gem"))
	grid[2][3].PlaceItem(item.NewItem("gem"))
	grid[3][3].PlaceItem(item.NewItem("gem"))
	grid[7][1].PlaceItem(item.NewItem("gem"))
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
	result := fmt.Sprintf("%d %d\n", m.GetWidth(), m.GetHeight())
	for _, row := range m.grid {
		for _, tile := range row {
			result += tile.Serialize()
		}
		result += "\n"

	}
	result += m.v.Serialize() + "\n"
	return result
}

func (m Map) HasPlayer(x, y int) bool {
	if m.IsOccupied(x, y) {
		_, ok := m.grid[y][x].c.(*creature.Player)
		return ok
	}
	return false
}

// Coordinates within confines of the map
func (m Map) IsValid(x, y int) bool {
	return x >= 0 && x < m.GetWidth() && y >= 0 && y < m.GetHeight()

}

func (m Map) IsPassable(x, y int) bool {
	return m.grid[y][x].passable
}

func (m Map) IsOccupied(x, y int) bool {
	return m.grid[y][x].c != nil
}

func (m Map) HasItems(x, y int) bool {
	return len(m.grid[y][x].items) > 0
}

// Bresenham algorithm to check if creature c can see square x1, y1.
func (m Map) IsVisible(c creature.Creature, x1, y1 int) bool {
	x0, y0 := c.GetCoordinates()
	var xStep, yStep int
	x, y := x0, y0
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	if dy < 0 {
		yStep = -1
		dy *= -1
	} else if dy > 0 {
		yStep = 1
	}
	if dx < 0 {
		xStep = -1
		dx *= -1
	} else if dx > 0 {
		xStep = 1
	}

	// Go down longest delta
	if dx >= dy {
		dErr := dy / dx
		e := dErr - 0.5
		for i := 0; i < int(dx); i++ {
			x += xStep
			e += dErr

			if e >= 0.5 {
				y += yStep
				e -= 1
			}
			// If any square along path is impassable, x1, y1 is invisible
			if m.IsValid(x, y) && !(x == x1 && y == y1) && !m.IsPassable(x, y) {
				return false
			}
		}
	} else {
		dErr := dx / dy
		e := dErr - 0.5
		for i := 0; i < int(dy); i++ {
			y += yStep
			e += dErr
			if e >= 0.5 {
				x += xStep
				e -= 1
			}
			// If any square along path is impassable, x1, y1 is invisible
			if m.IsValid(x, y) && !(x == x1 && y == y1) && !m.IsPassable(x, y) {
				return false
			}

		}
	}

	return true
}

func (m Map) ToggleDoor(x, y int, open bool) bool {
	message.PrintMessage("Which direction?")
	height := m.GetHeight()
	width := m.GetWidth()
	// Select direction
	for {
		validMove := true
		action := ui.GetInput()

		if action.IsMovementAction() {
			switch action {
			case ui.MoveWest:
				if x != 0 {
					x--
				}
			case ui.MoveEast:
				if x < width-1 {
					x++
				}
			case ui.MoveNorth:
				if y != 0 {
					y--
				}
			case ui.MoveSouth:
				if y < height-1 {
					y++
				}
			case ui.MoveSouthWest:
				if x != 0 && y < height-1 {
					x--
					y++
				}

			case ui.MoveSouthEast:
				if x < width-1 && y < height-1 {
					x++
					y++
				}
			case ui.MoveNorthWest:
				if x != 0 && y != 0 {
					x--
					y--
				}
			case ui.MoveNorthEast:
				if y != 0 && x < width-1 {
					y--
					x++
				}
			}
		} else if action == ui.CancelAction {
			message.PrintMessage("Never mind...")
			return false
		} else {
			message.PrintMessage("Invalid direction.")
			validMove = false
		}

		if validMove {
			break
		}
	}
	// If there is a door, toggle its position if it's not already there
	if m.grid[y][x].door {
		if m.grid[y][x].passable != open {
			m.grid[y][x].passable = open
			if open {
				message.Enqueue("The door opens.")
			} else {
				message.Enqueue("The door closes.")
			}
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

func (m Map) PlaceItem(x, y int, item item.Item) {
	m.grid[y][x].PlaceItem(item)
}

// Interface for player to find a ranged target.
func (m Map) FindTarget(p *creature.Player) creature.Creature {
	if !p.Ranged() {
		message.PrintMessage("You are not wielding a ranged weapon.")
		return nil
	}

	if !p.HasAmmo() {
		message.PrintMessage("You have no ammunition for this weapon.")
		return nil
	}
	x, y := p.GetCoordinates()
	// In terms of viewer space rather than world space
	rX, rY := x-m.v.x, y-m.v.y
	width, height := m.GetWidth(), m.GetHeight()
	vWidth, vHeight := m.v.width, m.v.height
	for {
		message.PrintMessage("Select target")
		termbox.SetCell(rX, rY, 'X', termbox.ColorYellow, termbox.ColorDefault)
		termbox.Flush()
		x, y = m.v.x+rX, m.v.y+rY
		m.grid[y][x].render(rX, rY)
		action := ui.GetInput()
		if action.IsMovementAction() {
			switch action {
			case ui.MoveWest:
				if rX != 0 && x != 0 {
					rX--
				}
			case ui.MoveEast:
				if rX < vWidth-1 && x < width-1 {
					rX++
				}
			case ui.MoveNorth:
				if rY != 0 && y != 0 {
					rY--
				}
			case ui.MoveSouth:
				if rY < vHeight-1 && y < height-1 {
					rY++
				}
			case ui.MoveSouthWest:
				if rX != 0 && rY < vHeight-1 && x != 0 && y < height-1 {
					rX--
					rY++
				}
			case ui.MoveSouthEast:
				if rX < vWidth-1 && rY < vHeight-1 && x < width-1 && y < height-1 {
					rX++
					rY++
				}
			case ui.MoveNorthWest:
				if rX != 0 && rY != 0 && x != 0 && y != 0 {
					rX--
					rY--
				}
			case ui.MoveNorthEast:
				if rY != 0 && rX < vWidth-1 && y != 0 && x < width-1 {
					rY--
					rX++
				}
			}
		} else if action == ui.CancelAction { // Counter intuitive at the moment
			if m.IsOccupied(x, y) {
				// If a creature is there, return it.
				return m.grid[y][x].c
			} else {
				message.PrintMessage("Never mind...")
				return nil
			}

		}
	}
}

func (m Map) GetWidth() int {
	return len(m.grid[0])
}

func (m Map) GetHeight() int {
	return len(m.grid)
}

// Same as MoveCreature but viewer is adjusted as well.
func (m Map) MovePlayer(c *creature.Player, action ui.PlayerAction) {

	x, y := c.GetCoordinates()

	newX, newY := x, y

	switch action {
	case ui.MoveWest:
		newX--
	case ui.MoveEast:
		newX++
	case ui.MoveNorth:
		newY--
	case ui.MoveSouth:
		newY++
	case ui.MoveSouthWest:
		newX--
		newY++
	case ui.MoveSouthEast:
		newX++
		newY++
	case ui.MoveNorthWest:
		newX--
		newY--
	case ui.MoveNorthEast:
		newY--
		newX++
	}

	// If out of bounds, reset to original position
	if newX < 0 || newY < 0 || newX >= m.GetWidth() || newY >= m.GetHeight() {
		newX, newY = x, y
	}

	m.MoveCreature(c, newX, newY)

	// Difference in coordinates from the window location
	rX := newX - m.v.x
	rY := newY - m.v.y

	//Adjust viewer
	if rX < padding && newX >= padding {
		m.v.x--
	}
	if rX > m.v.width-padding && newX <= m.GetWidth()-padding {
		m.v.x++
	}
	if rY < padding && newY >= padding {
		m.v.y--
	}
	if rY > m.v.height-padding && newY <= m.GetHeight()-padding {
		m.v.y++
	}
}

func (m Map) MoveCreature(c creature.Creature, x, y int) {

	if !m.grid[y][x].passable {
		return
	}
	// If occupied by another creature, melee attack
	if m.grid[y][x].c != nil && m.grid[y][x].c != c {
		c.MeleeAttack(m.grid[y][x].c)
		return
	}

	cX, cY := c.GetCoordinates()
	m.grid[cY][cX].c = nil
	cX = x
	cY = y
	c.SetCoordinates(cX, cY)
	m.grid[cY][cX].c = c
}

func (m Map) PickupItem() bool {
	player := m.GetPlayer()
	x, y := player.GetCoordinates()
	if m.grid[y][x].items == nil {
		message.PrintMessage("There is no item here.")
		return false
	}

	items := make(map[rune]([]item.Item))
	for _, itm := range m.GetItems(x, y) {
		existing := items[itm.GetKey()]
		if existing == nil {
			existing = make([]item.Item, 0)
		}
		existing = append(existing, itm)
		items[itm.GetKey()] = existing
	}
	for k := range items {
		for _, item := range items[k] {
			player.PickupItem(item)
		}
		message.Enqueue(fmt.Sprintf("You pick up %d %ss.", len(items[k]), items[k][0].GetName()))

	}
	return true
}

func (m Map) GetItems(x, y int) []item.Item {
	items := m.grid[y][x].items
	m.grid[y][x].items = make([]item.Item, 0)
	return items
}
func (m Map) DropItem() bool {
	player := m.GetPlayer()
	x, y := player.GetCoordinates()
	for {
		message.PrintMessage(fmt.Sprintf("What do you want to drop? [%s or *]", player.GetInventoryKeys()))
		s, c := ui.GetItemSelection()

		switch s {
		case ui.All:
			player.PrintInventory()
			continue
		case ui.Cancel:
			message.PrintMessage("Never mind.")
			return false
		case ui.SpecificItem:
			item := player.GetItem(c)
			if item == nil {
				message.PrintMessage("You don't have that item.")
				ui.GetInput()
			} else {
				m.grid[y][x].PlaceItem(item)
				message.Enqueue(fmt.Sprintf("You dropped a %s.", item.GetName()))
				return true
			}
		// Not selectable but still need to consider it
		case ui.AllRelevant:
			return false
		}
	}
}

func (m Map) GetPlayer() *creature.Player {
	for _, row := range m.grid {
		for _, tile := range row {
			if tile.c == nil {
				continue
			}
			player, ok := tile.c.(*creature.Player)
			if ok {
				return player
			}
		}
	}
	return nil
}

func (m Map) DeleteCreature(c creature.Creature) {
	x, y := c.GetCoordinates()
	m.grid[y][x].c = nil
}

func (m Map) Render() {
	player := m.GetPlayer()
	for y, row := range m.grid {
		for x, tile := range row {
			rX := x - m.v.x
			rY := y - m.v.y
			if rX >= 0 && rX < m.v.width && rY >= 0 && rY < m.v.height {
				if m.IsVisible(player, x, y) {
					tile.render(rX, rY)
				} else {
					termbox.SetCell(rX, rY, ' ', termbox.ColorDefault, termbox.ColorDefault)
				}
			}
		}
	}
	termbox.Flush()
}
