package world

import (
	"math"
	"math/rand"
	"strings"

	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/npc"
	"github.com/onorton/cowboysindians/worldmap"
)

func GenerateWorld(width, height, viewerWidth, viewerHeight int) (*worldmap.Map, []*npc.Mount, []*npc.Enemy, []*npc.Npc) {
	grid := worldmap.NewGrid(width, height)

	towns := make([]worldmap.Town, 0)
	buildings := make([]worldmap.Building, 0)

	for i := 0; i < 2; i++ {
		generateTown(grid, &towns, &buildings)
	}

	// Generate building outside town
	generateBuildingOutsideTown(grid, &towns, &buildings)

	world := worldmap.NewMap(grid, viewerWidth, viewerHeight)
	placeSignposts(world, towns)
	addItemsToBuildings(world, buildings)

	mounts := generateMounts(world, buildings, 5)
	enemies := generateEnemies(world, 2)
	npcs := generateNpcs(world, towns, buildings, 10)

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
func generateBuildingOutsideTown(grid *worldmap.Grid, towns *[]worldmap.Town, buildings *[]worldmap.Building) {
	width := grid.Width()
	height := grid.Height()

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
				grid.NewTile("wall", x, y1)
				grid.NewTile("wall", x, y2)
			}

			for y := y1; y <= y2; y++ {
				grid.NewTile("wall", x1, y)
				grid.NewTile("wall", x2, y)
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
			grid.NewTile("door", doorX, doorY)
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
				if wX != doorX && wY != doorY {
					grid.NewTile("window", wX, wY)
				} else {
					i--
				}
			}

			*buildings = append(*buildings, b)
		}
	}
}

func randomBuildingType(buildings *[]worldmap.Building) worldmap.BuildingType {
	// 1/2 chance of being residential
	residential := rand.Intn(2) == 0

	if residential {
		return worldmap.Residential
	} else {
		for {
			commercialType := worldmap.BuildingType(rand.Intn(3) + 1)
			// Only one sheriff
			if commercialType == worldmap.Sheriff {

				sheriffExists := false
				for _, b := range *buildings {
					if b.T == worldmap.Sheriff {
						sheriffExists = true
					}
				}
				if !sheriffExists {
					return commercialType
				}
			} else {
				return commercialType
			}
		}
	}
}

