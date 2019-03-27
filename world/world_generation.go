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
	buildings := make([]worldmap.Building, 0)

	for i := 0; i < 2; i++ {
		generateTown(&grid, &towns, &buildings)
	}
	// Generate building outside town
	generateBuildingOutsideTown(&grid, &towns, &buildings)

	world := worldmap.NewMap(grid, viewerWidth, viewerHeight)
	addItemsToBuildings(world, buildings)

	mounts := generateMounts(world, buildings, 5)
	enemies := generateEnemies(world, 2)
	npcs := generateNpcs(world, buildings, 10)

	return world, mounts, enemies, npcs
}

func addItemsToBuildings(world *worldmap.Map, buildings []worldmap.Building) {
	for _, b := range buildings {

		// Consider inner area (exclude walls)
		x1, y1 := b.X1+1, b.Y1+1
		x2, y2 := b.X2-1, b.Y2-1

		// If Saloon, place chairs and tables
		if b.T == worldmap.Saloon {

			// Chairs and tables should not be close to counter or front door
			if b.DoorLocation.X == b.X1 {
				x1 += 1
				x2 -= 3
			} else if b.DoorLocation.X == b.X2 {
				x1 += 3
				x2 -= 1
			} else if b.DoorLocation.Y == b.Y1 {
				y1 += 1
				y2 -= 3
			} else if b.DoorLocation.Y == b.Y2 {
				y1 += 3
				y2 -= 1
			}

			for x := x1; x <= x2; x += 4 {
				for y := y1; y <= y2; y += 4 {
					// If a free 3x3 space exists, add a table and four chairs
					if x >= x1 && x+2 <= x2 && y >= y1 && y+2 <= y2 {
						world.PlaceItem(x+1, y+1, item.NewNormalItem("table"))
						world.PlaceItem(x, y+1, item.NewNormalItem("chair"))
						world.PlaceItem(x+1, y, item.NewNormalItem("chair"))
						world.PlaceItem(x+2, y+1, item.NewNormalItem("chair"))
						world.PlaceItem(x+1, y+2, item.NewNormalItem("chair"))
					}
				}
			}

		} else {

			buildingArea := (x2 - x1) * (y2 - y1)

			numOfItems := buildingArea / 2

			for i := 0; i < numOfItems; i++ {

				// Select a random location
				x := x1 + rand.Intn(x2-x1)
				y := y1 + rand.Intn(y2-y1)

				if world.IsPassable(x, y) {
					world.PlaceItem(x, y, item.GenerateItem())
				} else {
					i--
				}
			}
		}
	}

}

// Generate a rectangular building and place on map
func generateBuildingOutsideTown(grid *[][]worldmap.Tile, towns *[]town, buildings *[]worldmap.Building) {
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

		b := worldmap.Building{x1, y1, x2, y2, worldmap.Residential, nil}
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
			doorX, doorY := 0, 0
			switch wallSelection {
			case 0:
				doorX = x1 + 1 + rand.Intn(buildingWidth-2)
				doorY = y1
			case 1:
				doorX = x1 + 1 + rand.Intn(buildingWidth-2)
				doorY = y2
			case 2:
				doorY = y1 + 1 + rand.Intn(buildingHeight-2)
				doorX = x2
			case 3:
				doorY = y1 + 1 + rand.Intn(buildingHeight-2)
				doorX = x1
			}
			(*grid)[doorY][doorX] = worldmap.NewTile("door")
			b.DoorLocation = &worldmap.Coordinates{doorX, doorY}

			// Add number of windows accbording total perimeter of building
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
				if _, ok := (*grid)[wY][wX].(*worldmap.NormalTile); ok {
					(*grid)[wY][wX] = worldmap.NewTile("window")
				} else {
					i--
				}
			}

			*buildings = append(*buildings, b)
		}
	}
}

func generateBuildingInTown(grid *[][]worldmap.Tile, t town, buildings *[]worldmap.Building) {

	validBuilding := false
	// Keeps trying until a usable building position and size found
	for !validBuilding {
		// Must be at least 3 in each dimension

		// Width along the street
		buildingWidth := rand.Intn(5) + 3
		depth := rand.Intn(5) + 3

		// 1/3 chance of being residential
		buildingType := worldmap.BuildingType(rand.Intn(3))

		// Commerical buildings would be larger
		if buildingType != worldmap.Residential {
			buildingWidth = rand.Intn(10) + 8
			depth = rand.Intn(10) + 8
		}

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

		b := worldmap.Building{x1, y1, x2, y2, buildingType, nil}

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

			doorX, doorY := 0, 0
			// Door is on side facing street
			if t.horizontal {
				doorX = x1 + 1 + rand.Intn(buildingWidth-2)
				if sideOfStreet {
					doorY = y2
				} else {
					doorY = y1
				}
			} else {
				doorY = y1 + 1 + rand.Intn(buildingWidth-2)
				if sideOfStreet {
					doorX = x2
				} else {
					doorX = x1
				}
			}
			(*grid)[doorY][doorX] = worldmap.NewTile("door")
			b.DoorLocation = &worldmap.Coordinates{doorX, doorY}

			// If not residential add counter
			if b.T != worldmap.Residential {
				// Choose the side flap will appear
				flapSide := rand.Intn(2) == 0
				if t.horizontal {
					counterY := 0
					if sideOfStreet {
						counterY = y1 + 2
					} else {
						counterY = y2 - 2
					}

					for x := x1 + 1; x < x2; x++ {
						(*grid)[counterY][x] = worldmap.NewTile("counter")
					}

					if flapSide {
						(*grid)[counterY][x1+1] = worldmap.NewTile("counter flap")
					} else {
						(*grid)[counterY][x2-1] = worldmap.NewTile("counter flap")
					}
				} else {
					counterX := 0
					if sideOfStreet {
						counterX = x1 + 2
					} else {
						counterX = x2 - 2
					}

					for y := y1 + 1; y < y2; y++ {
						(*grid)[y][counterX] = worldmap.NewTile("counter")
					}

					if flapSide {
						(*grid)[y1+1][counterX] = worldmap.NewTile("counter flap")
					} else {
						(*grid)[y2-1][counterX] = worldmap.NewTile("counter flap")
					}
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
				if _, ok := (*grid)[wY][wX].(*worldmap.NormalTile); ok {
					(*grid)[wY][wX] = worldmap.NewTile("window")
				} else {
					i--
				}

			}

			*buildings = append(*buildings, b)
		}
	}
}

