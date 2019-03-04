package worldmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/ui"
)

type Alignment int

const (
	Player Alignment = iota
	Enemy
	Neutral
)

const padding = 5

func NewMap(width, height, viewerWidth, viewerHeight int) *Map {

	grid := generateMap(width, height)

	viewer := new(Viewer)
	viewer.x = 0
	viewer.y = 0
	viewer.width = viewerWidth
	viewer.height = viewerHeight
	return &Map{grid, viewer}
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

func (v *Viewer) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	xValue, err := json.Marshal(v.x)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"X\":%s,", xValue))

	yValue, err := json.Marshal(v.y)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Y\":%s,", yValue))

	widthValue, err := json.Marshal(v.width)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"width\":%s,", widthValue))

	heightValue, err := json.Marshal(v.height)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"height\":%s", heightValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (v *Viewer) UnmarshalJSON(data []byte) error {
	type viewerJson struct {
		X      int
		Y      int
		Width  int
		Height int
	}

	value := viewerJson{}

	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	v.x = value.X
	v.y = value.Y
	v.width = value.Width
	v.height = value.Height

	return nil
}

func (m *Map) GetViewerX() int {
	return m.v.x
}

func (m *Map) GetViewerY() int {
	return m.v.y
}

func (m *Map) GetViewerWidth() int {
	return m.v.width
}

func (m *Map) GetViewerHeight() int {
	return m.v.height
}

func (m *Map) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	gridValue, err := json.Marshal(m.grid)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Map\":%s,", gridValue))

	viewerValue, err := json.Marshal(m.v)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Viewer\":%s", viewerValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (m *Map) UnmarshalJSON(data []byte) error {
	type mapJson struct {
		Map    [][]Tile
		Viewer *Viewer
	}

	v := mapJson{}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	m.grid = v.Map
	m.v = v.Viewer

	return nil
}

func (m Map) HasPlayer(x, y int) bool {
	if m.IsOccupied(x, y) {
		return m.grid[y][x].c.GetAlignment() == Player
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

func (m Map) blocksVision(x, y int) bool {
	return m.grid[y][x].blocksV
}

func (m Map) IsOccupied(x, y int) bool {
	return m.grid[y][x].c != nil
}

func (m Map) HasItems(x, y int) bool {
	return len(m.grid[y][x].items) > 0
}

// Bresenham algorithm to check if creature c can see square x1, y1.
func (m Map) IsVisible(c Creature, x1, y1 int) bool {
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
			// If any square along path blocks vision, target square is invisible
			if m.IsValid(x, y) && !(x == x1 && y == y1) && m.blocksVision(x, y) {
				return false
			}

			// If square in path gives cover, is adjacent to the target square and c is crouching then target square is invisible
			if m.IsValid(x, y) && m.givesCover(x, y) && m.isAdjacent(x, y, x1, y1) && c.IsCrouching() {
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
			// If any square along path blocks vision, target square is invisible
			if m.IsValid(x, y) && !(x == x1 && y == y1) && m.blocksVision(x, y) {
				return false
			}

			// If square in path gives cover, is adjacent to the target square and c is crouching then target square is invisible
			if m.IsValid(x, y) && m.givesCover(x, y) && m.isAdjacent(x, y, x1, y1) && c.IsCrouching() {
				return false
			}

		}
	}

	return true
}

func (m Map) TargetBehindCover(a, t Creature) bool {
	x0, y0 := a.GetCoordinates()
	x1, y1 := t.GetCoordinates()
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
			// If any square along path is impassable, target square is behind cover
			if m.IsValid(x, y) && !(x == x1 && y == y1) && !m.IsPassable(x, y) {
				return true
			}

			// If square in path gives cover, is adjacent to the target square and target is crouching then target is behind cover
			if m.IsValid(x, y) && m.givesCover(x, y) && m.isAdjacent(x, y, x1, y1) && t.IsCrouching() {
				return true
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
			// If any square along path is impassable, target square is behind cover
			if m.IsValid(x, y) && !(x == x1 && y == y1) && !m.IsPassable(x, y) {
				return false
			}

			// If square in path gives cover, is adjacent to the target square and target is crouching then target is behind cover
			if m.IsValid(x, y) && m.givesCover(x, y) && m.isAdjacent(x, y, x1, y1) && t.IsCrouching() {
				return true
			}
		}
	}
	return false
}

