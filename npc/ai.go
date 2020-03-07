package npc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"sort"

	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/structs"
	"github.com/onorton/cowboysindians/worldmap"
)

type aiAttributes struct {
	Senses  []map[string]interface{}
	Actions []map[string]interface{}
}

var aiData = fetchAiData()

func fetchAiData() map[string]aiAttributes {
	data, err := ioutil.ReadFile("data/ai.json")
	check(err)
	var aiD map[string]aiAttributes
	err = json.Unmarshal(data, &aiD)
	check(err)
	return aiD
}

type hasAi interface {
	consume(*item.Item)
	damageable
	worldmap.CanSee
	worldmap.CanCrouch
}

type holdsItems interface {
	dropItem(*item.Item)
	PickupItem(*item.Item)
	Inventory() []*item.Item
	overEncumbered() bool
	maximumLift() float64
	RemoveItem(*item.Item)
}

type usesItems interface {
	wieldItem() bool
	wearArmour() bool
	ranged() bool
	rangedAttack(worldmap.Creature, int)
	Weapon() item.WeaponComponent
	weaponLoaded() bool
	weaponFullyLoaded() bool
	hasAmmo() bool
	getAmmo() *item.Item
}

type damageable interface {
	bloodied() bool
	hp() *worldmap.Attribute
	AttackHits(int) bool
}

type ai interface {
	update(hasAi, *worldmap.Map) Action
}

func newAi(aiType string, id string, world *worldmap.Map, location worldmap.Coordinates, town *worldmap.Town, building *worldmap.Building, dialogue dialogue, protectee *string) ai {

	switch aiType {
	case "animal", "aggressive animal", "npc", "farmer", "sheriff", "bar patron", "enemy":
		return newGenericAi(aiType, id, dialogue, location, town, building, world, protectee)
	case "protector":
		if protectee != nil {
			return newGenericAi(aiType, id, dialogue, location, town, building, world, protectee)
		} else {
			return newGenericAi("npc", id, dialogue, location, town, building, world, protectee)
		}
	}
	return nil
}

type genericAi struct {
	sensory []senses
	actions []hasAction
	state   *string
}

func newGenericAi(aiType string, id string, dialogue dialogue, l worldmap.Coordinates, t *worldmap.Town, b *worldmap.Building, world *worldmap.Map, protectee *string) genericAi {
	state := "normal"

	otherData := make(map[string]interface{})
	otherData["location"] = l
	otherData["creatureID"] = id
	otherData["protecteeID"] = protectee
	otherData["town"] = t
	otherData["building"] = b
	otherData["world"] = world
	otherData["dialogue"] = dialogue

	sensory := make([]senses, 0)
	for _, s := range aiData[aiType].Senses {
		sensory = append(sensory, newSensesComponent(s, otherData))
	}

	actions := make([]hasAction, 0)
	for _, a := range aiData[aiType].Actions {
		actions = append(actions, newActionComponent(a, otherData))
	}

	return genericAi{sensory, actions, &state}
}

func (ai genericAi) setMap(world *worldmap.Map) {
	for _, a := range ai.actions {
		if waypoint, ok := a.(waypointComponent); ok {
			switch w := waypoint.waypoint.(type) {
			case *worldmap.RandomWaypoint:
				w.SetMap(world)
			case *worldmap.Patrol:
			case *worldmap.WithinArea:
				w.SetMap(world)
			}
		}
	}
}

func (ai genericAi) update(c hasAi, world *worldmap.Map) Action {
	threats := make([]worldmap.Creature, 0)
	for _, s := range ai.sensory {
		if sThreats, ok := s.(sensesThreats); ok {
			threats = append(threats, sThreats.threats(c, world)...)
		}
	}

	targets := getEnemies(c, world)
	for _, s := range ai.sensory {
		if sTargets, ok := s.(sensesTargets); ok {
			targets = append(targets, sTargets.targets(c, world)...)
		}
	}

	ai.nextState(c, world)

	for _, a := range ai.actions {
		if aTargets, ok := a.(hasTargets); ok {
			aTargets.addTargets(targets)
		}

		if aThreats, ok := a.(hasThreats); ok {
			aThreats.addThreats(threats)
		}
	}

	return ai.nextAction(c, world)
}

func (ai genericAi) nextState(c hasAi, world *worldmap.Map) {
	stateCounts := make(map[string]int)

	for _, sensory := range ai.sensory {
		state := sensory.nextState(*ai.state, c, world)
		if state != "" {
			stateCounts[state]++
		}
	}

	max := 0
	for _, count := range stateCounts {
		if count > max {
			max = count
		}
	}

	states := make([]string, 0)
	for state, count := range stateCounts {
		if count == max {
			states = append(states, state)
		}
	}

	if len(states) > 0 {
		// Pick random state if there is a tie
		*ai.state = states[rand.Intn(len(states))]
	}
}

