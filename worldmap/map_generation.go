package worldmap

import (
	"math"
	"math/rand"
)

type building struct {
	x1 int
	y1 int
	x2 int
	y2 int
}

type town struct {
	tX1 int
	tY1 int
	tX2 int
	tY2 int
	sX1 int
	sY1 int
	sX2 int
	sY2 int
}

func generateMap(width, height int) [][]Tile {
	grid := make([][]Tile, height)
	for i := 0; i < height; i++ {
		row := make([]Tile, width)
		for j := 0; j < width; j++ {
			row[j] = newTile("ground")

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

	return grid
}

// Generate a rectangular building and place on map
func generateBuildingOutsideTown(grid *[][]Tile, towns *[]town, buildings *[]building) {
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
		validBuilding = isValid(x1, y1, width, height) && isValid(x2, y2, width, height) && !overlap(*buildings, b) && inTowns(*towns, b)

		if validBuilding {
			// Add walls
			for x := x1; x <= x2; x++ {
				(*grid)[y1][x] = newTile("wall")
				(*grid)[y2][x] = newTile("wall")
			}

			for y := y1; y <= y2; y++ {
				(*grid)[y][x1] = newTile("wall")
				(*grid)[y][x2] = newTile("wall")
			}

			// Add door randomly as long it's not a corner
			wallSelection := rand.Intn(4)

			// Wall selection is North, South, East, West
			switch wallSelection {
			case 0:
				doorX := x1 + 1 + rand.Intn(buildingWidth-2)
				(*grid)[y1][doorX] = newTile("door")
			case 1:
				doorX := x1 + 1 + rand.Intn(buildingWidth-2)
				(*grid)[y2][doorX] = newTile("door")
			case 2:
				doorY := y1 + 1 + rand.Intn(buildingHeight-2)
				(*grid)[doorY][x2] = newTile("door")
			case 3:
				doorY := y1 + 1 + rand.Intn(buildingHeight-2)
				(*grid)[doorY][x1] = newTile("door")
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
				if !(*grid)[wY][wX].door {
					(*grid)[wY][wX] = newTile("window")
				} else {
					i--
				}

			}

			*buildings = append(*buildings, b)
		}
	}
}

func generateBuildingInTown(grid *[][]Tile, t town, buildings *[]building) {

	validBuilding := false
	// Keeps trying until a usable building position and size found
	for !validBuilding {
		// Has to be at least three squares wide
		buildingWidth := rand.Intn(3) + 3
		buildingHeight := rand.Intn(3) + 3

		posWidth := 0
		negWidth := 0

		if buildingWidth%2 == 0 {
			posWidth, negWidth = buildingWidth/2, buildingWidth/2
		} else {
			posWidth, negWidth = (buildingWidth+1)/2, buildingWidth/2
		}

		// stop it from reaching the edges of the map
		cX := t.sX1 + rand.Intn(t.sX2-t.sX1)
		northOfStreet := rand.Intn(2) == 0
		x1 := cX - negWidth
		x2 := cX + posWidth

		y1, y2 := 0, 0

		if northOfStreet {
			y2 = t.sY1 - 1
			y1 = y2 - buildingHeight
		} else {
			y1 = t.sY2 + 1
			y2 = y1 + buildingHeight
		}

		b := building{x1, y1, x2, y2}
		validBuilding = inTown(t, b) && !overlap(*buildings, b)

		if validBuilding {

			// Add walls
			for x := x1; x <= x2; x++ {
				(*grid)[y1][x] = newTile("wall")
				(*grid)[y2][x] = newTile("wall")
			}

			for y := y1; y <= y2; y++ {
				(*grid)[y][x1] = newTile("wall")
				(*grid)[y][x2] = newTile("wall")
			}

			doorX := x1 + 1 + rand.Intn(buildingWidth-2)
			// Door is on side facing street
			if northOfStreet {
				(*grid)[y2][doorX] = newTile("door")
			} else {
				(*grid)[y1][doorX] = newTile("door")
			}
			// Add number of windows according total perimeter of building
			perimeter := 2*(y2-y1) + 2*(x2-x1)
			minNumWindows := perimeter / 5
			maxNumWindows := perimeter / 3
			numWindows := minNumWindows + rand.Intn(maxNumWindows-minNumWindows)

			for i := 0; i < numWindows; i++ {

				wallSelection := rand.Intn(4)
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
				if !(*grid)[wY][wX].door {
					(*grid)[wY][wX] = newTile("window")
				} else {
					i--
				}

			}
			*buildings = append(*buildings, b)
		}
	}
}

// Generate small town (single street with buildings)
func generateTown(grid *[][]Tile, towns *[]town, buildings *[]building) {
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

		// For now single horizontal street
		streetBreadth := 1 + rand.Intn(5)

		streetX1, streetY1 := x1, cY-streetBreadth/2
		streetX2, streetY2 := x2, cY+(streetBreadth+1)/2

		t := town{x1, y1, x2, y2, streetX1, streetY1, streetX2, streetY2}

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

func isValid(x, y, width, height int) bool {
	return x >= 0 && y >= 0 && x < width && y < height
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
		x1y1Overlaps := t.tX1 >= otherTown.tX1-1 && t.tX1 <= otherTown.tX2+1 && t.tY1 >= otherTown.tY1-1 && t.tY1 <= otherTown.tY2+1
		x1y2Overlaps := t.tX1 >= otherTown.tX1-1 && t.tX1 <= otherTown.tX2+1 && t.tY2 >= otherTown.tY1-1 && t.tY2 <= otherTown.tY2+1
		x2y1Overlaps := t.tX2 >= otherTown.tX1-1 && t.tX2 <= otherTown.tX2+1 && t.tY1 >= otherTown.tY1-1 && t.tY1 <= otherTown.tY2+1
		x2y2Overlaps := t.tX2 >= otherTown.tX1-1 && t.tX2 <= otherTown.tX2+1 && t.tY2 >= otherTown.tY1-1 && t.tY2 <= otherTown.tY2+1

		if x1y1Overlaps || x1y2Overlaps || x2y1Overlaps || x2y2Overlaps {
			return true
		}
	}
	return false

}
