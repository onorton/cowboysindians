package worldmap

import (
	"math/rand"
)

type building struct {
	x1 int
	y1 int
	x2 int
	y2 int
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

	buildings := make([]building, 0)
	// Generate a number of buildings
	for i := 0; i < 3; i++ {
		generateBuilding(&grid, &buildings)
	}

	return grid
}

// Generate a rectangular building and place on map
func generateBuilding(grid *[][]Tile, buildings *[]building) {
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
		validBuilding = isValid(x1, y1, width, height) && isValid(x2, y2, width, height) && !overlap(*buildings, b)

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
			*buildings = append(*buildings, b)
		}
	}
}

func isValid(x, y, width, height int) bool {
	return x >= 0 && y >= 0 && x < width && y < height
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
