package world

import (
	"math"
	"math/rand"

	"github.com/onorton/cowboysindians/enemy"
	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/mount"
	"github.com/onorton/cowboysindians/npc"
	"github.com/onorton/cowboysindians/worldmap"
)

type building struct {
	x1 int
	y1 int
	x2 int
	y2 int
}

type town struct {
	tX1        int
	tY1        int
	tX2        int
	tY2        int
	sX1        int
	sY1        int
	sX2        int
	sY2        int
	horizontal bool
}

func GenerateWorld(width, height, viewerWidth, viewerHeight int) (*worldmap.Map, []*mount.Mount, []*enemy.Enemy, []*npc.Npc) {
	grid := make([][]worldmap.Tile, height)
	for i := 0; i < height; i++ {
		row := make([]worldmap.Tile, width)
		for j := 0; j < width; j++ {
			row[j] = worldmap.NewTile("ground")

		}
		grid[i] = row
	}

	towns := make([]town, 0)
	buildings := make([]building, 0)

	for i := 0; i < 2; i++ {
		generateTown(&grid, &towns, &buildings)
	}
	// Generate building outside town
	generateBuildingOutsideTown(&grid, &towns, &buildings)

	world := worldmap.NewMap(grid, viewerWidth, viewerHeight)
	mounts := generateMounts(world, buildings, 5)
	enemies := generateEnemies(world, 2)
	npcs := generateNpcs(world, buildings, 5)

	return world, mounts, enemies, npcs
}

func addItemsToBuilding(grid *[][]worldmap.Tile, b building) {
	// Consider inner area (exclude walls)
	x1, y1 := b.x1+1, b.y1+1
	x2, y2 := b.x2-1, b.y2-1

	buildingArea := (x2 - x1) * (y2 - y1)

	numOfItems := buildingArea / 2

	for i := 0; i < numOfItems; i++ {
		// Select a random item
		itm := item.GenerateItem()

		// Select a random
		x := x1 + rand.Intn(x2-x1)
		y := y1 + rand.Intn(y2-y1)
		(*grid)[y][x].PlaceItem(itm)

	}

}

// Generate a rectangular building and place on map
func generateBuildingOutsideTown(grid *[][]worldmap.Tile, towns *[]town, buildings *[]building) {
	width := len((*grid)[0])
	height := len(*grid)

	validBuilding := false

	// Keeps trying until a usable building position and size found
	for !validBuilding {
		// Has to be at least three squares wide
		buildingWidth := rand.Intn(3) + 3
		buildingHeight := rand.Intn(3) + 3

		posWidth := 0
		negWidth := 0
		posHeight := 0
		negHeight := 0

		if buildingWidth%2 == 0 {
			posWidth, negWidth = buildingWidth/2, buildingWidth/2
		} else {
			posWidth, negWidth = (buildingWidth+1)/2, buildingWidth/2
		}

		if buildingHeight%2 == 0 {
			posHeight, negHeight = buildingHeight/2, buildingHeight/2
		} else {
			posHeight, negHeight = (buildingHeight+1)/2, buildingHeight/2
		}

		// stop it from reaching the edges of the map
		cX := rand.Intn(width-posWidth) + negWidth
		cY := rand.Intn(height-posHeight) + negHeight

		x1, y1 := cX-negWidth, cY-negHeight
		x2, y2 := cX+posWidth, cY+posHeight

		b := building{x1, y1, x2, y2}
		validBuilding = isValid(x1, y1, width, height) && isValid(x2, y2, width, height) && !overlap(*buildings, b) && !inTowns(*towns, b)

		if validBuilding {
			// Add walls
			for x := x1; x <= x2; x++ {
				(*grid)[y1][x] = worldmap.NewTile("wall")
				(*grid)[y2][x] = worldmap.NewTile("wall")
			}

			for y := y1; y <= y2; y++ {
				(*grid)[y][x1] = worldmap.NewTile("wall")
				(*grid)[y][x2] = worldmap.NewTile("wall")
			}

			// Add door randomly as long it's not a corner
			wallSelection := rand.Intn(4)

			// Wall selection is North, South, East, West
			switch wallSelection {
			case 0:
				doorX := x1 + 1 + rand.Intn(buildingWidth-2)
				(*grid)[y1][doorX] = worldmap.NewTile("door")
			case 1:
				doorX := x1 + 1 + rand.Intn(buildingWidth-2)
				(*grid)[y2][doorX] = worldmap.NewTile("door")
			case 2:
				doorY := y1 + 1 + rand.Intn(buildingHeight-2)
				(*grid)[doorY][x2] = worldmap.NewTile("door")
			case 3:
				doorY := y1 + 1 + rand.Intn(buildingHeight-2)
				(*grid)[doorY][x1] = worldmap.NewTile("door")
			}

			// Add number of windows according total perimeter of building
			perimeter := 2*(y2-y1) + 2*(x2-x1)
			minNumWindows := perimeter / 5
			maxNumWindows := perimeter / 3
			numWindows := minNumWindows + rand.Intn(maxNumWindows-minNumWindows)

			for i := 0; i < numWindows; i++ {

				wallSelection = rand.Intn(4)
				wX, wY := 0, 0

				switch wallSelection {
				case 0:
					doorX := x1 + 1 + rand.Intn(buildingWidth-2)
					wX, wY = doorX, y1
				case 1:
					doorX := x1 + 1 + rand.Intn(buildingWidth-2)
					wX, wY = doorX, y2
				case 2:
					doorY := y1 + 1 + rand.Intn(buildingHeight-2)
					wX, wY = x2, doorY
				case 3:
					doorY := y1 + 1 + rand.Intn(buildingHeight-2)
					wX, wY = x1, doorY
				}

				// If a door is not in place, add window. Otherwise, try again.
				if !(*grid)[wY][wX].IsDoor() {
					(*grid)[wY][wX] = worldmap.NewTile("window")
				} else {
					i--
				}

			}

			// Finally, add items to the building
			addItemsToBuilding(grid, b)

			*buildings = append(*buildings, b)
		}
	}
}