func (ai genericAi) nextAction(c hasAi, world *worldmap.Map) Action {
	actions := ai.actions

	sort.Slice(actions, func(i, j int) bool {
		return actions[i].shouldHappen(*ai.state) > actions[j].shouldHappen(*ai.state)
	})

	for _, a := range actions {
		if a.shouldHappen(*ai.state) == 0 {
			break
		}

		if action := a.action(c, world); action != nil {
			return action
		}
	}
	return NoAction{}
}

func (ai genericAi) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	buffer.WriteString("\"Type\":\"generic\",")

	sensoryValue, err := json.Marshal(ai.sensory)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Senses\":%s,", sensoryValue))

	actionsValue, err := json.Marshal(ai.actions)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Actions\":%s,", actionsValue))

	stateValue, err := json.Marshal(ai.state)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"State\":%s", stateValue))

	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (ai *genericAi) UnmarshalJSON(data []byte) error {
	type genericAiJson struct {
		Senses  []map[string]interface{}
		Actions []map[string]interface{}
		State   *string
	}

	var v genericAiJson
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	ai.sensory = unmarshalSenses(v.Senses)
	ai.actions = unmarshalActions(v.Actions)
	ai.state = v.State

	return nil
}

func unmarshalAi(ai map[string]interface{}) ai {
	aiJson, err := json.Marshal(ai)
	check(err)

	switch ai["Type"] {
	case "generic":
		var sAi genericAi
		err = json.Unmarshal(aiJson, &sAi)
		check(err)
		return sAi
	}
	return nil
}

func possibleLocationsFromAiMap(c hasAi, world *worldmap.Map, aiMap [][]float64, tileValid func(int, int) bool) []worldmap.Coordinates {
	d := c.GetVisionDistance()
	cX, cY := c.GetCoordinates()
	possibleLocations := make([]worldmap.Coordinates, 0)
	current := aiMap[d][d]

	// Find adjacent locations closer to the goal
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if aiMap[j+d][i+d] < current {
				x := cX + i
				y := cY + j
				if world.IsValid(x, y) && (tileValid(x, y)) {
					possibleLocations = append(possibleLocations, worldmap.Coordinates{x, y})
				}
			}
		}
	}

	return possibleLocations
}

func healIfWeak(c hasAi) Action {
	// If at half health heal up
	if itemHolder, ok := c.(holdsItems); ok && c.bloodied() {
		for _, itm := range itemHolder.Inventory() {
			if consumable, ok := itm.Component("consumable").(item.ConsumableComponent); ok && len(consumable.Effects["hp"]) > 0 {
				return ConsumeAction{c, itm}
			}
		}
	}
	return nil
}

func moveThroughCover(c hasAi, coverMap [][]float64) Action {
	// If moving into or out of cover and not mounted toggle crouch
	if r, ok := c.(Rider); !ok || r.Mount() == nil {
		d := c.GetVisionDistance()
		if coverMap[d][d] == 0 && !c.IsCrouching() {
			return CrouchAction{c}
		} else if coverMap[d][d] > 0 && c.IsCrouching() {
			return StandupAction{c}
		}
	}
	return nil
}

func tryOpeningDoor(c hasAi, world *worldmap.Map) Action {
	// If adjacent to closed door attempt to open it
	cX, cY := c.GetCoordinates()
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			x, y := cX+j, cY+i
			if world.IsValid(x, y) && world.IsDoor(x, y) && !world.Door(x, y).Open() {
				if itemHolder, ok := c.(holdsItems); ok && world.Door(x, y).Locked() {
					// If there is a key that fits, unlock door
					for _, itm := range itemHolder.Inventory() {
						if itm.HasComponent("key") && world.Door(x, y).KeyFits(itm) {
							return LockAction{itm, world, x, y}
						}
					}
				}
				return OpenAction{world, x, y}
			}
		}
	}
	return nil
}

func mount(c hasAi, world *worldmap.Map, mountMap [][]float64) Action {
	// If adjacent to mount, attempt to mount it
	cX, cY := c.GetCoordinates()
	if r, ok := c.(Rider); ok && r.Mount() == nil {
		for i := -1; i <= 1; i++ {
			for j := -1; j <= 1; j++ {
				x, y := cX+j, cY+i
				if world.IsValid(x, y) && mountMap[c.GetVisionDistance()+i][c.GetVisionDistance()+j] == 0 {
					return MountAction{r, world, x, y}
				}
			}
		}
	}
	return nil
}

func moveIfMounted(c hasAi, world *worldmap.Map, locations []worldmap.Coordinates) Action {
	// If mounted, can move first before executing another action
	if r, ok := c.(Rider); ok && r.Mount() != nil && !r.Mount().Moved() {
		if len(locations) > 0 {
			if itemHolder, ok := c.(holdsItems); ok && itemHolder.overEncumbered() {
				for _, itm := range itemHolder.Inventory() {
					return DropAction{itemHolder, itm}
				}
			} else {
				l := locations[rand.Intn(len(locations))]
				return MountedMoveAction{r, world, l.X, l.Y}
			}
		}
	}
	return nil
}