// Generate small town (single street with buildings)
func generateTown(grid *[][]worldmap.Tile, towns *[]town, buildings *[]worldmap.Building) {
	// Generate area of town
	width := len((*grid)[0])
	height := len(*grid)

	validTown := false

	for !validTown {
		townWidth := 15 + rand.Intn(width)
		townHeight := 15 + rand.Intn(height)

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

func generateMounts(m *worldmap.Map, buildings []worldmap.Building, n int) []*mount.Mount {
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

func generateNpcs(m *worldmap.Map, buildings []worldmap.Building, n int) []*npc.Npc {
	width := m.GetWidth()
	height := m.GetHeight()
	npcs := make([]*npc.Npc, n)

	usedBuildings := make([]worldmap.Building, 0)

	for i := 0; i < n; i++ {
		x, y := 0, 0

		// 50/50 chance of being placed inside or outside a building
		insideBuilding := rand.Intn(2) == 0
		if insideBuilding {
			b := buildings[rand.Intn(len(buildings))]
			x = b.X1 + 1 + rand.Intn(b.X2-b.X1-1)
			y = b.Y1 + 1 + rand.Intn(b.Y2-b.Y1-1)
			usedBefore := false
			for _, building := range usedBuildings {
				if b == building {
					usedBefore = true
				}
			}

			if usedBefore || !m.IsPassable(x, y) || m.IsOccupied(x, y) {
				i--
				continue
			}
			usedBuildings = append(usedBuildings, b)
			switch b.T {
			case worldmap.Residential:
				npcs[i] = npc.NewNpc("townsman", x, y, m)
			case worldmap.GunShop:
				npcs[i] = npc.NewShopkeeper("shopkeeper", x, y, m, b)
			case worldmap.Saloon:
				npcs[i] = npc.NewShopkeeper("bartender", x, y, m, b)

			}
		} else {
			x, y = rand.Intn(width), rand.Intn(height)
			if !m.IsPassable(x, y) || m.IsOccupied(x, y) {
				i--
				continue
			}
			npcs[i] = npc.NewNpc("townsman", x, y, m)
		}

		m.Move(npcs[i], x, y)

	}
	return npcs
}

func isValid(x, y, width, height int) bool {
	return x >= 0 && y >= 0 && x < width && y < height
}

func outside(buildings []worldmap.Building, x, y int) bool {
	for _, b := range buildings {
		if b.Inside(x, y) {
			return false
		}
	}
	return true
}

func inTown(t town, b worldmap.Building) bool {
	return b.X1 >= t.tX1 && b.X1 <= t.tX2 && b.X2 >= t.tX1 && b.X2 <= t.tX2 && b.Y1 >= t.tY1 && b.Y1 <= t.tY2 && b.Y2 >= t.tY1 && b.Y2 <= t.tY2
}

func inTowns(towns []town, b worldmap.Building) bool {
	for _, t := range towns {
		if inTown(t, b) {
			return true
		}
	}
	return false
}

func overlap(buildings []worldmap.Building, b worldmap.Building) bool {
	for _, otherBuilding := range buildings {
		x1y1Overlaps := b.X1 >= otherBuilding.X1-1 && b.X1 <= otherBuilding.X2+1 && b.Y1 >= otherBuilding.Y1-1 && b.Y1 <= otherBuilding.Y2+1
		x1y2Overlaps := b.X1 >= otherBuilding.X1-1 && b.X1 <= otherBuilding.X2+1 && b.Y2 >= otherBuilding.Y1-1 && b.Y2 <= otherBuilding.Y2+1
		x2y1Overlaps := b.X2 >= otherBuilding.X1-1 && b.X2 <= otherBuilding.X2+1 && b.Y1 >= otherBuilding.Y1-1 && b.Y1 <= otherBuilding.Y2+1
		x2y2Overlaps := b.X2 >= otherBuilding.X1-1 && b.X2 <= otherBuilding.X2+1 && b.Y2 >= otherBuilding.Y1-1 && b.Y2 <= otherBuilding.Y2+1

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
