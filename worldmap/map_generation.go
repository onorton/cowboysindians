package worldmap

import (
	"math/rand"
)

func generateMap(width, height int) [][]Tile {
	grid := make([][]Tile, height)
	for i := 0; i < height; i++ {
		row := make([]Tile, width)
		for j := 0; j < width; j++ {
			row[j] = newTile("ground", j, i)

		}
		grid[i] = row
	}

	// Generate a number of buildings
	for i := 0; i < 3; i++ {
		generateBuilding(&grid)
	}

	return grid
}

// Generate a rectangular building and place on map
func generateBuilding(grid *([][]Tile)) {
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

		validBuilding = isValid(x1, y1, width, height) && isValid(x2, y2, width, height)

		if validBuilding {
			// Add walls
			for x := x1; x <= x2; x++ {
				(*grid)[y1][x] = newTile("wall", x, y1)
				(*grid)[y2][x] = newTile("wall", x, y2)
			}

			for y := y1; y <= y2; y++ {
				(*grid)[y][x1] = newTile("wall", x1, y)
				(*grid)[y][x2] = newTile("wall", x2, y)
			}

			// Add door on longest side
			if buildingWidth > buildingHeight {
				(*grid)[cY][x2] = newTile("door", x2, cY)
			} else {
				(*grid)[y1][cX] = newTile("door", x2, cY)
			}
			break
		}
	}

}

func isValid(x, y, width, height int) bool {
	return x >= 0 && y >= 0 && x < width && y < height
}