func generateBuildingInTown(grid *worldmap.Grid, t *worldmap.Town, buildings *[]worldmap.Building) {

	validBuilding := false

	buildingType := randomBuildingType(buildings)

	// Keeps trying until a usable building position and size found
	for !validBuilding {

		// Must be at least 3 in each dimension

		// Width along the street
		buildingWidth := rand.Intn(5) + 3
		depth := rand.Intn(5) + 3
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

		if t.Horizontal {
			centreAlongStreet := t.SX1 + rand.Intn(t.SX2-t.SX1)
			x1 = centreAlongStreet - negWidth
			x2 = centreAlongStreet + posWidth

			if sideOfStreet {
				y2 = t.SY1 - 1
				y1 = y2 - depth
			} else {
				y1 = t.SY2 + 1
				y2 = y1 + depth
			}
		} else {
			centreAlongStreet := t.SY1 + rand.Intn(t.SY2-t.SY1)
			y1 = centreAlongStreet - negWidth
			y2 = centreAlongStreet + posWidth

			if sideOfStreet {
				x2 = t.SX1 - 1
				x1 = x2 - depth
			} else {
				x1 = t.SX2 + 1
				x2 = x1 + depth
			}
		}

		b := worldmap.Building{x1, y1, x2, y2, buildingType, nil}

		validBuilding = inTown(*t, b) && !overlap(*buildings, b)

		if validBuilding {
			// Add walls
			for x := x1; x <= x2; x++ {
				grid.NewTile("wall", x, y1)
				grid.NewTile("wall", x, y2)
			}

			for y := y1; y <= y2; y++ {
				grid.NewTile("wall", x1, y)
				grid.NewTile("wall", x2, y)
			}

			doorX, doorY := 0, 0
			// Door is on side facing street
			if t.Horizontal {
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
			grid.NewTile("door", doorX, doorY)
			b.DoorLocation = &worldmap.Coordinates{doorX, doorY}

			// If not residential add counter
			if b.T != worldmap.Residential {
				// Choose the side flap will appear
				flapSide := rand.Intn(2) == 0
				if t.Horizontal {
					counterY := 0
					if sideOfStreet {
						counterY = y1 + 2
					} else {
						counterY = y2 - 2
					}

					for x := x1 + 1; x < x2; x++ {
						grid.NewTile("counter", x, counterY)
					}

					if flapSide {
						grid.NewTile("counter flap", x1+1, counterY)
					} else {
						grid.NewTile("counter flap", x2-1, counterY)
					}
				} else {
					counterX := 0
					if sideOfStreet {
						counterX = x1 + 2
					} else {
						counterX = x2 - 2
					}

					for y := y1 + 1; y < y2; y++ {
						grid.NewTile("counter", counterX, y)
					}

					if flapSide {
						grid.NewTile("counter flap", counterX, y1+1)
					} else {
						grid.NewTile("counter flap", counterX, y2-1)
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
				if wX != doorX && wY != doorY {
					grid.NewTile("window", wX, wY)
				} else {
					i--
				}

			}

			*buildings = append(*buildings, b)
			t.Buildings = append(t.Buildings, b)
		}
	}
}

// Generate small town (single street with buildings)
func generateTown(grid *worldmap.Grid, towns *[]worldmap.Town, buildings *[]worldmap.Building) {
	// Generate area of town
	width := grid.Width()
	height := grid.Height()

	validTown := false

	for !validTown {
		townWidth := 15 + rand.Intn(width/5)
		townHeight := 15 + rand.Intn(height/5)

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

		t := &worldmap.Town{generateTownName(), x1, y1, x2, y2, streetX1, streetY1, streetX2, streetY2, horizontalStreet, make([]worldmap.Building, 0)}

		validTown := isValid(x1, y1, width, height) && isValid(x2, y2, width, height) && !townsOverlap(*towns, *t)
		if validTown {

			//Select random number of buildings, assuming 2 on each side of street
			minNumBuildings, maxNumBuildings := int(math.Max(1, float64(townWidth/10))), int(math.Max(1, float64(townWidth/5)))
			numBuildings := minNumBuildings + rand.Intn(maxNumBuildings-minNumBuildings)
			// Generate a number of buildings
			for i := 0; i < numBuildings; i++ {
				generateBuildingInTown(grid, t, buildings)
			}
			*towns = append(*towns, *t)
			break
		}
	}
}

func generateTownName() string {
	noun := npc.Names.Towns["Nouns"][rand.Intn(len(npc.Names.Towns["Nouns"]))]
	withAdjective := rand.Intn(2) == 0
	if withAdjective {
		adjective := npc.Names.Towns["Adjectives"][rand.Intn(len(npc.Names.Towns["Adjectives"]))]
		name := noun

		joined := rand.Intn(2) == 0
		if joined {
			name = adjective + strings.ToLower(noun)
		} else {
			name = adjective + " " + noun
		}

		return name
	}
	return noun
}

func placeSignposts(m *worldmap.Map, towns []worldmap.Town) {
	for _, t := range towns {
		sX, sY := 0, 0

		if t.Horizontal {
			sY = [2]int{t.SY1 - 2, t.SY2 + 2}[rand.Intn(2)]
			sX = [2]int{t.SX1, t.SX2}[rand.Intn(2)]
		} else {
			sX = [2]int{t.SX1 - 2, t.SX2 + 2}[rand.Intn(2)]
			sY = [2]int{t.SY1, t.SY2}[rand.Intn(2)]
		}
		m.PlaceItem(sX, sY, item.NewReadable("signpost", map[string]string{"town": t.Name}))

	}
}

func generateMounts(m *worldmap.Map, buildings []worldmap.Building, n int) []*npc.Mount {
	width := m.GetWidth()
	height := m.GetHeight()
	mounts := make([]*npc.Mount, n)
	for i := 0; i < n; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		if !m.IsPassable(x, y) || m.IsOccupied(x, y) || !outside(buildings, x, y) {
			i--
			continue
		}
		mounts[i] = npc.NewMount("horse", x, y, m)
		m.Move(mounts[i], x, y)
	}
	return mounts
}

func generateEnemies(m *worldmap.Map, n int) []*npc.Enemy {
	width := m.GetWidth()
	height := m.GetHeight()
	enemies := make([]*npc.Enemy, n)
	for i := 0; i < n; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		if !m.IsPassable(x, y) || m.IsOccupied(x, y) {
			i--
			continue
		}
		enemies[i] = npc.NewEnemy("bandit", x, y, m)
		m.Move(enemies[i], x, y)

	}
	return enemies
}

func generateNpcs(m *worldmap.Map, towns []worldmap.Town, buildings []worldmap.Building, n int) []*npc.Npc {
	width := m.GetWidth()
	height := m.GetHeight()
	npcs := make([]*npc.Npc, n)

	usedBuildings := make([]worldmap.Building, 0)

	commercialBuildings := make([]worldmap.Building, 0)
	for _, b := range buildings {
		if b.T != worldmap.Residential {
			commercialBuildings = append(commercialBuildings, b)
		}
	}
	i := 0

	// Place npcs in commerical buildings first
	for ; i < n && i < len(commercialBuildings); i++ {
		b := commercialBuildings[i]
		npcs[i] = placeNpcInBuilding(m, findTown(towns, b), b)
		x, y := npcs[i].GetCoordinates()
		m.Move(npcs[i], x, y)
		usedBuildings = append(usedBuildings, b)
	}

	for ; i < n; i++ {
		x, y := 0, 0

		// 50/50 chance of being placed inside or outside a building
		insideBuilding := rand.Intn(2) == 0
		if insideBuilding {
			b := buildings[rand.Intn(len(buildings))]
			usedBefore := false
			for _, building := range usedBuildings {
				if b == building {
					usedBefore = true
				}
			}

			if usedBefore {
				i--
				continue
			}
			npcs[i] = placeNpcInBuilding(m, findTown(towns, b), b)
			usedBuildings = append(usedBuildings, b)

		} else {
			x, y = rand.Intn(width), rand.Intn(height)
			if !m.IsPassable(x, y) || m.IsOccupied(x, y) {
				i--
				continue
			}
			npcs[i] = npc.NewNpc("townsman", x, y, m)
		}
		x, y = npcs[i].GetCoordinates()
		m.Move(npcs[i], x, y)

	}
	return npcs
}

func placeNpcInBuilding(m *worldmap.Map, t worldmap.Town, b worldmap.Building) *npc.Npc {
	var n *npc.Npc
	for n == nil {
		x := b.X1 + 1 + rand.Intn(b.X2-b.X1-1)
		y := b.Y1 + 1 + rand.Intn(b.Y2-b.Y1-1)

		if !m.IsPassable(x, y) || m.IsOccupied(x, y) {
			continue
		}

		switch b.T {
		case worldmap.Residential:
			n = npc.NewNpc("townsman", x, y, m)
		case worldmap.GunShop:
			n = npc.NewShopkeeper("shopkeeper", x, y, m, t, b)
		case worldmap.Saloon:
			n = npc.NewShopkeeper("bartender", x, y, m, t, b)
		case worldmap.Sheriff:
			n = npc.NewShopkeeper("sheriff", x, y, m, t, b)
		}
	}
	return n
}

func findTown(towns []worldmap.Town, b worldmap.Building) worldmap.Town {
	// Find town building is in
	for _, town := range towns {
		for _, building := range town.Buildings {
			if building == b {
				return town
			}
		}
	}
	return worldmap.Town{}
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

func inTown(t worldmap.Town, b worldmap.Building) bool {
	return b.X1 >= t.TX1 && b.X1 <= t.TX2 && b.X2 >= t.TX1 && b.X2 <= t.TX2 && b.Y1 >= t.TY1 && b.Y1 <= t.TY2 && b.Y2 >= t.TY1 && b.Y2 <= t.TY2
}

func inTowns(towns []worldmap.Town, b worldmap.Building) bool {
	for _, t := range towns {
		if inTown(t, b) {
			return true
		}
	}
	return false
}

func overlap(buildings []worldmap.Building, b worldmap.Building) bool {
	for _, otherBuilding := range buildings {
		if !(b.X2 < otherBuilding.X1-1 || otherBuilding.X2 < b.X1-1 || b.Y2 < otherBuilding.Y1-1 || otherBuilding.Y2 < b.Y1-1) {
			return true
		}
	}
	return false

}

func townsOverlap(towns []worldmap.Town, t worldmap.Town) bool {
	for _, otherTown := range towns {
		if !(t.TX2 < otherTown.TX1 || otherTown.TX2 < t.TX1 || t.TY2 < otherTown.TY1 || otherTown.TY2 < t.TY1) {
			return true
		}

		t1cX, t1cY := (t.TX1+t.TX2)/2, (t.TY1+t.TY2)/2
		t2cX, t2cY := (otherTown.TX1+otherTown.TX2)/2, (otherTown.TY1+otherTown.TY2)/2
		distance := math.Sqrt(math.Pow(float64(t1cX-t2cX), 2) + math.Pow(float64(t1cY-t2cY), 2))

		if distance < 40 {
			return true
		}
	}
	return false

}
