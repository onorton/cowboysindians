package worldmap

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/ui"
)

const chunkSize = 64
const padding = 5

type Alignment int

type Map struct {
	activeChunks [3][3]*Grid
	filename     string
	v            *Viewer
	width        int
	height       int
	player       Creature
	creatures    []Creature
}

type Coordinates struct {
	X int
	Y int
}

type ChunkCoordinates struct {
	ChunkX int
	ChunkY int
	Local  Coordinates
}

func NewMap(filename string, width, height int, viewer *Viewer, player Creature, creatures []Creature) *Map {
	newMap := new(Map)
	newMap.v = viewer
	newMap.width = width
	newMap.height = height
	newMap.filename = filename
	newMap.player = player
	newMap.creatures = creatures
	for _, c := range creatures {
		c.SetMap(newMap)
	}
	return newMap
}

func (m *Map) LoadActiveChunks() {

	m.activeChunks = [3][3]*Grid{}
	pX, pY := m.player.GetCoordinates()
	newPlayerLocation := globalToChunkCoordinates(pX, pY)

	file, err := os.Open(m.filename)
	check(err)
	defer file.Close()

	reader := bufio.NewReader(file)
	// Determine what the line numbers are for the chunks
	lineToChunks := map[int]Coordinates{}
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			index := m.verticalChunks()*(newPlayerLocation.ChunkY+y-1) + newPlayerLocation.ChunkX + x - 1
			if index >= 0 && index < m.verticalChunks()*m.horizontalChunks() {
				lineToChunks[index] = Coordinates{x, y}
			}
		}
	}
	currentLine := 1
	reader.ReadString('\n')
	for {
		line, err := reader.ReadString('\n')
		if coordinates, ok := lineToChunks[currentLine-1]; ok {
			var chunk Grid
			err := json.Unmarshal([]byte(strings.Trim(line, ",\n")), &chunk)
			check(err)
			m.activeChunks[coordinates.Y][coordinates.X] = &chunk
		}

		currentLine++
		if err != nil {
			break
		}
	}

	// Place creatures
	for _, c := range m.creatures {
		x, y := c.GetCoordinates()
		if m.InActiveChunks(x, y) {
			m.Move(c, x, y)
		}
	}
}

func (m *Map) SaveChunks() {
	pX, pY := m.player.GetCoordinates()
	newPlayerLocation := globalToChunkCoordinates(pX, pY)

	file, err := os.Open(m.filename)
	check(err)
	outputFile, err := os.Create("tmp_" + m.filename)
	check(err)
	defer file.Close()
	defer outputFile.Close()

	reader := bufio.NewReader(file)
	writer := bufio.NewWriter(outputFile)
	// Determine what the line numbers are for the chunks

	lineToChunks := map[int]Coordinates{}
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			index := m.verticalChunks()*(newPlayerLocation.ChunkY+y-1) + newPlayerLocation.ChunkX + x - 1
			if index >= 0 && index < m.verticalChunks()*m.horizontalChunks() {
				lineToChunks[index] = Coordinates{x, y}
			}
		}
	}
	currentLine := 1
	firstLine, err := reader.ReadString('\n')
	check(err)
	writer.WriteString(firstLine)
	for {
		line, err := reader.ReadString('\n')
		if coordinates, ok := lineToChunks[currentLine-1]; ok {
			chunkData, err := m.activeChunks[coordinates.Y][coordinates.X].MarshalJSON()
			check(err)
			writer.Write(chunkData)
			writer.WriteString("\n")
		} else {
			writer.WriteString(line)
		}

		currentLine++
		if err != nil {
			break
		}
	}
	err = writer.Flush()
	check(err)
	err = os.Rename("tmp_"+m.filename, m.filename)
	check(err)
}

