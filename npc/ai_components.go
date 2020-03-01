package npc

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/onorton/cowboysindians/event"
	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/structs"
	"github.com/onorton/cowboysindians/worldmap"
)

type senses interface {
	nextState(string, hasAi, *worldmap.Map) string
}

type sensesTargets interface {
	targets(hasAi, *worldmap.Map) []worldmap.Creature
}

type sensesThreats interface {
	threats(hasAi, *worldmap.Map) []worldmap.Creature
}

type bountiesComponent struct {
	t        worldmap.Town
	bounties *Bounties
}

func (c bountiesComponent) ProcessEvent(e event.Event) {
	switch ev := e.(type) {
	case event.WitnessedCrimeEvent:
		{
			crime := ev.Crime
			location := crime.Location()
			if location.X >= c.t.TownArea.X1() && location.X <= c.t.TownArea.X2() && location.Y >= c.t.TownArea.Y1() && location.Y <= c.t.TownArea.Y2() {
				c.bounties.addBounty(crime)
			}
		}
	}

}

func (c bountiesComponent) targets(ai hasAi, world *worldmap.Map) []worldmap.Creature {
	d := ai.GetVisionDistance()
	aiX, aiY := ai.GetCoordinates()

	targets := make([]worldmap.Creature, 0)

	for i := -d; i < d+1; i++ {
		for j := -d; j < d+1; j++ {
			// Translate location into world coordinates
			wX, wY := aiX+j, aiY+i
			if world.IsValid(wX, wY) && world.IsVisible(ai, wX, wY) && world.GetCreature(wX, wY) != nil && c.bounties.hasBounty(world.GetCreature(wX, wY).GetID()) {
				targets = append(targets, world.GetCreature(wX, wY))
			}
		}
	}

	return targets
}

func (c bountiesComponent) nextState(currState string, ai hasAi, world *worldmap.Map) string {
	if currState == "normal" && len(c.targets(ai, world)) > 0 {
		return "fighting"
	}

	if currState == "fighting" && len(c.targets(ai, world)) > 0 {
		return "normal"
	}

	return currState
}

func (c bountiesComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	buffer.WriteString("\"Type\": \"bounties\",")

	townValue, err := json.Marshal(c.t)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Town\":%s,", townValue))

	bountiesValue, err := json.Marshal(c.bounties)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Bounties\":%s", bountiesValue))

	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (c *bountiesComponent) UnmarshalJSON(data []byte) error {
	type bountiesJSON struct {
		Town     worldmap.Town
		Bounties *Bounties
	}

	var v bountiesJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	c.t = v.Town
	c.bounties = v.Bounties

	return nil
}

type threatsComponent struct {
	possibleThreats *structs.Set
	creatureID      string
}

func (c threatsComponent) ProcessEvent(e event.Event) {
	switch ev := e.(type) {
	case event.AttackEvent:
		{
			if ev.Victim().GetID() == c.creatureID {
				c.possibleThreats.Add(ev.Perpetrator().GetID())
			}
		}
	}

}

func (c threatsComponent) threats(ai hasAi, world *worldmap.Map) []worldmap.Creature {
	d := ai.GetVisionDistance()
	aiX, aiY := ai.GetCoordinates()

	visibleThreats := make([]worldmap.Creature, 0)

	for i := -d; i < d+1; i++ {
		for j := -d; j < d+1; j++ {
			// Translate location into world coordinates
			wX, wY := aiX+j, aiY+i
			if world.IsValid(wX, wY) && world.IsVisible(ai, wX, wY) && world.GetCreature(wX, wY) != nil && c.possibleThreats.Exists(world.GetCreature(wX, wY).GetID()) {
				visibleThreats = append(visibleThreats, world.GetCreature(wX, wY))
			}
		}
	}

	// Only consider visible threats
	c.possibleThreats = structs.Initialise()
	for _, t := range visibleThreats {
		c.possibleThreats.Add(t.GetID())
	}

	return visibleThreats
}

func (c threatsComponent) targets(ai hasAi, world *worldmap.Map) []worldmap.Creature {
	return c.threats(ai, world)
}

func (c threatsComponent) nextState(currState string, ai hasAi, world *worldmap.Map) string {
	if (currState == "fleeing" || currState == "fighting") && len(c.threats(ai, world)) == 0 {
		return "normal"
	}

	if (currState == "normal" || currState == "fighting") && len(c.threats(ai, world)) > 0 {
		return "fighting"
	}

	return currState
}

