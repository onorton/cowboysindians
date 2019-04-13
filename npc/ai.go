package npc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"

	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/worldmap"
)

type hasAi interface {
	heal(int)
	damageable
	worldmap.CanSee
	worldmap.CanCrouch
}

type holdsItems interface {
	dropItem(item.Item)
	pickupItem(item.Item)
	inventory() []item.Item
	overEncumbered() bool
	removeItem(item.Item)
}

type usesItems interface {
	wieldItem() bool
	wearArmour() bool
	ranged() bool
	rangedAttack(worldmap.Creature, int)
	Weapon() *item.Weapon
	weaponLoaded() bool
	weaponFullyLoaded() bool
	hasAmmo() bool
	getAmmo() *item.Ammo
}

type damageable interface {
	bloodied() bool
	AttackHits(int) bool
}

type ai interface {
	update(hasAi, *worldmap.Map) (int, int)
}

type mountAi struct {
	waypoint worldmap.WaypointSystem
}

func (ai mountAi) update(c hasAi, world *worldmap.Map) (int, int) {
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

func (ai mountAi) setMap(world *worldmap.Map) {
	switch w := ai.waypoint.(type) {
	case *worldmap.RandomWaypoint:
		w.SetMap(world)
	case *worldmap.Patrol:
	case *worldmap.WithinBuilding:
		w.SetMap(world)
	}
}

func (ai mountAi) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	buffer.WriteString("\"Type\":\"mount\",")

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

type npcAi struct {
	waypoint worldmap.WaypointSystem
}

func (ai npcAi) update(c hasAi, world *worldmap.Map) (int, int) {
	cX, cY := c.GetCoordinates()
	location := worldmap.Coordinates{cX, cY}
	waypoint := ai.waypoint.NextWaypoint(location)
	aiMap := getWaypointMap(waypoint, world, location, c.GetVisionDistance())
	mountMap := getMountMap(c, world)

	current := aiMap[c.GetVisionDistance()][c.GetVisionDistance()]
	possibleLocations := make([]worldmap.Coordinates, 0)

	// Find adjacent locations closer to the goal
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			nX := location.X + i
			nY := location.Y + j
			if aiMap[nY-location.Y+c.GetVisionDistance()][nX-location.X+c.GetVisionDistance()] < current {
				// Add if not occupied
				if world.IsValid(nX, nY) && !world.IsOccupied(nX, nY) {
					possibleLocations = append(possibleLocations, worldmap.Coordinates{nX, nY})
				}
			}
		}
	}

	// If can ride things and mounted, can move first before executing another action
	if r, ok := c.(Rider); ok && r.Mount() != nil && r.Mount().Moved() {
		if len(possibleLocations) > 0 {
			if itemHolder, ok := c.(holdsItems); ok && itemHolder.overEncumbered() {
				for _, itm := range itemHolder.inventory() {
					if itm.GetWeight() > 1 {
						itemHolder.dropItem(itm)
						return cX, cY
					}
				}
			} else {
				l := possibleLocations[rand.Intn(len(possibleLocations))]
				r.Mount().Move()
				c.SetCoordinates(l.X, l.Y)
				// Can choose new action again
				return ai.update(c, world)
			}
		}
	}

	// If at half health heal up
	if itemHolder, ok := c.(holdsItems); ok && c.bloodied() {
		for _, itm := range itemHolder.inventory() {
			if con, ok := itm.(*item.Consumable); ok && con.GetEffect("hp") > 0 {
				c.heal(con.GetEffect("hp"))
				itemHolder.removeItem(con)
				return cX, cY
			}
		}
	}

	// If adjacent to mount, attempt to mount it
	if r, ok := c.(Rider); ok && r.Mount() == nil {
		for i := -1; i <= 1; i++ {
			for j := -1; j <= 1; j++ {
				x, y := location.X+j, location.Y+i
				if world.IsValid(x, y) && mountMap[c.GetVisionDistance()+i][c.GetVisionDistance()+j] == 0 {
					m, _ := world.GetCreature(x, y).(*Mount)
					m.AddRider(r)
					world.DeleteCreature(m)
					r.AddMount(m)
					c.Standup()
					return x, y
				}
			}
		}
	}

	// If adjacent to closed door attempt to open it
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			x, y := location.X+j, location.Y+i
			if world.IsValid(x, y) && world.IsDoor(x, y) && !world.IsPassable(x, y) {
				world.ToggleDoor(x, y, true)
				return x, y
			}
		}
	}

	if len(possibleLocations) > 0 {
		if itemHolder, ok := c.(holdsItems); ok && itemHolder.overEncumbered() {
			for _, itm := range itemHolder.inventory() {
				if itm.GetWeight() > 1 {
					itemHolder.dropItem(itm)
					return cX, cY
				}
			}
		} else if r, ok := c.(Rider); ok && (r.Mount() == nil || !r.Mount().Moved()) {
			l := possibleLocations[rand.Intn(len(possibleLocations))]
			return l.X, l.Y
		}
	} else if itemHolder, ok := c.(holdsItems); ok {
		if items := world.GetItems(location.X, location.Y); len(items) > 0 {
			for _, item := range items {
				itemHolder.pickupItem(item)
			}
			return cX, cY
		}
	}

	// If the npc can do nothing else, try moving randomly
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			x, y := cX+j, cY+i
			if world.IsValid(x, y) && world.IsPassable(x, y) && !world.IsOccupied(x, y) {
				possibleLocations = append(possibleLocations, worldmap.Coordinates{x, y})
			}
		}
	}

	if len(possibleLocations) > 0 {
		l := possibleLocations[rand.Intn(len(possibleLocations))]
		return l.X, l.Y
	}

	return cX, cY
}