func generateBuildingInTown(grid *[][]worldmap.Tile, t town, buildings *[]building) {

	validBuilding := false
	// Keeps trying until a usable building position and size found
	for !validBuilding {
		// Must be at least 3 in each dimension

		// Width along the street
		buildingWidth := rand.Intn(3) + 3
		depth := rand.Intn(3) + 3

		posWidth := 0
		negWidth := 0

		if buildingWidth%2 == 0 {
			posWidth, negWidth = buildingWidth/2, buildingWidth/2
		} else {
			posWidth, negWidth = (buildingWidth+1)/2, buildingWidth/2
		}

		x1, y1 := 0, 0
		x2, y2 := 0, 0

		sideOfStreet := rand.Intn(2) == 0

		if t.horizontal {
			centreAlongStreet := t.sX1 + rand.Intn(t.sX2-t.sX1)
			x1 = centreAlongStreet - negWidth
			x2 = centreAlongStreet + posWidth

			if sideOfStreet {
				y2 = t.sY1 - 1
				y1 = y2 - depth
			} else {
				y1 = t.sY2 + 1
				y2 = y1 + depth
			}
		} else {
			centreAlongStreet := t.sY1 + rand.Intn(t.sY2-t.sY1)
			y1 = centreAlongStreet - negWidth
			y2 = centreAlongStreet + posWidth

			if sideOfStreet {
				x2 = t.sX1 - 1
				x1 = x2 - depth
			} else {
				x1 = t.sX2 + 1
				x2 = x1 + depth
			}
		}

		b := building{x1, y1, x2, y2}

		validBuilding = inTown(t, b) && !overlap(*buildings, b)

		if validBuilding {

			// Add walls

			for x := x1; x <= x2; x++ {
				(*grid)[y1][x] = worldmap.NewTile("wall")
				(*grid)[y2][x] = worldmap.NewTile("wall")
			}

			for y := y1; y <= y2; y++ {
				(*grid)[y][x1] = worldmap.NewTile("wall")
				(*grid)[y][x2] = worldmap.NewTile("wall")
			}

			// Door is on side facing street
			if t.horizontal {
				doorLocation := x1 + 1 + rand.Intn(buildingWidth-2)
				if sideOfStreet {
					(*grid)[y2][doorLocation] = worldmap.NewTile("door")
				} else {
					(*grid)[y1][doorLocation] = worldmap.NewTile("door")
				}
			} else {
				doorLocation := y1 + 1 + rand.Intn(buildingWidth-2)
				if sideOfStreet {
					(*grid)[doorLocation][x2] = worldmap.NewTile("door")
				} else {
					(*grid)[doorLocation][x1] = worldmap.NewTile("door")
				}
			}
			// Add number of windows according total perimeter of building
			perimeter := 2*buildingWidth + 2*depth
			minNumWindows := perimeter / 5
			maxNumWindows := perimeter / 3
			numWindows := minNumWindows + rand.Intn(maxNumWindows-minNumWindows)

			for i := 0; i < numWindows; i++ {

				wallSelection := rand.Intn(4)
				wX, wY := 0, 0

				switch wallSelection {
				case 0:
					windowX := x1 + 1 + rand.Intn(x2-x1-2)
					wX, wY = windowX, y1
				case 1:
					windowX := x1 + 1 + rand.Intn(x2-x1-2)
					wX, wY = windowX, y2
				case 2:
					windowY := y1 + 1 + rand.Intn(y2-y1-2)
					wX, wY = x2, windowY
				case 3:
					windowY := y1 + 1 + rand.Intn(y2-y1-2)
					wX, wY = x1, windowY
				}

				// If a door is not in place, add window. Otherwise, try again.
				if !(*grid)[wY][wX].IsDoor() {
					(*grid)[wY][wX] = worldmap.NewTile("window")
				} else {
					i--
				}

			}

			// Finally, add items to the building
			addItemsToBuilding(grid, b)

			*buildings = append(*buildings, b)
		}
	}
}