func (c threatsComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	buffer.WriteString("\"Type\": \"threats\",")

	possibleThreatsValue, err := json.Marshal(c.possibleThreats.Items())
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"PossibleThreats\":%s,", possibleThreatsValue))

	creatureIDValue, err := json.Marshal(c.creatureID)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"CreatureId\":%s", creatureIDValue))

	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (c *threatsComponent) UnmarshalJSON(data []byte) error {
	type threatsJSON struct {
		PossibleThreats []string
		CreatureId      string
	}

	var v threatsJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	c.possibleThreats = structs.Initialise()
	for _, t := range v.PossibleThreats {
		c.possibleThreats.Add(t)
	}

	c.creatureID = v.CreatureId
	return nil
}

type isWeakComponent struct {
	threshold float64
}

func (c isWeakComponent) weak(ai damageable) bool {
	curr := float64(ai.hp().Value())
	max := float64(ai.hp().Maximum())
	return curr/max <= c.threshold
}

func (c isWeakComponent) nextState(currState string, ai hasAi, world *worldmap.Map) string {
	if c.weak(ai) {
		return "fleeing"
	}

	if currState == "fleeing" && !c.weak(ai) {
		return "normal"
	}
	return currState
}

func (c isWeakComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	buffer.WriteString("\"Type\": \"isWeak\",")

	thresholdValue, err := json.Marshal(c.threshold)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Threshold\":%s", thresholdValue))

	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (c *isWeakComponent) UnmarshalJSON(data []byte) error {
	type isWeakJSON struct {
		Threshold float64
	}

	var v isWeakJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	c.threshold = v.Threshold
	return nil
}

type hasMountComponent struct{}

func (c hasMountComponent) hasMount(ai hasAi) bool {
	r, ok := ai.(Rider)
	return ok && r.Mount() != nil
}

func (c hasMountComponent) nextState(currState string, ai hasAi, world *worldmap.Map) string {
	if currState == "normal" && !c.hasMount(ai) {
		return "finding mount"
	}

	if currState == "finding mount" && c.hasMount(ai) {
		return "normal"
	}
	return currState
}

func (c hasMountComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{\"Type\": \"hasMount\"}")
	return buffer.Bytes(), nil
}

func (c *hasMountComponent) UnmarshalJSON(data []byte) error {
	return nil
}

type findMountComponent struct{}

func (c findMountComponent) findMount(ai hasAi, world *worldmap.Map) Action {
	mountMap := getMountMap(ai, world)
	if action := mount(ai, world, mountMap); action != nil {
		return action
	}

	tileUnoccupied := func(x, y int) bool {
		return !world.IsOccupied(x, y) && world.IsPassable(x, y)
	}

	locations := possibleLocationsFromAiMap(ai, world, mountMap, tileUnoccupied)
	if action := move(ai, world, locations); action != nil {
		return action
	}
	return nil
}

func (c findMountComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{}")
	return buffer.Bytes(), nil
}

func (c *findMountComponent) UnmarshalJSON(data []byte) error {
	return nil
}

type fleeComponent struct{}

func (c fleeComponent) flee(ai hasAi, world *worldmap.Map, threats []worldmap.Creature) Action {
	fleeMap := getFleeMap(ai, world, threats)
	tileUnoccupied := func(x, y int) bool {
		return !world.IsOccupied(x, y) && world.IsPassable(x, y)
	}
	locations := possibleLocationsFromAiMap(ai, world, fleeMap, tileUnoccupied)

	if action := moveIfMounted(ai, world, locations); action != nil {
		return action
	}

	if action := move(ai, world, locations); action != nil {
		return action
	}
	return nil
}

func (c fleeComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{}")
	return buffer.Bytes(), nil
}

func (c *fleeComponent) UnmarshalJSON(data []byte) error {
	return nil
}

type consumeComponent struct {
	attribute string
}

func (c consumeComponent) consume(ai hasAi) Action {
	if itemHolder, ok := ai.(holdsItems); ok {
		for _, itm := range itemHolder.Inventory() {
			if consumable, ok := itm.Component("consumable").(item.ConsumableComponent); ok && len(consumable.Effects[c.attribute]) > 0 {
				return ConsumeAction{ai, itm}
			}
		}
	}
	return nil
}