func (m Map) BehindCover(x1, y1 int, a Creature) bool {
	x0, y0 := a.GetCoordinates()
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
			// If any square along path is impassable, target square is behind cover
			if m.IsValid(x, y) && !(x == x1 && y == y1) && !m.IsPassable(x, y) {
				return true
			}

			// If square in path gives cover, is adjacent to the target square then target square would be behind cover
			if m.IsValid(x, y) && m.givesCover(x, y) && m.isAdjacent(x, y, x1, y1) {
				return true
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
			// If any square along path is impassable, target square is behind cover
			if m.IsValid(x, y) && !(x == x1 && y == y1) && !m.IsPassable(x, y) {
				return true
			}

			// If square in path gives cover, is adjacent to the target square then target square would be behind cover
			if m.IsValid(x, y) && m.givesCover(x, y) && m.isAdjacent(x, y, x1, y1) {
				return true
			}
		}
	}
	return false
}

func (m Map) PlaceItem(x, y int, item item.Item) {
	m.grid[y][x].PlaceItem(item)
}

func (m Map) GetWidth() int {
	return len(m.grid[0])
}

func (m Map) GetHeight() int {
	return len(m.grid)
}

// Same as MoveCreature but viewer is adjusted as well.
func (m Map) MovePlayer(c Creature, action ui.PlayerAction) {

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

func (m Map) MoveCreature(c Creature, x, y int) {

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

func (m Map) GetItems(x, y int) []item.Item {
	items := m.grid[y][x].items
	m.grid[y][x].items = make([]item.Item, 0)
	return items
}

func (m Map) GetPlayer() Creature {
	for _, row := range m.grid {
		for _, tile := range row {
			if tile.c == nil {
				continue
			}
			if tile.c.GetAlignment() == Player {
				return tile.c
			}
		}
	}
	return nil
}

func (m Map) GetCreature(x, y int) Creature {
	return m.grid[y][x].c
}

func (m Map) RenderTile(x, y int) ui.Element {
	return m.grid[y][x].render()
}

func (m Map) DeleteCreature(c Creature) {
	x, y := c.GetCoordinates()
	m.grid[y][x].c = nil
}

func (m Map) Render() {
	player := m.GetPlayer()

	elems := make([][]ui.Element, m.v.height, m.v.height)

	for i, _ := range elems {
		if i >= m.v.height {
			log.Panic("Sup")
		}
		elems[i] = make([]ui.Element, m.v.width, m.v.width)
	}

	for y, row := range m.grid {
		for x, tile := range row {
			rX := x - m.v.x
			rY := y - m.v.y
			if rX >= 0 && rX < m.v.width && rY >= 0 && rY < m.v.height {
				if m.IsVisible(player, x, y) {
					elems[rY][rX] = tile.render()
				} else {
					elems[rY][rX] = ui.EmptyElement()
				}
			}
		}
	}
	ui.RenderGrid(0, 0, elems)
}

func (m *Map) IsDoor(x, y int) bool {
	return m.grid[y][x].door
}

func (m *Map) GetPassable(x, y int) bool {
	return m.grid[y][x].passable
}

func (m *Map) SetPassable(x, y int, passable bool) {
	m.grid[y][x].passable = passable
}

func (m *Map) SetBlocksVision(x, y int, blocksV bool) {
	m.grid[y][x].blocksV = blocksV
}

func (m *Map) givesCover(x, y int) bool {
	return m.grid[y][x].givesCover()
}

func (m *Map) isAdjacent(x1, y1, x2, y2 int) bool {
	if x1 == x2 && y1 == y2 {
		return false
	}
	return math.Abs(float64(x1-x2)) <= 1 && math.Abs(float64(y1-y2)) <= 1
}

func GetBonus(score int) int {
	return (score - 10) / 2
}

// Interface shared by Player and Enemy
type Creature interface {
	GetCoordinates() (int, int)
	SetCoordinates(int, int)
	Render() ui.Element
	GetInitiative() int
	MeleeAttack(Creature)
	TakeDamage(int)
	IsDead() bool
	IsCrouching() bool
	AttackHits(int) bool
	GetName() string
	GetAlignment() Alignment
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
}
