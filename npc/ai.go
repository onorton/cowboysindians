package npc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/onorton/cowboysindians/event"
	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/structs"
	"github.com/onorton/cowboysindians/worldmap"
)

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
	AttackHits(int) bool
}

type ai interface {
	update(hasAi, *worldmap.Map) Action
}

func newAi(aiType string, world *worldmap.Map, location worldmap.Coordinates, town *worldmap.Town, building *worldmap.Building, dialogue dialogue) ai {
	switch aiType {
	case "animal":
		return animalAi{worldmap.NewRandomWaypoint(world, location)}
	case "npc":
		if building != nil {
			return npcAi{worldmap.NewWithinBuilding(world, *building, location)}
		} else {
			return npcAi{worldmap.NewRandomWaypoint(world, location)}
		}
	case "sheriff":
		return newSheriffAi(location, *town)
	case "enemy":
		return enemyAi{dialogue.(*enemyDialogue)}
	}
	return nil
}

type animalAi struct {
	waypoint worldmap.WaypointSystem
}

func (ai animalAi) update(c hasAi, world *worldmap.Map) Action {
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
		return MoveAction{c, world, l.X, l.Y}
	}

	return NoAction{}
}

func (ai animalAi) setMap(world *worldmap.Map) {
	switch w := ai.waypoint.(type) {
	case *worldmap.RandomWaypoint:
		w.SetMap(world)
	case *worldmap.Patrol:
	case *worldmap.WithinBuilding:
		w.SetMap(world)
	}
}

func (ai animalAi) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	buffer.WriteString("\"Type\":\"animal\",")

	waypointValue, err := json.Marshal(ai.waypoint)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Waypoint\":%s", waypointValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (ai *animalAi) UnmarshalJSON(data []byte) error {
	type animalAiJson struct {
		Waypoint map[string]interface{}
	}

	var v animalAiJson
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	ai.waypoint = worldmap.UnmarshalWaypointSystem(v.Waypoint)
	return nil
}

type npcAi struct {
	waypoint worldmap.WaypointSystem
}