// Generate small town (single street with buildings)
func generateTown(grid *[][]worldmap.Tile, towns *[]town, buildings *[]building) {
	// Generate area of town
	width := len((*grid)[0])
	height := len(*grid)

	validTown := false

	for !validTown {
		townWidth := 10 + rand.Intn(width)
		townHeight := 10 + rand.Intn(height)

		posWidth := 0
		negWidth := 0
		posHeight := 0
		negHeight := 0

		if townWidth%2 == 0 {
			posWidth, negWidth = townWidth/2, townWidth/2
		} else {
			posWidth, negWidth = (townWidth+1)/2, townWidth/2
		}

		if townHeight%2 == 0 {
			posHeight, negHeight = townHeight/2, townHeight/2
		} else {
			posHeight, negHeight = (townHeight+1)/2, townHeight/2
		}

		// stop it from reaching the edges of the map
		cX := rand.Intn(width-posWidth) + negWidth
		cY := rand.Intn(height-posHeight) + negHeight

		x1, y1 := cX-negWidth, cY-negHeight
		x2, y2 := cX+posWidth, cY+posHeight

		streetBreadth := 1 + rand.Intn(5)

		streetX1, streetY1 := 0, 0
		streetX2, streetY2 := 0, 0

		horizontalStreet := rand.Intn(2) == 0
		if horizontalStreet {
			streetX1, streetY1 = x1, cY-streetBreadth/2
			streetX2, streetY2 = x2, cY+(streetBreadth+1)/2
		} else {
			streetX1, streetY1 = cX-streetBreadth/2, y1
			streetX2, streetY2 = cX+(streetBreadth+1)/2, y2
		}

		t := town{x1, y1, x2, y2, streetX1, streetY1, streetX2, streetY2, horizontalStreet}

		validTown := isValid(x1, y1, width, height) && isValid(x2, y2, width, height) && !townsOverlap(*towns, t)
		if validTown {

			//Select random number of buildings, assuming 2 on each side of street
			minNumBuildings, maxNumBuildings := int(math.Max(1, float64(townWidth/10))), int(math.Max(1, float64(townWidth/5)))
			numBuildings := minNumBuildings + rand.Intn(maxNumBuildings-minNumBuildings)
			// Generate a number of buildings
			for i := 0; i < numBuildings; i++ {
				generateBuildingInTown(grid, t, buildings)
			}
			*towns = append(*towns, t)
			break
		}
	}
}

func generateMounts(m *worldmap.Map, buildings []building, n int) []*mount.Mount {
	width := m.GetWidth()
	height := m.GetHeight()
	mounts := make([]*mount.Mount, n)
	for i := 0; i < n; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		if !m.IsPassable(x, y) || m.IsOccupied(x, y) || !outside(buildings, x, y) {
			i--
			continue
		}
		mounts[i] = mount.NewMount("horse", x, y, m)
	}
	return mounts
}

func generateEnemies(m *worldmap.Map, n int) []*enemy.Enemy {
	width := m.GetWidth()
	height := m.GetHeight()
	enemies := make([]*enemy.Enemy, n)
	for i := 0; i < n; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		if !m.IsPassable(x, y) || m.IsOccupied(x, y) {
			i--
			continue
		}
		enemies[i] = enemy.NewEnemy("bandit", x, y, m)
	}
	return enemies
}

