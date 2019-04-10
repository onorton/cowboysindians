package npc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/onorton/cowboysindians/worldmap"
)

type mountAi struct {
	waypoint worldmap.WaypointSystem
}

func (ai mountAi) update(c worldmap.Creature, world *worldmap.Map) (int, int) {
	x, y := c.GetCoordinates()
	location := worldmap.Coordinates{x, y}
	waypoint := ai.waypoint.NextWaypoint(location)
	aiMap := getWaypointMap(waypoint, world, location, c.GetVisionDistance())
	current := aiMap[c.GetVisionDistance()][c.GetVisionDistance()]
	possibleLocations := make([]worldmap.Coordinates, 0)
	// Find adjacent locations closer to the goal
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			nX := location.X + i
			nY := location.Y + j
			if aiMap[nY-location.Y+c.GetVisionDistance()][nX-location.X+c.GetVisionDistance()] <= current {
				// Add if not occupied
				if world.IsValid(nX, nY) && !world.IsOccupied(nX, nY) {
					possibleLocations = append(possibleLocations, worldmap.Coordinates{nX, nY})
				}
			}
		}
	}
	if len(possibleLocations) > 0 {
		l := possibleLocations[rand.Intn(len(possibleLocations))]
		return l.X, l.Y
	}

	return c.GetCoordinates()
}

func (ai mountAi) SetMap(world *worldmap.Map) {
	if w, ok := ai.waypoint.(*worldmap.RandomWaypoint); ok {
		w.SetMap(world)
	}
}

func (ai mountAi) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	waypointValue, err := json.Marshal(ai.waypoint)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Waypoint\":%s", waypointValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (ai *mountAi) UnmarshalJSON(data []byte) error {
	type mountAiJson struct {
		Waypoint map[string]interface{}
	}

	var v mountAiJson
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	ai.waypoint = worldmap.UnmarshalWaypointSystem(v.Waypoint)
	return nil
}

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

func getWaypointMap(waypoint worldmap.Coordinates, world *worldmap.Map, location worldmap.Coordinates, d int) [][]int {

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
			loc := worldmap.Coordinates{location.X + j, location.Y + i}
			if waypoint == loc {
				aiMap[y][x] = 0
			} else {
				aiMap[y][x] = width * width
			}
		}
	}
	return generateMap(aiMap, world, location)
}

func getMountMap(c worldmap.Creature, world *worldmap.Map) [][]int {
	d := c.GetVisionDistance()
	cX, cY := c.GetCoordinates()
	location := worldmap.Coordinates{cX, cY}
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
			wX, wY := location.X+j, location.Y+i
			// Looks for mount on its own
			if world.IsValid(wX, wY) && world.IsVisible(c, wX, wY) {
				if m, ok := world.GetCreature(wX, wY).(*Mount); ok && m != nil {
					aiMap[y][x] = 0
				} else {
					aiMap[y][x] = width * width
				}
			}
		}
	}
	return generateMap(aiMap, world, location)
}

func getChaseMap(c worldmap.Creature, world *worldmap.Map) [][]int {
	d := c.GetVisionDistance()
	cX, cY := c.GetCoordinates()
	location := worldmap.Coordinates{cX, cY}
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
			wX, wY := location.X+j, location.Y+i
			if world.IsValid(wX, wY) && world.IsVisible(c, wX, wY) && world.HasPlayer(wX, wY) {
				aiMap[y][x] = 0
			} else {
				aiMap[y][x] = width * width
			}
		}
	}
	return generateMap(aiMap, world, location)
}

func getItemMap(c worldmap.Creature, world *worldmap.Map) [][]int {
	d := c.GetVisionDistance()
	cX, cY := c.GetCoordinates()
	location := worldmap.Coordinates{cX, cY}
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
			wX, wY := location.X+j, location.Y+i
			if world.IsValid(wX, wY) && world.IsVisible(c, wX, wY) && world.HasItems(wX, wY) {
				aiMap[y][x] = 0
			} else {
				aiMap[y][x] = width * width
			}
		}
	}
	return generateMap(aiMap, world, location)
}

func getCoverMap(c worldmap.Creature, world *worldmap.Map) [][]int {
	d := c.GetVisionDistance()
	cX, cY := c.GetCoordinates()
	location := worldmap.Coordinates{cX, cY}
	// Creature will be at location d,d in this AI map
	width := 2*d + 1
	aiMap := make([][]int, width)

	player := world.GetPlayer()
	pX, pY := player.GetCoordinates()

	// Initialise Dijkstra map with goals.
	// Max is size of grid.
	for i := -d; i < d+1; i++ {
		aiMap[i+d] = make([]int, width)
		for j := -d; j < d+1; j++ {
			x := j + d
			y := i + d
			// Translate location into world coordinates
			wX, wY := location.X+j, location.Y+i
			// Enemy must be able to see player in order to know it would be behind cover
			if world.IsValid(wX, wY) && world.IsVisible(c, wX, wY) && world.IsVisible(c, pX, pY) && world.BehindCover(wX, wY, player) {
				aiMap[y][x] = 0
			} else {
				aiMap[y][x] = width * width
			}
		}
	}
	return generateMap(aiMap, world, location)
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