func (ai npcAi) update(c hasAi, world *worldmap.Map) Action {

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
				for _, itm := range itemHolder.Inventory() {
					return DropAction{itemHolder, itm}
				}
			} else {
				l := possibleLocations[rand.Intn(len(possibleLocations))]
				return MountedMoveAction{r, world, l.X, l.Y}
			}
		}
	}

	// If at half health heal up
	if itemHolder, ok := c.(holdsItems); ok && c.bloodied() {
		for _, itm := range itemHolder.Inventory() {
			if consumable, ok := itm.Component("consumable").(item.ConsumableComponent); ok && len(consumable.Effects["hp"]) > 0 {
				return ConsumeAction{c, itm}
			}
		}
	}

	// If adjacent to mount, attempt to mount it
	if r, ok := c.(Rider); ok && r.Mount() == nil {
		for i := -1; i <= 1; i++ {
			for j := -1; j <= 1; j++ {
				x, y := location.X+j, location.Y+i
				if world.IsValid(x, y) && mountMap[c.GetVisionDistance()+i][c.GetVisionDistance()+j] == 0 {
					return MountAction{r, world, x, y}
				}
			}
		}
	}

	// If adjacent to closed door attempt to open it
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			x, y := location.X+j, location.Y+i
			if world.IsValid(x, y) && world.IsDoor(x, y) && !world.Door(x, y).Open() {
				return OpenAction{world, x, y}
			}
		}
	}

	if len(possibleLocations) > 0 {
		if itemHolder, ok := c.(holdsItems); ok && itemHolder.overEncumbered() {
			for _, itm := range itemHolder.Inventory() {
				return DropAction{itemHolder, itm}
			}
		} else if r, ok := c.(Rider); ok && (r.Mount() == nil || !r.Mount().Moved()) {
			l := possibleLocations[rand.Intn(len(possibleLocations))]
			return MoveAction{c, world, l.X, l.Y}
		}
	} else if itemHolder, ok := c.(holdsItems); ok {
		if world.HasItems(location.X, location.Y) {
			return PickupAction{itemHolder, world, location.X, location.Y}
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
		return MoveAction{c, world, l.X, l.Y}
	}

	return NoAction{}
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

type sheriffAi struct {
	waypoint *worldmap.Patrol
	t        worldmap.Town
	bounties *Bounties
}

func newSheriffAi(l worldmap.Coordinates, t worldmap.Town) *sheriffAi {
	// Patrol between ends of the town and sheriff's office
	points := make([]worldmap.Coordinates, 3)
	points[0] = l
	if t.Horizontal {
		points[1] = worldmap.Coordinates{t.TX1, (t.SY1 + t.SY2) / 2}
		points[2] = worldmap.Coordinates{t.TX2, (t.SY1 + t.SY2) / 2}
	} else {
		points[1] = worldmap.Coordinates{(t.SX1 + t.SX2) / 2, t.SY1}
		points[2] = worldmap.Coordinates{(t.SX1 + t.SX2) / 2, t.SY1}
	}
	ai := &sheriffAi{worldmap.NewPatrol(points), t, &Bounties{}}
	event.Subscribe(ai)
	return ai
}

func (ai sheriffAi) ProcessEvent(e event.Event) {
	switch ev := e.(type) {
	case event.WitnessedCrimeEvent:
		{
			crime := ev.Crime
			location := crime.Location()
			if location.X >= ai.t.TX1 && location.X <= ai.t.TX2 && location.Y >= ai.t.TY1 && location.Y <= ai.t.TY2 {
				ai.bounties.addBounty(crime)
			}
		}
	}

}

func (ai sheriffAi) update(c hasAi, world *worldmap.Map) Action {

	cX, cY := c.GetCoordinates()
	location := worldmap.Coordinates{cX, cY}
	waypoint := ai.waypoint.NextWaypoint(location)

	targets := append(getEnemies(c, world), visibleBounties(c, world, ai.bounties)...)

	coefficients := []float64{0.2, 0.5, 0.3, 0.0}

	// Focus on getting a mount if possible
	if r, ok := c.(Rider); ok && r.Mount() == nil {
		coefficients = []float64{0.1, 0.3, 0.2, 0.4}
	}
	coverMap := getCoverMap(c, world, targets)
	mountMap := getMountMap(c, world)
	aiMap := addMaps([][][]int{getChaseMap(c, world, targets), getWaypointMap(waypoint, world, location, c.GetVisionDistance()), coverMap, mountMap}, coefficients)

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
				for _, itm := range itemHolder.Inventory() {
					return DropAction{itemHolder, itm}
				}
			} else {
				l := possibleLocations[rand.Intn(len(possibleLocations))]
				return MountedMoveAction{r, world, l.X, l.Y}
			}
		}
	}

	// If at half health heal up
	if itemHolder, ok := c.(holdsItems); ok && c.bloodied() {
		for _, itm := range itemHolder.Inventory() {
			if consumable, ok := itm.Component("consumable").(item.ConsumableComponent); ok && len(consumable.Effects["hp"]) > 0 {
				return ConsumeAction{c, itm}
			}
		}
	}

	// If moving into or out of cover and not mounted toggle crouch
	if r, ok := c.(Rider); ok && r.Mount() == nil {
		if coverMap[c.GetVisionDistance()][c.GetVisionDistance()] == 0 && !c.IsCrouching() {
			return CrouchAction{c}
		} else if coverMap[c.GetVisionDistance()][c.GetVisionDistance()] > 0 && c.IsCrouching() {
			return StandupAction{c}
		}
	}

	// Try and wield best weapon
	if itemUser, ok := c.(usesItems); ok && itemUser.wieldItem() {
		return NoAction{}
	}
	// Try and wear best armour
	if itemUser, ok := c.(usesItems); ok && itemUser.wearArmour() {
		return NoAction{}
	}

	if len(targets) > 0 {
		closestTarget := targets[0]
		tX, tY := targets[0].GetCoordinates()
		min := worldmap.Distance(location.X, location.Y, tX, tY)

		for _, e := range targets {
			tX, tY = e.GetCoordinates()
			d := worldmap.Distance(location.X, location.Y, tX, tY)
			if d < min {
				min = d
				closestTarget = e
			}
		}

		if itemUser, ok := c.(usesItems); ok {
			if itemUser.ranged() && min < float64(itemUser.Weapon().Range) {

				// if weapon loaded, shoot at target else if enemy has ammo, load weapon
				if itemUser.weaponLoaded() {
					return RangedAttackAction{c, world, closestTarget}
				} else if itemUser.hasAmmo() {
					return LoadAction{itemUser}
				}
			}
		}
	}

	// If adjacent to mount, attempt to mount it
	if r, ok := c.(Rider); ok && r.Mount() == nil {
		for i := -1; i <= 1; i++ {
			for j := -1; j <= 1; j++ {
				x, y := location.X+j, location.Y+i
				if world.IsValid(x, y) && mountMap[c.GetVisionDistance()+i][c.GetVisionDistance()+j] == 0 {
					return MountAction{r, world, x, y}
				}
			}
		}
	}

	// If adjacent to closed door attempt to open it
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			x, y := location.X+j, location.Y+i
			if world.IsValid(x, y) && world.IsDoor(x, y) && !world.Door(x, y).Open() {
				return OpenAction{world, x, y}
			}
		}
	}

	if len(possibleLocations) > 0 {
		if itemHolder, ok := c.(holdsItems); ok && itemHolder.overEncumbered() {
			for _, itm := range itemHolder.Inventory() {
				return DropAction{itemHolder, itm}
			}
		} else if r, ok := c.(Rider); ok && (r.Mount() == nil || !r.Mount().Moved()) {
			l := possibleLocations[rand.Intn(len(possibleLocations))]
			return MoveAction{c, world, l.X, l.Y}
		}
	} else if itemHolder, ok := c.(holdsItems); ok {
		if world.HasItems(location.X, location.Y) {
			return PickupAction{itemHolder, world, location.X, location.Y}
		}
	}

	return NoAction{}
}