func generateNpcs(m *worldmap.Map, buildings []building, n int) []*npc.Npc {
	width := m.GetWidth()
	height := m.GetHeight()
	npcs := make([]*npc.Npc, n)
	for i := 0; i < n; i++ {
		x, y := 0, 0

		// 50/50 chance of being placed inside or outside a building
		insideBuilding := rand.Intn(2) == 0
		if insideBuilding {
			b := buildings[rand.Intn(len(buildings))]
			x = b.x1 + 1 + rand.Intn(b.x2-b.x1-1)
			y = b.y1 + 1 + rand.Intn(b.y2-b.y1-1)
		} else {
			x, y = rand.Intn(width), rand.Intn(height)
		}
		if !m.IsPassable(x, y) || m.IsOccupied(x, y) {
			i--
			continue
		}
		npcs[i] = npc.NewNpc("townsman", x, y, m)
		m.Move(npcs[i], x, y)

	}
	return npcs
}

func isValid(x, y, width, height int) bool {
	return x >= 0 && y >= 0 && x < width && y < height
}

func outside(buildings []building, x, y int) bool {
	for _, b := range buildings {
		if x >= b.x1 && x <= b.x2 && y >= b.x1 && y <= b.x2 {
			return false
		}
	}
	return true
}

func inTown(t town, b building) bool {
	return b.x1 >= t.tX1 && b.x1 <= t.tX2 && b.x2 >= t.tX1 && b.x2 <= t.tX2 && b.y1 >= t.tY1 && b.y1 <= t.tY2 && b.y2 >= t.tY1 && b.y2 <= t.tY2
}

func inTowns(towns []town, b building) bool {
	for _, t := range towns {
		if inTown(t, b) {
			return true
		}
	}
	return false
}

func overlap(buildings []building, b building) bool {
	for _, otherBuilding := range buildings {
		x1y1Overlaps := b.x1 >= otherBuilding.x1-1 && b.x1 <= otherBuilding.x2+1 && b.y1 >= otherBuilding.y1-1 && b.y1 <= otherBuilding.y2+1
		x1y2Overlaps := b.x1 >= otherBuilding.x1-1 && b.x1 <= otherBuilding.x2+1 && b.y2 >= otherBuilding.y1-1 && b.y2 <= otherBuilding.y2+1
		x2y1Overlaps := b.x2 >= otherBuilding.x1-1 && b.x2 <= otherBuilding.x2+1 && b.y1 >= otherBuilding.y1-1 && b.y1 <= otherBuilding.y2+1
		x2y2Overlaps := b.x2 >= otherBuilding.x1-1 && b.x2 <= otherBuilding.x2+1 && b.y2 >= otherBuilding.y1-1 && b.y2 <= otherBuilding.y2+1

		if x1y1Overlaps || x1y2Overlaps || x2y1Overlaps || x2y2Overlaps {
			return true
		}
	}
	return false

}

func townsOverlap(towns []town, t town) bool {
	for _, otherTown := range towns {
		x1y1Overlaps := t.tX1 >= otherTown.tX1 && t.tX1 <= otherTown.tX2 && t.tY1 >= otherTown.tY1 && t.tY1 <= otherTown.tY2
		x1y2Overlaps := t.tX1 >= otherTown.tX1 && t.tX1 <= otherTown.tX2 && t.tY2 >= otherTown.tY1 && t.tY2 <= otherTown.tY2
		x2y1Overlaps := t.tX2 >= otherTown.tX1 && t.tX2 <= otherTown.tX2 && t.tY1 >= otherTown.tY1 && t.tY1 <= otherTown.tY2
		x2y2Overlaps := t.tX2 >= otherTown.tX1 && t.tX2 <= otherTown.tX2 && t.tY2 >= otherTown.tY1 && t.tY2 <= otherTown.tY2

		t1cX, t1cY := (t.tX1+t.tX2)/2, (t.tY1+t.tY2)/2
		t2cX, t2cY := (otherTown.tX1+otherTown.tX2)/2, (otherTown.tY1+otherTown.tY2)/2
		distance := math.Sqrt(math.Pow(float64(t1cX-t2cX), 2) + math.Pow(float64(t1cY-t2cY), 2))

		if x1y1Overlaps || x1y2Overlaps || x2y1Overlaps || x2y2Overlaps || distance < 40 {
			return true
		}
	}
	return false

}
