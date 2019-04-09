package npc

import "github.com/onorton/cowboysindians/worldmap"

func generateMap(aiMap [][]int, world *worldmap.Map, location worldmap.Coordinates) [][]int {
	width, height := len(aiMap[0]), len(aiMap)
	prev := make([][]int, height)
	for i := range prev {
		prev[i] = make([]int, width)
	}
	// While map changes, update
	for !compareMaps(aiMap, prev) {
		prev = copyMap(aiMap)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				wX, wY := location.X+x-width/2, location.Y+y-height/2
				if !world.IsValid(wX, wY) || !world.IsPassable(wX, wY) {
					continue
				}
				min := height * width
				for i := -1; i <= 1; i++ {
					for j := -1; j <= 1; j++ {
						nX := x + i
						nY := y + j

						if nX >= 0 && nX < width && nY >= 0 && nY < height && aiMap[nY][nX] < min {
							min = aiMap[nY][nX]
						}
					}
				}

				if aiMap[y][x] > min {
					aiMap[y][x] = min + 1
				}

			}
		}
	}
	return aiMap
}

func (m *Mount) getWaypointMap(waypoint worldmap.Coordinates) [][]int {
	d := m.GetVisionDistance()
	// Creature will be at location d,d in this AI map
	width := 2*d + 1
	aiMap := make([][]int, width)

	// Initialise Dijkstra map with goals.
	// Max is size of grid.
	for i := -d; i < d+1; i++ {
		aiMap[i+d] = make([]int, width)
		for j := -d; j < d+1; j++ {
			x := j + d
			y := i + d
			location := worldmap.Coordinates{m.location.X + j, m.location.Y + i}
			if waypoint == location {
				aiMap[y][x] = 0
			} else {
				aiMap[y][x] = width * width
			}
		}
	}
	return generateMap(aiMap, m.world, m.location)
}

func compareMaps(m, o [][]int) bool {
	for i := 0; i < len(m); i++ {
		for j := 0; j < len(m[0]); j++ {
			if m[i][j] != o[i][j] {
				return false
			}
		}
	}
	return true

}

func addMaps(maps [][][]int, weights []float64) [][]float64 {
	result := make([][]float64, len(maps[0]))

	for y, row := range maps[0] {
		result[y] = make([]float64, len(row))
	}

	for i, _ := range maps {
		for y, row := range maps[i] {
			for x, location := range row {
				result[y][x] += weights[i] * float64(location)
			}
		}
	}
	return result
}

func copyMap(o [][]int) [][]int {
	h := len(o)
	w := len(o[0])
	c := make([][]int, h)
	for i, _ := range o {
		c[i] = make([]int, w)
		copy(c[i], o[i])
	}
	return c
}

func (npc *Npc) getWaypointMap(waypoint worldmap.Coordinates) [][]int {
	d := npc.GetVisionDistance()
	// Creature will be at location d,d in this AI map
	width := 2*d + 1
	aiMap := make([][]int, width)

	// Initialise Dijkstra map with goals.
	// Max is size of grid.
	for i := -d; i < d+1; i++ {
		aiMap[i+d] = make([]int, width)
		for j := -d; j < d+1; j++ {
			x := j + d
			y := i + d
			location := worldmap.Coordinates{npc.location.X + j, npc.location.Y + i}
			if waypoint == location {
				aiMap[y][x] = 0
			} else {
				aiMap[y][x] = width * width
			}
		}
	}
	return generateMap(aiMap, npc.world, npc.location)
}

func (npc *Npc) getMountMap() [][]int {
	d := npc.GetVisionDistance()
	// Creature will be at location d,d in this AI map
	width := 2*d + 1
	aiMap := make([][]int, width)

	// Initialise Dijkstra map with goals.
	// Max is size of grid.
	for i := -d; i < d+1; i++ {
		aiMap[i+d] = make([]int, width)
		for j := -d; j < d+1; j++ {
			x := j + d
			y := i + d
			// Translate location into world coordinates
			wX, wY := npc.location.X+j, npc.location.Y+i
			// Looks for mount on its own
			if npc.world.IsValid(wX, wY) && npc.world.IsVisible(npc, wX, wY) {
				if m, ok := npc.world.GetCreature(wX, wY).(*Mount); ok && m != nil {
					aiMap[y][x] = 0
				} else {
					aiMap[y][x] = width * width
				}
			}
		}
	}
	return generateMap(aiMap, npc.world, npc.location)
}