func (ai sheriffAi) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	buffer.WriteString("\"Type\":\"sheriff\",")

	waypointValue, err := json.Marshal(ai.waypoint)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Waypoint\":%s,", waypointValue))

	townValue, err := json.Marshal(ai.t)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Town\":%s,", townValue))

	bountiesValue, err := json.Marshal(ai.bounties)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Bounties\":%s", bountiesValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (ai *sheriffAi) UnmarshalJSON(data []byte) error {
	type sheriffAiJson struct {
		Waypoint *worldmap.Patrol
		Town     worldmap.Town
		Bounties *Bounties
	}

	var v sheriffAiJson
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	ai.waypoint = v.Waypoint
	ai.t = v.Town
	ai.bounties = v.Bounties

	event.Subscribe(ai)
	return nil
}

type enemyAi struct {
	dialogue *enemyDialogue
}

func (ai enemyAi) update(c hasAi, world *worldmap.Map) Action {
	target := world.GetPlayer()
	tX, tY := target.GetCoordinates()

	if world.InConversationRange(c.(worldmap.Creature), target) {
		ai.dialogue.initialGreeting()
	}

	cX, cY := c.GetCoordinates()
	location := worldmap.Coordinates{cX, cY}

	coefficients := []float64{0.5, 0.2, 0.3, 0.0}

	// Focus on getting a mount if possible
	if r, ok := c.(Rider); ok && r.Mount() == nil {
		coefficients = []float64{0.3, 0.2, 0.1, 0.4}
	}
	coverMap := getCoverMap(c, world, []worldmap.Creature{world.GetPlayer()})
	mountMap := getMountMap(c, world)
	aiMap := addMaps([][][]int{getChaseMap(c, world, []worldmap.Creature{world.GetPlayer()}), getItemMap(c, world), coverMap, mountMap}, coefficients)

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
				for _, itm := range itemHolder.Inventory() {
					return DropAction{itemHolder, itm}
				}
			} else {
				l := possibleLocations[rand.Intn(len(possibleLocations))]
				if l == (worldmap.Coordinates{tX, tY}) {
					ai.dialogue.potentiallyThreaten()
				}
				return MountedMoveAction{r, world, l.X, l.Y}
			}
		}
	}

	// If at half health heal up
	if itemHolder, ok := c.(holdsItems); ok && c.bloodied() {
		for _, itm := range itemHolder.Inventory() {
			if consumable, ok := itm.Component("consumable").(item.ConsumableComponent); ok && len(consumable.Effects["hp"]) > 0 {
				return ConsumeAction{c, itm}
			}
		}
	}

	// If moving into or out of cover and not mounted toggle crouch
	if r, ok := c.(Rider); ok && r.Mount() == nil {
		if coverMap[c.GetVisionDistance()][c.GetVisionDistance()] == 0 && !c.IsCrouching() {
			return CrouchAction{c}
		} else if coverMap[c.GetVisionDistance()][c.GetVisionDistance()] > 0 && c.IsCrouching() {
			return StandupAction{c}
		}
	}

	// Try and wield best weapon
	if itemUser, ok := c.(usesItems); ok && itemUser.wieldItem() {
		return NoAction{}
	}
	// Try and wear best armour
	if itemUser, ok := c.(usesItems); ok && itemUser.wearArmour() {
		return NoAction{}
	}

	if itemUser, ok := c.(usesItems); ok {
		if distance := worldmap.Distance(location.X, location.Y, tX, tY); itemUser.ranged() && distance < float64(itemUser.Weapon().Range) && world.IsVisible(c, tX, tY) {

			ai.dialogue.potentiallyThreaten()
			// if weapon loaded, shoot at target else if enemy has ammo, load weapon
			if itemUser.weaponLoaded() {
				return RangedAttackAction{c, world, target}
			} else if itemUser.hasAmmo() {
				return LoadAction{itemUser}
			}
		}

	}

	// If adjacent to mount, attempt to mount it
	if r, ok := c.(Rider); ok && r.Mount() == nil {
		for i := -1; i <= 1; i++ {
			for j := -1; j <= 1; j++ {
				x, y := location.X+j, location.Y+i
				if world.IsValid(x, y) && mountMap[c.GetVisionDistance()+i][c.GetVisionDistance()+j] == 0 {
					return MountAction{r, world, x, y}
				}
			}
		}
	}

	// If adjacent to closed door attempt to open it
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			x, y := location.X+j, location.Y+i
			if world.IsValid(x, y) && world.IsDoor(x, y) && !world.Door(x, y).Open() {
				return OpenAction{world, x, y}
			}
		}
	}

	if len(possibleLocations) > 0 {
		if itemHolder, ok := c.(holdsItems); ok && itemHolder.overEncumbered() {
			for _, itm := range itemHolder.Inventory() {
				return DropAction{itemHolder, itm}
			}
		} else if r, ok := c.(Rider); ok && (r.Mount() == nil || !r.Mount().Moved()) {
			l := possibleLocations[rand.Intn(len(possibleLocations))]
			if l == (worldmap.Coordinates{tX, tY}) {
				ai.dialogue.potentiallyThreaten()
			}
			return MoveAction{c, world, l.X, l.Y}
		}
	} else if itemHolder, ok := c.(holdsItems); ok {
		if world.HasItems(location.X, location.Y) {
			return PickupAction{itemHolder, world, location.X, location.Y}
		}
	}
	return NoAction{}
}