func (c consumeComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	attributeValue, err := json.Marshal(c.attribute)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Attribute\":%s", attributeValue))

	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (c *consumeComponent) UnmarshalJSON(data []byte) error {
	type consumeJSON struct {
		Attribute string
	}

	var v consumeJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	c.attribute = v.Attribute
	return nil
}

type waypointComponent struct {
	waypoint worldmap.WaypointSystem
}

func (c waypointComponent) move(ai hasAi, world *worldmap.Map) Action {
	aiX, aiY := ai.GetCoordinates()
	location := worldmap.Coordinates{aiX, aiY}
	waypoint := c.waypoint.NextWaypoint(location)
	waypointMap := getWaypointMap(ai, waypoint, world)

	tileUnoccupied := func(x, y int) bool {
		return !world.IsOccupied(x, y)
	}

	locations := possibleLocationsFromAiMap(ai, world, waypointMap, tileUnoccupied)
	if action := moveIfMounted(ai, world, locations); action != nil {
		return action
	}

	if action := move(ai, world, locations); action != nil {
		return action
	}
	return nil
}

func (c waypointComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	waypointValue, err := json.Marshal(c.waypoint)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Waypoint\":%s", waypointValue))

	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (c *waypointComponent) UnmarshalJSON(data []byte) error {
	type waypointJSON struct {
		Waypoint map[string]interface{}
	}

	var v waypointJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	c.waypoint = worldmap.UnmarshalWaypointSystem(v.Waypoint)
	return nil
}

type chaseComponent struct {
	cover float64
	chase float64
}

func (c chaseComponent) chaseTargets(ai hasAi, world *worldmap.Map, targets []worldmap.Creature) Action {
	coefficients := []float64{c.cover, c.chase}
	coverMap := getCoverMap(ai, world, targets)
	chaseMap := getChaseMap(ai, world, targets)
	aiMap := addMaps([][][]float64{coverMap, chaseMap}, coefficients)

	tileUnoccupied := func(x, y int) bool {
		return !world.IsOccupied(x, y)
	}

	locations := possibleLocationsFromAiMap(ai, world, aiMap, tileUnoccupied)

	if action := moveIfMounted(ai, world, locations); action != nil {
		return action
	}

	if action := move(ai, world, locations); action != nil {
		return action
	}
	return nil
}

func (c chaseComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	coverValue, err := json.Marshal(c.cover)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Cover\":%s,", coverValue))

	chaseValue, err := json.Marshal(c.chase)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Chase\":%s", chaseValue))

	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (c *chaseComponent) UnmarshalJSON(data []byte) error {
	type chaseJSON struct {
		Cover float64
		Chase float64
	}

	var v chaseJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	c.cover = v.Cover
	c.chase = v.Chase

	return nil
}

type coverComponent struct{}

func (c coverComponent) cover(ai hasAi, world *worldmap.Map, targets []worldmap.Creature) Action {
	coverMap := getCoverMap(ai, world, targets)
	return moveThroughCover(ai, coverMap)
}

func (c coverComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{}")
	return buffer.Bytes(), nil
}

func (c *coverComponent) UnmarshalJSON(data []byte) error {
	return nil
}

type itemsComponent struct{}

func (c itemsComponent) pickupItems(ai hasAi, world *worldmap.Map) Action {
	return pickupItems(ai, world)
}

func (c itemsComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{}")
	return buffer.Bytes(), nil
}

func (c *itemsComponent) UnmarshalJSON(data []byte) error {
	return nil
}

func unmarshalComponents(cs []map[string]interface{}) []senses {
	components := make([]senses, 0)
	for _, c := range cs {
		componentJSON, err := json.Marshal(c)
		check(err)
		var component senses
		switch c["Type"] {
		case "bounties":
			var bounties bountiesComponent
			err := json.Unmarshal(componentJSON, &bounties)
			check(err)
			component = bounties
			event.Subscribe(bounties)
		case "threats":
			var threats threatsComponent
			err := json.Unmarshal(componentJSON, &threats)
			check(err)
			component = threats
			event.Subscribe(threats)
		case "isWeak":
			var isWeak isWeakComponent
			err := json.Unmarshal(componentJSON, &isWeak)
			check(err)
			component = isWeak
		case "hasMount":
			var hasMount hasMountComponent
			err := json.Unmarshal(componentJSON, &hasMount)
			check(err)
			component = hasMount
		}
		components = append(components, component)
	}
	return components
}