func rangedAttack(c hasAi, world *worldmap.Map, targets []worldmap.Creature) Action {
	if itemUser, ok := c.(usesItems); ok {
		if len(targets) > 0 {
			cX, cY := c.GetCoordinates()
			closestTarget := targets[0]
			tX, tY := targets[0].GetCoordinates()
			min := worldmap.Distance(cX, cY, tX, tY)

			for _, e := range targets {
				tX, tY = e.GetCoordinates()
				d := worldmap.Distance(cX, cY, tX, tY)
				if d < min {
					min = d
					closestTarget = e
				}
			}

			tX, tY = closestTarget.GetCoordinates()

			if itemUser.ranged() && min < float64(itemUser.Weapon().Range) && world.IsVisible(c, tX, tY) {
				// if weapon loaded, shoot at target else if enemy has ammo, load weapon
				if itemUser.weaponLoaded() {
					return RangedAttackAction{c, world, closestTarget}
				} else if itemUser.hasAmmo() {
					return LoadAction{itemUser}
				}
			}
		}
	}
	return nil
}

func move(c hasAi, world *worldmap.Map, locations []worldmap.Coordinates) Action {
	if len(locations) > 0 {
		if itemHolder, ok := c.(holdsItems); ok && itemHolder.overEncumbered() {
			for _, itm := range itemHolder.Inventory() {
				return DropAction{itemHolder, itm}
			}
		} else if r, ok := c.(Rider); !ok || (r.Mount() == nil || !r.Mount().Moved()) {
			l := locations[rand.Intn(len(locations))]
			return MoveAction{c, world, l.X, l.Y}
		}
	}
	return nil
}

func moveRandomly(c hasAi, world *worldmap.Map) Action {
	possibleLocations := make([]worldmap.Coordinates, 0)
	cX, cY := c.GetCoordinates()
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			x, y := cX+j, cY+i
			if world.IsValid(x, y) && world.IsPassable(x, y) && !world.IsOccupied(x, y) {
				possibleLocations = append(possibleLocations, worldmap.Coordinates{x, y})
			}
		}
	}

	return move(c, world, possibleLocations)
}

func pickupItems(c hasAi, world *worldmap.Map) Action {
	cX, cY := c.GetCoordinates()
	if itemHolder, ok := c.(holdsItems); ok {
		if world.HasItems(cX, cY) {
			return PickupAction{itemHolder, world, cX, cY}
		}
	}
	return nil
}

func generateMap(c hasAi, world *worldmap.Map, goals []worldmap.Coordinates) [][]float64 {
	d := c.GetVisionDistance()
	cX, cY := c.GetCoordinates()
	location := worldmap.Coordinates{cX, cY}

	visitedNodes := make(map[worldmap.Coordinates]float64)

	type nodeValue struct {
		location worldmap.Coordinates
		value    float64
	}

	nodes := structs.Queue{}
	for _, goal := range goals {
		nodes.Enqueue(nodeValue{goal, 0})
	}

	npcFound := false
	width := 2*d + 1

	for !nodes.IsEmpty() {
		node := nodes.Dequeue().(nodeValue)

		if _, ok := visitedNodes[node.location]; !ok {
			if node.location == location {
				npcFound = true
			}
			visitedNodes[node.location] = node.value
			if !npcFound {
				// Add adjacent
				for i := -1; i <= 1; i++ {
					for j := -1; j <= 1; j++ {
						x, y := node.location.X+i, node.location.Y+j
						aiX, aiY := x+d-location.X, y+d-location.Y
						if aiX >= 0 && aiX <= width && aiY >= 0 && aiY <= width && world.IsValid(x, y) && world.IsPassable(x, y) && !(i == 0 && j == 0) {
							nodes.Enqueue(nodeValue{worldmap.Coordinates{x, y}, node.value + 1})
						}
					}
				}
			}
		}
	}

	aiMap := make([][]float64, width)

	for y := 0; y < width; y++ {
		aiMap[y] = make([]float64, width)
		for x := 0; x < width; x++ {

			if v, ok := visitedNodes[worldmap.Coordinates{x - d + location.X, y - d + location.Y}]; ok {
				aiMap[y][x] = v
			} else {
				aiMap[y][x] = float64(width * width)
			}
		}
	}
	return aiMap
}