func (e *Enemy) getChaseMap() [][]int {
	d := e.GetVisionDistance()
	// Creature will be at location d,d in this AI map
	width := 2*d + 1
	aiMap := make([][]int, width)

	// Initialise Dijkstra map with goals.
	// Max is size of grid.
	for i := -d; i < d+1; i++ {
		aiMap[i+d] = make([]int, width)
		for j := -d; j < d+1; j++ {
			x := j + d
			y := i + d
			// Translate location into world coordinates
			wX, wY := e.location.X+j, e.location.Y+i
			if e.world.IsValid(wX, wY) && e.world.IsVisible(e, wX, wY) && e.world.HasPlayer(wX, wY) {
				aiMap[y][x] = 0
			} else {
				aiMap[y][x] = width * width
			}
		}
	}
	return generateMap(aiMap, e.world, e.location)
}

func (e *Enemy) getItemMap() [][]int {
	d := e.GetVisionDistance()
	// Creature will be at location d,d in this AI map
	width := 2*d + 1
	aiMap := make([][]int, width)

	// Initialise Dijkstra map with goals.
	// Max is size of grid.
	for i := -d; i < d+1; i++ {
		aiMap[i+d] = make([]int, width)
		for j := -d; j < d+1; j++ {
			x := j + d
			y := i + d
			// Translate location into world coordinates
			wX, wY := e.location.X+j, e.location.Y+i
			if e.world.IsValid(wX, wY) && e.world.IsVisible(e, wX, wY) && e.world.HasItems(wX, wY) {
				aiMap[y][x] = 0
			} else {
				aiMap[y][x] = width * width
			}
		}
	}
	return generateMap(aiMap, e.world, e.location)
}

func (e *Enemy) getCoverMap() [][]int {
	d := e.GetVisionDistance()
	// Creature will be at location d,d in this AI map
	width := 2*d + 1
	aiMap := make([][]int, width)

	player := e.world.GetPlayer()
	pX, pY := player.GetCoordinates()

	// Initialise Dijkstra map with goals.
	// Max is size of grid.
	for i := -d; i < d+1; i++ {
		aiMap[i+d] = make([]int, width)
		for j := -d; j < d+1; j++ {
			x := j + d
			y := i + d
			// Translate location into world coordinates
			wX, wY := e.location.X+j, e.location.Y+i
			// Enemy must be able to see player in order to know it would be behind cover
			if e.world.IsValid(wX, wY) && e.world.IsVisible(e, wX, wY) && e.world.IsVisible(e, pX, pY) && e.world.BehindCover(wX, wY, player) {
				aiMap[y][x] = 0
			} else {
				aiMap[y][x] = width * width
			}
		}
	}
	return generateMap(aiMap, e.world, e.location)
}

func (e *Enemy) getMountMap() [][]int {
	d := e.GetVisionDistance()
	// Creature will be at location d,d in this AI map
	width := 2*d + 1
	aiMap := make([][]int, width)

	// Initialise Dijkstra map with goals.
	// Max is size of grid.
	for i := -d; i < d+1; i++ {
		aiMap[i+d] = make([]int, width)
		for j := -d; j < d+1; j++ {
			x := j + d
			y := i + d
			// Translate location into world coordinates
			wX, wY := e.location.X+j, e.location.Y+i
			// Looks for mount on its own
			if e.world.IsValid(wX, wY) && e.world.IsVisible(e, wX, wY) {
				if m, ok := e.world.GetCreature(wX, wY).(*Mount); ok && m != nil {
					aiMap[y][x] = 0
				} else {
					aiMap[y][x] = width * width
				}
			}
		}
	}
	return generateMap(aiMap, e.world, e.location)
}
