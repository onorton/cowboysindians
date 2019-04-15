package worldmap

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func NewMap(grid [][]Tile, viewerWidth, viewerHeight int) *Map {

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
		Map    [][](map[string]interface{})
		Viewer *Viewer
	}

	v := mapJson{}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	width := len(v.Map[0])
	height := len(v.Map)

	grid := make([][]Tile, width)

	for y := 0; y < height; y++ {
		grid[y] = make([]Tile, width)
		for x := 0; x < width; x++ {
			grid[y][x] = unmarshalTile(v.Map[y][x])
		}
	}
	m.grid = grid

	m.v = v.Viewer

	return nil
}

func (m Map) HasPlayer(x, y int) bool {
	if m.IsOccupied(x, y) {
		return m.grid[y][x].getCreature().GetAlignment() == Player
	}
	return false
}

// Coordinates within confines of the map
func (m Map) IsValid(x, y int) bool {
	return x >= 0 && x < m.GetWidth() && y >= 0 && y < m.GetHeight()

}

func (m Map) IsPassable(x, y int) bool {
	return m.grid[y][x].isPassable()
}

func (m Map) blocksVision(x, y int) bool {
	return m.grid[y][x].blocksVision()
}

func (m Map) IsOccupied(x, y int) bool {
	return m.grid[y][x].isOccupied()
}

func (m Map) HasItems(x, y int) bool {
	return m.grid[y][x].hasItems()
}

// Bresenham algorithm to check if creature c can see square x1, y1.
func (m Map) IsVisible(c CanSee, x1, y1 int) bool {
	x0, y0 := c.GetCoordinates()
	distance := math.Sqrt(math.Pow(float64(x1-x0), 2) + math.Pow(float64(y1-y0), 2))
	if distance > float64(c.GetVisionDistance()) {
		return false
	}

	crouching := false
	if canCrouch, ok := c.(CanCrouch); ok {
		crouching = canCrouch.IsCrouching()
	}

	// If square adjacent, it is visible
	if math.Abs(float64(x1-x0)) <= 1 && math.Abs(float64(x1-x0)) <= 1 {
		return true
	}

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
			if m.IsValid(x, y) && m.givesCover(x, y) && m.isAdjacent(x, y, x1, y1) && crouching {
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
			if m.IsValid(x, y) && m.givesCover(x, y) && m.isAdjacent(x, y, x1, y1) && crouching {
				return false
			}

		}
	}

	return true
}

// Bresenham algorithm to check if creature c can talk to t
func (m Map) InConversationRange(c, t Creature) bool {

	x0, y0 := c.GetCoordinates()
	x1, y1 := t.GetCoordinates()

	// No point talking if they cannot see other
	if !m.IsVisible(c, x1, y1) || !m.IsVisible(t, x0, y0) {
		return false
	}

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

			// If any square along path is impassable, c cannot talk to t
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

			// If any square along path is impassable, c cannot talk to t
			if m.IsValid(x, y) && !(x == x1 && y == y1) && !m.IsPassable(x, y) {
				return false
			}

		}
	}

	return true
}

func (m Map) TargetBehindCover(a hasPosition, t Creature) bool {
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
	m.grid[y][x].placeItem(item)
}

func (m Map) GetWidth() int {
	return len(m.grid[0])
}

func (m Map) GetHeight() int {
	return len(m.grid)
}

// Adjust the viewer according to the new position of the player
func (m Map) AdjustViewer() {
	x, y := m.GetPlayer().GetCoordinates()
	// Difference in coordinates from the window location
	rX := x - m.v.x
	rY := y - m.v.y

	//Adjust viewer
	if rX < padding && x >= padding {
		m.v.x--
	}
	if rX > m.v.width-padding && x <= m.GetWidth()-padding {
		m.v.x++
	}
	if rY < padding && y >= padding {
		m.v.y--
	}
	if rY > m.v.height-padding && y <= m.GetHeight()-padding {
		m.v.y++
	}
}

func (m Map) MoveCreature(c Creature, x, y int) {

	// If occupied by another creature, melee attack
	if m.grid[y][x].isOccupied() && m.grid[y][x].getCreature() != c {
		c.MeleeAttack(m.grid[y][x].getCreature())
		return
	}

	m.Move(c, x, y)
}

func (m Map) Move(c Creature, x, y int) {

	if !m.grid[y][x].isPassable() {
		return
	}

	cX, cY := c.GetCoordinates()
	m.grid[cY][cX].setCreature(nil)
	cX = x
	cY = y
	c.SetCoordinates(cX, cY)
	m.grid[cY][cX].setCreature(c)
}

func (m Map) GetItems(x, y int) []item.Item {
	return m.grid[y][x].getItems()
}

func (m Map) GetPlayer() Creature {
	for _, row := range m.grid {
		for _, tile := range row {
			if !tile.isOccupied() {
				continue
			}
			if tile.getCreature().GetAlignment() == Player {
				return tile.getCreature()
			}
		}
	}
	return nil
}

func (m Map) GetCreature(x, y int) Creature {
	return m.grid[y][x].getCreature()
}

func (m Map) RenderTile(x, y int) ui.Element {
	return m.grid[y][x].render()
}

func (m Map) DeleteCreature(c Creature) {
	x, y := c.GetCoordinates()
	m.grid[y][x].setCreature(nil)
}

func (m Map) Render() {
	player := m.GetPlayer()

	elems := make([][]ui.Element, m.v.height, m.v.height)

	for i, _ := range elems {
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
	_, ok := m.grid[y][x].(*Door)
	return ok
}

func (m *Map) GetPassable(x, y int) bool {
	return m.grid[y][x].isPassable()
}

func (m *Map) ToggleDoor(x, y int, open bool) {
	if d, ok := m.grid[y][x].(*Door); ok {
		if open {
			d.passable = true
			d.blocksV = false
		} else {
			d.passable = false
			d.blocksV = d.blocksVClosed
		}
	}
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

type Coordinates struct {
	X int
	Y int
}

// Interface shared by Player, Npc and Enemy
type Creature interface {
	CanSee
	Render() ui.Element
	GetInitiative() int
	MeleeAttack(Creature)
	TakeDamage(int)
	IsDead() bool
	IsCrouching() bool
	AttackHits(int) bool
	GetName() ui.Name
	GetAlignment() Alignment
	Update()
	GetID() string
}

type hasPosition interface {
	GetCoordinates() (int, int)
	SetCoordinates(int, int)
}

type CanSee interface {
	GetVisionDistance() int
	hasPosition
}

type CanCrouch interface {
	IsCrouching() bool
	Standup()
	Crouch()
}