func getWaypointMap(c hasAi, waypoint worldmap.Coordinates, world *worldmap.Map) [][]float64 {
	d := float64(c.GetVisionDistance())
	cX, cY := c.GetCoordinates()
	dX := float64(waypoint.X - cX)
	dY := float64(waypoint.Y - cY)
	if math.Abs(dX) > d || math.Abs(dY) > d {
		distance := float64(worldmap.Distance(cX, cY, waypoint.X, waypoint.Y))
		// If not within vision distance, pick point within vision distance in that direction
		newX, newY := int(float64(cX)+dX*(d/distance)), int(float64(cY)+dY*(d/distance))
		waypoint = worldmap.Coordinates{newX, newY}
	}

	return generateMap(c, world, []worldmap.Coordinates{waypoint})
}

func getMountMap(c hasAi, world *worldmap.Map) [][]float64 {
	tileHasMount := func(x, y int) bool {
		if world.IsValid(x, y) && world.IsVisible(c, x, y) {
			m, ok := world.GetCreature(x, y).(*Mount)
			return ok && m != nil
		}
		return false
	}

	return getMap(c, world, tileHasMount)
}

func getChaseMap(c hasAi, world *worldmap.Map, targets []worldmap.Creature) [][]float64 {
	targetLocations := make([]worldmap.Coordinates, 0)

	for _, t := range targets {
		x, y := t.GetCoordinates()
		if world.IsVisible(c, x, y) {
			targetLocations = append(targetLocations, worldmap.Coordinates{x, y})
		}
	}

	return generateMap(c, world, targetLocations)
}

func getFleeMap(c hasAi, world *worldmap.Map, threats []worldmap.Creature) [][]float64 {
	// Fleep map is just the chase map inverted
	fleeMap := getChaseMap(c, world, threats)

	for y := 0; y < len(fleeMap); y++ {
		for x := 0; x < len(fleeMap[0]); x++ {
			fleeMap[y][x] = -fleeMap[y][x]
		}
	}
	return fleeMap

}

func getItemMap(c hasAi, world *worldmap.Map) [][]float64 {
	tileHasItems := func(x, y int) bool {
		return world.IsValid(x, y) && world.IsVisible(c, x, y) && world.HasItems(x, y)
	}

	return getMap(c, world, tileHasItems)
}

func getCoverMap(c hasAi, world *worldmap.Map, targets []worldmap.Creature) [][]float64 {
	tileWouldGiveCover := func(x, y int) bool {
		if world.IsValid(x, y) && world.IsVisible(c, x, y) {
			for _, t := range targets {
				tX, tY := t.GetCoordinates()
				// Creature must be able to see target in order to know it would be behind cover
				return world.IsVisible(c, tX, tY) && world.BehindCover(x, y, t)
			}
		}
		return false
	}

	return getMap(c, world, tileWouldGiveCover)
}

func getMap(c hasAi, world *worldmap.Map, tileValid func(int, int) bool) [][]float64 {
	d := c.GetVisionDistance()
	cX, cY := c.GetCoordinates()
	location := worldmap.Coordinates{cX, cY}
	locations := make([]worldmap.Coordinates, 0)

	for i := -d; i < d+1; i++ {
		for j := -d; j < d+1; j++ {
			// Translate location into world coordinates
			wX, wY := location.X+j, location.Y+i
			if tileValid(wX, wY) {
				locations = append(locations, worldmap.Coordinates{wX, wY})
			}
		}
	}

	return generateMap(c, world, locations)
}

func addMaps(maps [][][]float64, weights []float64) [][]float64 {
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

func visibleCreatures(c hasAi, world *worldmap.Map) []worldmap.Creature {
	d := c.GetVisionDistance()
	cX, cY := c.GetCoordinates()
	location := worldmap.Coordinates{cX, cY}

	creatures := make([]worldmap.Creature, 0)

	for i := -d; i < d+1; i++ {
		for j := -d; j < d+1; j++ {
			// Translate location into world coordinates
			wX, wY := location.X+j, location.Y+i
			if !(wX == cX && wY == cY) && world.IsValid(wX, wY) && world.IsVisible(c, wX, wY) && world.GetCreature(wX, wY) != nil {
				creatures = append(creatures, world.GetCreature(wX, wY))
			}
		}
	}

	return creatures
}

func getEnemies(c hasAi, world *worldmap.Map) []worldmap.Creature {
	d := c.GetVisionDistance()
	cX, cY := c.GetCoordinates()
	location := worldmap.Coordinates{cX, cY}

	enemies := make([]worldmap.Creature, 0)

	for i := -d; i < d+1; i++ {
		for j := -d; j < d+1; j++ {
			// Translate location into world coordinates
			wX, wY := location.X+j, location.Y+i
			if world.IsValid(wX, wY) && world.IsVisible(c, wX, wY) && world.GetCreature(wX, wY) != nil && world.GetCreature(wX, wY).GetAlignment() == worldmap.Enemy {
				enemies = append(enemies, world.GetCreature(wX, wY))
			}
		}
	}

	return enemies
}