type Viewer struct {
	x      int
	y      int
	width  int
	height int
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

func NewViewer(x, y, viewerWidth, viewerHeight int) *Viewer {
	return &Viewer{x, y, viewerWidth, viewerHeight}
}

func (m Map) GetViewerX() int {
	return m.v.x
}

func (m Map) GetViewerY() int {
	return m.v.y
}

func (m Map) GetViewerWidth() int {
	return m.v.width
}

func (m Map) GetViewerHeight() int {
	return m.v.height
}

func (m Map) HasPlayer(x, y int) bool {
	if m.IsOccupied(x, y) {
		return m.GetCreature(x, y).GetAlignment() == Player
	}
	return false
}

// Coordinates within confines of the map
func (m Map) IsValid(x, y int) bool {
	inWorld := x >= 0 && x < m.width && y >= 0 && y < m.height
	chunk, _, _ := m.globalToChunkAndLocal(x, y)
	return inWorld && chunk != nil
}

func (m Map) IsPassable(x, y int) bool {
	if !m.InActiveChunks(x, y) {
		return false
	}
	chunk, cX, cY := m.globalToChunkAndLocal(x, y)
	return chunk.passable[cY][cX]
}

func (m Map) blocksVision(x, y int) bool {
	chunk, cX, cY := m.globalToChunkAndLocal(x, y)
	return chunk.blocksVision[cY][cX]
}

func (m Map) IsOccupied(x, y int) bool {
	return m.GetCreature(x, y) != nil
}

func (m Map) HasItems(x, y int) bool {
	chunk, cX, cY := m.globalToChunkAndLocal(x, y)
	return len(chunk.items[cY][cX]) > 0
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
			if m.IsValid(x, y) && isAdjacent(x, y, x1, y1) && crouching && m.givesCover(x, y) {
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
			if m.IsValid(x, y) && isAdjacent(x, y, x1, y1) && crouching && m.givesCover(x, y) {
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
			if m.IsValid(x, y) && isAdjacent(x, y, x1, y1) && t.IsCrouching() && m.givesCover(x, y) {
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
			if m.IsValid(x, y) && isAdjacent(x, y, x1, y1) && t.IsCrouching() && m.givesCover(x, y) {
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
			if m.IsValid(x, y) && isAdjacent(x, y, x1, y1) && m.givesCover(x, y) {
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
			if m.IsValid(x, y) && isAdjacent(x, y, x1, y1) && m.givesCover(x, y) {
				return true
			}
		}
	}
	return false
}

func (m *Map) PlaceItem(x, y int, itm *item.Item) {
	chunk, cX, cY := m.globalToChunkAndLocal(x, y)
	chunk.items[cY][cX] = append([]*item.Item{itm}, chunk.items[cY][cX]...)
}

func (m Map) GetWidth() int {
	return m.width
}

func (m Map) GetHeight() int {
	return m.height
}

func (m Map) horizontalChunks() int {
	return m.width / chunkSize
}

func (m Map) verticalChunks() int {
	return m.height / chunkSize
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

func (m *Map) MovePlayer(player Creature, x, y int) {
	oldX, oldY := player.GetCoordinates()

	oldChunkCoordinates, newChunkCoordinates := globalToChunkCoordinates(oldX, oldY), globalToChunkCoordinates(x, y)
	movingToNewChunk := oldChunkCoordinates.ChunkX != newChunkCoordinates.ChunkX || oldChunkCoordinates.ChunkY != newChunkCoordinates.ChunkY

	if movingToNewChunk {
		m.SaveChunks()
	}
	m.Move(player, x, y)
	// If player moves to a new chunk, reload chunks
	if movingToNewChunk {
		m.LoadActiveChunks()
	}

}

func (m Map) Move(c Creature, x, y int) {

	if !m.IsPassable(x, y) {
		return
	}

	oldX, oldY := c.GetCoordinates()
	oldChunk, cX, cY := m.globalToChunkAndLocal(oldX, oldY)
	oldChunk.c[cY][cX] = nil
	c.SetCoordinates(x, y)
	newChunk, newX, newY := m.globalToChunkAndLocal(x, y)
	newChunk.c[newY][newX] = c
}

func (m Map) GetItems(x, y int) []*item.Item {
	chunk, cX, cY := m.globalToChunkAndLocal(x, y)
	items := chunk.items[cY][cX]
	chunk.items[cY][cX] = make([]*item.Item, 0)
	return items
}

func (m Map) GetPlayer() Creature {
	return m.player
}

func (m Map) GetCreature(x, y int) Creature {
	chunk, cX, cY := m.globalToChunkAndLocal(x, y)
	return chunk.c[cY][cX]
}

func (m Map) CreatureById(id string) Creature {
	for _, c := range m.creatures {
		if c.GetID() == id {
			return c
		}
	}
	return nil
}

func (m *Map) RenderTile(x, y int) ui.Element {
	chunk, cX, cY := m.globalToChunkAndLocal(x, y)
	if chunk == nil {
		return ui.EmptyElement()
	}

	if m.GetCreature(x, y) != nil {
		return m.GetCreature(x, y).Render()
	} else if m.IsPassable(x, y) {
		if m.HasItems(x, y) {
			// pick an item that gives cover if it exists
			for _, item := range chunk.items[cY][cX] {
				if item.HasComponent("cover") {
					return item.Render()
				}
			}

			return chunk.items[cY][cX][0].Render()
		}

		if m.IsDoor(x, y) && m.Door(x, y).Open() {
			return terrainData["ground"].Icon.Render()
		}
	}
	return chunk.terrain[cY][cX].Render()
}

func (m Map) DeleteCreature(c Creature) {
	x, y := c.GetCoordinates()
	chunk, cX, cY := m.globalToChunkAndLocal(x, y)
	chunk.c[cY][cX] = nil
}

func (m Map) Render() {
	player := m.GetPlayer()

	elems := make([][]ui.Element, m.v.height, m.v.height)

	for i, _ := range elems {
		elems[i] = make([]ui.Element, m.v.width, m.v.width)
	}

	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
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

func (m Map) IsDoor(x, y int) bool {
	chunk, cX, cY := m.globalToChunkAndLocal(x, y)
	return chunk.door[cY][cX] != nil
}

func (m Map) Door(x, y int) *doorComponent {
	chunk, cX, cY := m.globalToChunkAndLocal(x, y)
	return chunk.door[cY][cX]
}

func (m Map) ToggleDoor(x, y int, open bool) {
	chunk, cX, cY := m.globalToChunkAndLocal(x, y)

	if chunk.door[cY][cX] != nil {
		if open && !chunk.door[cY][cX].locked {
			chunk.passable[cY][cX] = true
			chunk.blocksVision[cY][cX] = false
			chunk.door[cY][cX].open = true
		} else {
			chunk.passable[cY][cX] = false
			chunk.blocksVision[cY][cX] = chunk.door[cY][cX].blocksVClosed
			chunk.door[cY][cX].open = false
		}
	}

}

func (m Map) givesCover(x, y int) bool {
	chunk, cX, cY := m.globalToChunkAndLocal(x, y)
	cover := !chunk.passable[cY][cX]

	for _, item := range chunk.items[cY][cX] {
		cover = cover || item.HasComponent("cover")
	}
	return cover
}

func isAdjacent(x1, y1, x2, y2 int) bool {
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

func (m Map) InActiveChunks(x, y int) bool {
	return m.chunk(globalToChunkCoordinates(x, y)) != nil
}

func (m Map) chunk(location ChunkCoordinates) *Grid {
	pX, pY := m.player.GetCoordinates()
	playerChunkCoordinates := globalToChunkCoordinates(pX, pY)
	dX := location.ChunkX - playerChunkCoordinates.ChunkX
	dY := location.ChunkY - playerChunkCoordinates.ChunkY

	if math.Abs(float64(dX)) > 1 || math.Abs(float64(dY)) > 1 {
		return nil
	}
	return m.activeChunks[dY+1][dX+1]
}

func globalToChunkCoordinates(x, y int) ChunkCoordinates {
	return ChunkCoordinates{x / chunkSize, y / chunkSize, Coordinates{x % chunkSize, y % chunkSize}}
}

func (m Map) globalToChunkAndLocal(x, y int) (*Grid, int, int) {
	chunkCoordinates := globalToChunkCoordinates(x, y)
	chunk := m.chunk(chunkCoordinates)
	return chunk, chunkCoordinates.Local.X, chunkCoordinates.Local.Y
}