func unmarshalAi(ai map[string]interface{}) ai {
	aiJson, err := json.Marshal(ai)
	check(err)

	switch ai["Type"] {
	case "animal":
		var mAi animalAi
		err = json.Unmarshal(aiJson, &mAi)
		check(err)
		return mAi
	case "npc":
		var nAi npcAi
		err = json.Unmarshal(aiJson, &nAi)
		check(err)
		return nAi
	case "sheriff":
		var sAi sheriffAi
		err = json.Unmarshal(aiJson, &sAi)
		check(err)
		return sAi
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

	buffer.WriteString("\"Type\":\"enemy\",")

	dialogueValue, err := json.Marshal(ai.dialogue)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Dialogue\":%s", dialogueValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (ai *enemyAi) UnmarshalJSON(data []byte) error {
	type enemyAiJson struct {
		Dialogue map[string]interface{}
	}

	var v enemyAiJson
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	ai.dialogue = unmarshalDialogue(v.Dialogue).(*enemyDialogue)
	return nil
}

func generateMap(world *worldmap.Map, goals []worldmap.Coordinates, location worldmap.Coordinates, d int) [][]int {
	visitedNodes := make(map[worldmap.Coordinates]int)

	type nodeValue struct {
		location worldmap.Coordinates
		value    int
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

	aiMap := make([][]int, width)

	for y := 0; y < width; y++ {
		aiMap[y] = make([]int, width)
		for x := 0; x < width; x++ {

			if v, ok := visitedNodes[worldmap.Coordinates{x - d + location.X, y - d + location.Y}]; ok {
				aiMap[y][x] = v
			} else {
				aiMap[y][x] = width * width
			}
		}
	}
	return aiMap
}

func getWaypointMap(waypoint worldmap.Coordinates, world *worldmap.Map, location worldmap.Coordinates, d int) [][]int {
	return generateMap(world, []worldmap.Coordinates{waypoint}, location, d)
}

func getMountMap(c hasAi, world *worldmap.Map) [][]int {
	d := c.GetVisionDistance()
	cX, cY := c.GetCoordinates()
	location := worldmap.Coordinates{cX, cY}

	mountLocations := make([]worldmap.Coordinates, 0)

	for i := 0; i < d+1; i++ {
		for j := -d; j < d+1; j++ {
			// Translate location into world coordinates
			wX, wY := location.X+j, location.Y+i
			// Looks for mount on its own
			if world.IsValid(wX, wY) && world.IsVisible(c, wX, wY) {
				if m, ok := world.GetCreature(wX, wY).(*Mount); ok && m != nil {
					mountLocations = append(mountLocations, worldmap.Coordinates{wX, wY})
				}
			}
		}
	}
	return generateMap(world, mountLocations, location, d)
}

func getChaseMap(c hasAi, world *worldmap.Map, targets []worldmap.Creature) [][]int {
	d := c.GetVisionDistance()
	cX, cY := c.GetCoordinates()
	location := worldmap.Coordinates{cX, cY}

	targetLocations := make([]worldmap.Coordinates, 0)

	for _, t := range targets {
		x, y := t.GetCoordinates()
		if world.IsVisible(c, x, y) {
			targetLocations = append(targetLocations, worldmap.Coordinates{x, y})
		}
	}

	return generateMap(world, targetLocations, location, d)
}

func getItemMap(c hasAi, world *worldmap.Map) [][]int {
	d := c.GetVisionDistance()
	cX, cY := c.GetCoordinates()
	location := worldmap.Coordinates{cX, cY}

	itemLocations := make([]worldmap.Coordinates, 0)

	for i := -d; i < d+1; i++ {
		for j := -d; j < d+1; j++ {
			// Translate location into world coordinates
			wX, wY := location.X+j, location.Y+i
			if world.IsValid(wX, wY) && world.IsVisible(c, wX, wY) && world.HasItems(wX, wY) {
				itemLocations = append(itemLocations, worldmap.Coordinates{wX, wY})
			}
		}
	}
	return generateMap(world, itemLocations, location, d)
}

func getCoverMap(c hasAi, world *worldmap.Map, targets []worldmap.Creature) [][]int {
	d := c.GetVisionDistance()
	cX, cY := c.GetCoordinates()
	location := worldmap.Coordinates{cX, cY}

	player := world.GetPlayer()
	pX, pY := player.GetCoordinates()

	coverLocations := make([]worldmap.Coordinates, 0)

	for i := -d; i < d+1; i++ {
		for j := -d; j < d+1; j++ {

			// Translate location into world coordinates
			wX, wY := location.X+j, location.Y+i
			// Enemy must be able to see player in order to know it would be behind cover
			if world.IsValid(wX, wY) && world.IsVisible(c, wX, wY) && world.IsVisible(c, pX, pY) && world.BehindCover(wX, wY, player) {
				for _, t := range targets {
					tX, tY := t.GetCoordinates()
					if world.IsVisible(c, tX, tY) && world.BehindCover(wX, wY, t) {
						coverLocations = append(coverLocations, worldmap.Coordinates{wX, wY})
						break
					}
				}
			}
		}
	}
	return generateMap(world, coverLocations, location, d)
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

func visibleBounties(c hasAi, world *worldmap.Map, bounties *Bounties) []worldmap.Creature {
	d := c.GetVisionDistance()
	cX, cY := c.GetCoordinates()
	location := worldmap.Coordinates{cX, cY}

	targets := make([]worldmap.Creature, 0)

	for i := -d; i < d+1; i++ {
		for j := -d; j < d+1; j++ {
			// Translate location into world coordinates
			wX, wY := location.X+j, location.Y+i
			if world.IsValid(wX, wY) && world.IsVisible(c, wX, wY) && world.GetCreature(wX, wY) != nil && bounties.hasBounty(world.GetCreature(wX, wY).GetID()) {
				targets = append(targets, world.GetCreature(wX, wY))
			}
		}
	}
	return targets
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