func (ai npcAi) setMap(world *worldmap.Map) {
	switch w := ai.waypoint.(type) {
	case *worldmap.RandomWaypoint:
		w.SetMap(world)
	case *worldmap.Patrol:
	case *worldmap.WithinBuilding:
		w.SetMap(world)
	}
}

func (ai npcAi) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	buffer.WriteString("\"Type\":\"npc\",")

	waypointValue, err := json.Marshal(ai.waypoint)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Waypoint\":%s", waypointValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (ai *npcAi) UnmarshalJSON(data []byte) error {
	type npcAiJson struct {
		Waypoint map[string]interface{}
	}

	var v npcAiJson
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	ai.waypoint = worldmap.UnmarshalWaypointSystem(v.Waypoint)
	return nil
}

type enemyAi struct {
}

func (ai enemyAi) update(c hasAi, world *worldmap.Map) (int, int) {
	cX, cY := c.GetCoordinates()
	location := worldmap.Coordinates{cX, cY}

	coefficients := []float64{0.5, 0.2, 0.3, 0.0}

	// Focus on getting a mount if possible
	if r, ok := c.(Rider); ok && r.Mount() == nil {
		coefficients = []float64{0.3, 0.2, 0.1, 0.4}
	}
	coverMap := getCoverMap(c, world)
	mountMap := getMountMap(c, world)
	aiMap := addMaps([][][]int{getChaseMap(c, world), getItemMap(c, world), coverMap, mountMap}, coefficients)

	current := aiMap[c.GetVisionDistance()][c.GetVisionDistance()]
	possibleLocations := make([]worldmap.Coordinates, 0)
	// Find adjacent locations closer to the goal
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			nX := location.X + i
			nY := location.Y + j
			if aiMap[nY-location.Y+c.GetVisionDistance()][nX-location.X+c.GetVisionDistance()] < current {
				// Add if not occupied by another enemy
				if world.IsValid(nX, nY) && (world.HasPlayer(nX, nY) || !world.IsOccupied(nX, nY)) {
					possibleLocations = append(possibleLocations, worldmap.Coordinates{nX, nY})
				}
			}
		}
	}
	// If mounted, can move first before executing another action
	if r, ok := c.(Rider); ok && r.Mount() != nil && !r.Mount().Moved() {
		if len(possibleLocations) > 0 {
			if itemHolder, ok := c.(holdsItems); ok && itemHolder.overEncumbered() {
				for _, itm := range itemHolder.inventory() {
					if itm.GetWeight() > 1 {
						itemHolder.dropItem(itm)
						return cX, cY
					}
				}
			} else {
				l := possibleLocations[rand.Intn(len(possibleLocations))]
				r.Mount().Move()
				c.SetCoordinates(l.X, l.Y)
				// Can choose new action again
				return ai.update(c, world)
			}
		}
	}

	// If at half health heal up
	if itemHolder, ok := c.(holdsItems); ok && c.bloodied() {
		for _, itm := range itemHolder.inventory() {
			if con, ok := itm.(*item.Consumable); ok && con.GetEffect("hp") > 0 {
				c.heal(con.GetEffect("hp"))
				itemHolder.removeItem(con)
				return cX, cY
			}
		}
	}

	// If moving into or out of cover and not mounted toggle crouch
	if r, ok := c.(Rider); ok && r.Mount() == nil {
		if coverMap[c.GetVisionDistance()][c.GetVisionDistance()] == 0 && !c.IsCrouching() {
			c.Crouch()
			return cX, cY
		} else if coverMap[c.GetVisionDistance()][c.GetVisionDistance()] > 0 && c.IsCrouching() {
			c.Standup()
			return cX, cY
		}
	}

	// Try and wield best weapon
	if itemUser, ok := c.(usesItems); ok && itemUser.wieldItem() {
		return cX, cY
	}
	// Try and wear best armour
	if itemUser, ok := c.(usesItems); ok && itemUser.wearArmour() {
		return cX, cY
	}

	target := world.GetPlayer()
	tX, tY := target.GetCoordinates()

	if itemUser, ok := c.(usesItems); ok {
		if distance := math.Sqrt(math.Pow(float64(location.X-tX), 2) + math.Pow(float64(location.Y-tY), 2)); itemUser.ranged() && distance < float64(itemUser.Weapon().GetRange()) && world.IsVisible(c, tX, tY) {
			// if weapon loaded, shoot at target else if enemy has ammo, load weapon
			if itemUser.weaponLoaded() {
				itemUser.Weapon().Fire()
				coverPenalty := 0
				if world.TargetBehindCover(c, target) {
					coverPenalty = 5
				}
				itemUser.rangedAttack(target, -coverPenalty)
				return cX, cY
			} else if itemUser.hasAmmo() {
				for !itemUser.weaponFullyLoaded() && itemUser.hasAmmo() {
					itemUser.getAmmo()
					itemUser.Weapon().Load()
				}
				return cX, cY
			}
		}

	}

	// If adjacent to mount, attempt to mount it
	if r, ok := c.(Rider); ok && r.Mount() == nil {
		for i := -1; i <= 1; i++ {
			for j := -1; j <= 1; j++ {
				x, y := location.X+j, location.Y+i
				if world.IsValid(x, y) && mountMap[c.GetVisionDistance()+i][c.GetVisionDistance()+j] == 0 {
					m, _ := world.GetCreature(x, y).(*Mount)
					m.AddRider(r)
					world.DeleteCreature(m)
					r.AddMount(m)
					c.Standup()
					return x, y
				}
			}
		}
	}

	// If adjacent to closed door attempt to open it
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			x, y := location.X+j, location.Y+i
			if world.IsValid(x, y) && world.IsDoor(x, y) && !world.IsPassable(x, y) {
				world.ToggleDoor(x, y, true)
				return cX, cY
			}
		}
	}

	if len(possibleLocations) > 0 {
		if itemHolder, ok := c.(holdsItems); ok && itemHolder.overEncumbered() {
			for _, itm := range itemHolder.inventory() {
				if itm.GetWeight() > 1 {
					itemHolder.dropItem(itm)
					return cX, cY
				}
			}
		} else if r, ok := c.(Rider); ok && (r.Mount() == nil || !r.Mount().Moved()) {
			l := possibleLocations[rand.Intn(len(possibleLocations))]
			return l.X, l.Y
		}
	} else if itemHolder, ok := c.(holdsItems); ok {
		if items := world.GetItems(location.X, location.Y); len(items) > 0 {
			for _, item := range items {
				itemHolder.pickupItem(item)
			}
		}
	}
	return cX, cY
}

func unmarshalAi(ai map[string]interface{}) ai {
	aiJson, err := json.Marshal(ai)
	check(err)

	switch ai["Type"] {
	case "mount":
		var mAi mountAi
		err = json.Unmarshal(aiJson, &mAi)
		check(err)
		return mAi
	case "npc":
		var nAi npcAi
		err = json.Unmarshal(aiJson, &nAi)
		check(err)
		return nAi
	case "enemy":
		var eAi enemyAi
		err = json.Unmarshal(aiJson, &eAi)
		check(err)
		return eAi
	}
	return nil
}

func (ai enemyAi) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	buffer.WriteString("\"Type\":\"enemy\"")
	buffer.WriteString("}")

	return buffer.Bytes(), nil
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

func getMountMap(c hasAi, world *worldmap.Map) [][]int {
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

func getChaseMap(c hasAi, world *worldmap.Map) [][]int {
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

func getItemMap(c hasAi, world *worldmap.Map) [][]int {
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

func getCoverMap(c hasAi, world *worldmap.Map) [][]int {
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
