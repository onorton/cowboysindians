package worldmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"

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

func NewMap(grid *Grid, viewerWidth, viewerHeight int) *Map {

	viewer := new(Viewer)
	viewer.x = 0
	viewer.y = 0
	viewer.width = viewerWidth
	viewer.height = viewerHeight
	return &Map{grid, viewer}
}

type Viewer struct {
	x      int
	y      int
	width  int
	height int
}

type Map struct {
	grid *Grid
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
		Map    *Grid
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
		return m.grid.c[y][x].GetAlignment() == Player
	}
	return false
}

// Coordinates within confines of the map
func (m Map) IsValid(x, y int) bool {
	return x >= 0 && x < m.GetWidth() && y >= 0 && y < m.GetHeight()

}

func (m Map) IsPassable(x, y int) bool {
	return m.grid.passable[y][x]
}

func (m Map) blocksVision(x, y int) bool {
	return m.grid.blocksVision[y][x]
}

func (m Map) IsOccupied(x, y int) bool {
	return m.grid.c[y][x] != nil
}

func (m Map) HasItems(x, y int) bool {
	return len(m.grid.items[y][x]) > 0
}

// Bresenham algorithm to check if creature c can see square x1, y1.
func (m Map) IsVisible(c CanSee, x1, y1 int) bool {
	x0, y0 := c.GetCoordinates()
	distance := Distance(x0, y0, x1, y1)
	if distance > float64(c.GetVisionDistance()) {
		return false
	}

	crouching := false
	if canCrouch, ok := c.(CanCrouch); ok {
		crouching = canCrouch.IsCrouching()
	}

	// If square adjacent, it is visible
	if math.Abs(float64(x1-x0)) <= 1 && math.Abs(float64(y1-y0)) <= 1 {
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

func (m Map) PlaceItem(x, y int, itm item.Item) {
	m.grid.items[y][x] = append([]item.Item{itm}, m.grid.items[y][x]...)
}

func (m Map) GetWidth() int {
	return m.grid.Width()
}

func (m Map) GetHeight() int {
	return m.grid.Height()
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
	if m.IsOccupied(x, y) && m.GetCreature(x, y) != c {
		c.MeleeAttack(m.GetCreature(x, y))
		return
	}

	m.Move(c, x, y)
}

func (m Map) Move(c Creature, x, y int) {

	if !m.IsPassable(x, y) {
		return
	}

	cX, cY := c.GetCoordinates()
	m.grid.c[cY][cX] = nil
	cX = x
	cY = y
	c.SetCoordinates(cX, cY)
	m.grid.c[cY][cX] = c
}

func (m Map) GetItems(x, y int) []item.Item {
	items := m.grid.items[y][x]
	m.grid.items[y][x] = make([]item.Item, 0)
	return items
}

func (m Map) GetPlayer() Creature {
	for _, row := range m.grid.c {
		for _, c := range row {
			if c == nil {
				continue
			}

			if c.GetAlignment() == Player {
				return c
			}

		}
	}
	return nil
}

func (m Map) GetCreature(x, y int) Creature {
	return m.grid.c[y][x]
}

func (m Map) RenderTile(x, y int) ui.Element {

	if m.GetCreature(x, y) != nil {
		return m.GetCreature(x, y).Render()
	} else if m.IsPassable(x, y) {
		if m.HasItems(x, y) {
			// pick an item that gives cover if it exists
			for _, item := range m.grid.items[y][x] {
				if item.GivesCover() {
					return item.Render()
				}
			}

			return m.grid.items[y][x][0].Render()
		}
		if m.IsDoor(x, y) {
			return terrainData["ground"].Icon.Render()
		}
	}
	return m.grid.terrain[y][x].Render()

}

func (m Map) DeleteCreature(c Creature) {
	x, y := c.GetCoordinates()
	m.grid.c[y][x] = nil
}

func (m Map) Render() {
	player := m.GetPlayer()

	elems := make([][]ui.Element, m.v.height, m.v.height)

	for i, _ := range elems {
		elems[i] = make([]ui.Element, m.v.width, m.v.width)
	}

	for y := 0; y < m.GetHeight(); y++ {
		for x := 0; x < m.GetWidth(); x++ {
			rX := x - m.v.x
			rY := y - m.v.y
			if rX >= 0 && rX < m.v.width && rY >= 0 && rY < m.v.height {
				if m.IsVisible(player, x, y) {
					elems[rY][rX] = m.RenderTile(x, y)
				} else {
					elems[rY][rX] = ui.EmptyElement()
				}
			}
		}
	}
	ui.RenderGrid(0, 0, elems)
}

func (m *Map) IsDoor(x, y int) bool {
	return m.grid.door[y][x]
}

func (m *Map) ToggleDoor(x, y int, open bool) {

	if m.grid.door[y][x] {
		if open {
			m.grid.passable[y][x] = true
			m.grid.blocksVision[y][x] = false
		} else {
			m.grid.passable[y][x] = false
			m.grid.blocksVision[y][x] = m.grid.blocksVClosed[y][x]
		}
	}
}

func (m *Map) givesCover(x, y int) bool {

	cover := !m.grid.passable[y][x]

	for _, item := range m.grid.items[y][x] {
		cover = cover || item.GivesCover()
	}
	return cover
}

func (m *Map) isAdjacent(x1, y1, x2, y2 int) bool {
	if x1 == x2 && y1 == y2 {
		return false
	}
	return math.Abs(float64(x1-x2)) <= 1 && math.Abs(float64(y1-y2)) <= 1
}

func Distance(x1, y1, x2, y2 int) float64 {
	return math.Sqrt(float64((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1)))
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
