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

type hasAction interface {
	action(hasAi, *worldmap.Map) Action
	shouldHappen(state string) float64
}

type hasTargets interface {
	addTargets([]worldmap.Creature)
}

type hasThreats interface {
	addThreats([]worldmap.Creature)
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

func (c findMountComponent) action(ai hasAi, world *worldmap.Map) Action {
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

func (c findMountComponent) shouldHappen(state string) float64 {
	if state == "finding mount" {
		return 1
	}
	return 0
}

func (c findMountComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{\"Type\": \"findMount\"}")
	return buffer.Bytes(), nil
}

func (c *findMountComponent) UnmarshalJSON(data []byte) error {
	return nil
}

type fleeComponent struct {
	threats []worldmap.Creature
}

func (c fleeComponent) action(ai hasAi, world *worldmap.Map) Action {
	fleeMap := getFleeMap(ai, world, c.threats)
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

func (c fleeComponent) shouldHappen(state string) float64 {
	// More likely to flee as more threats
	if state == "fleeing" {
		threshold := 3.0
		return (1.0 / threshold) * float64(len(c.threats))
	}
	return 0
}

func (c *fleeComponent) addThreats(threats []worldmap.Creature) {
	c.threats = threats
}

func (c fleeComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{\"Type\": \"flee\"}")
	return buffer.Bytes(), nil
}

func (c *fleeComponent) UnmarshalJSON(data []byte) error {
	return nil
}

type consumeComponent struct {
	attribute string
}

func (c consumeComponent) action(ai hasAi, world *worldmap.Map) Action {
	if itemHolder, ok := ai.(holdsItems); ok {
		for _, itm := range itemHolder.Inventory() {
			if consumable, ok := itm.Component("consumable").(item.ConsumableComponent); ok && len(consumable.Effects[c.attribute]) > 0 {
				return ConsumeAction{ai, itm}
			}
		}
	}
	return nil
}

func (c consumeComponent) shouldHappen(state string) float64 {
	if state == "fleeing" {
		return 1
	}
	return 0
}

func (c consumeComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	buffer.WriteString("\"Type\": \"consume\",")
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

func (c waypointComponent) action(ai hasAi, world *worldmap.Map) Action {
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

func (c waypointComponent) shouldHappen(state string) float64 {
	if state == "normal" {
		return 0.5
	}
	return 0
}

func (c waypointComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	buffer.WriteString("\"Type\": \"waypoint\",")

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
	cover   float64
	chase   float64
	targets []worldmap.Creature
}

func (c chaseComponent) action(ai hasAi, world *worldmap.Map) Action {
	coefficients := []float64{c.cover, c.chase}
	coverMap := getCoverMap(ai, world, c.targets)
	chaseMap := getChaseMap(ai, world, c.targets)
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

func (c chaseComponent) shouldHappen(state string) float64 {
	if state == "fighting" {
		return 0.5
	}
	return 0
}

func (c *chaseComponent) addTargets(targets []worldmap.Creature) {
	c.targets = targets
}

func (c chaseComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	buffer.WriteString("\"Type\": \"chase\",")

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

type coverComponent struct {
	targets []worldmap.Creature
}

func (c coverComponent) action(ai hasAi, world *worldmap.Map) Action {
	coverMap := getCoverMap(ai, world, c.targets)
	return moveThroughCover(ai, coverMap)
}

func (c coverComponent) shouldHappen(state string) float64 {
	if state == "fighting" {
		return 1
	}
	return 0
}

func (c *coverComponent) addTargets(targets []worldmap.Creature) {
	c.targets = targets
}

func (c coverComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{\"Type\": \"cover\"}")
	return buffer.Bytes(), nil
}

func (c *coverComponent) UnmarshalJSON(data []byte) error {
	return nil
}

type itemsComponent struct{}

func (c itemsComponent) action(ai hasAi, world *worldmap.Map) Action {
	return pickupItems(ai, world)
}

func (c itemsComponent) shouldHappen(state string) float64 {
	if state == "normal" {
		return 0.25
	}
	return 0
}

func (c itemsComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{\"Type\": \"items\"}")
	return buffer.Bytes(), nil
}

func (c *itemsComponent) UnmarshalJSON(data []byte) error {
	return nil
}

type doorComponent struct{}

func (c doorComponent) action(ai hasAi, world *worldmap.Map) Action {
	return tryOpeningDoor(ai, world)
}

func (c doorComponent) shouldHappen(state string) float64 {
	return 1
}

func (c doorComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{\"Type\": \"door\"}")
	return buffer.Bytes(), nil
}

func (c *doorComponent) UnmarshalJSON(data []byte) error {
	return nil
}

type rangedComponent struct {
	targets []worldmap.Creature
}

func (c rangedComponent) action(ai hasAi, world *worldmap.Map) Action {
	return rangedAttack(ai, world, c.targets)
}

func (c rangedComponent) shouldHappen(state string) float64 {
	if state == "fighting" || state == "fleeing" {
		return 0.75
	}
	return 0
}

func (c *rangedComponent) addTargets(targets []worldmap.Creature) {
	c.targets = targets
}

func (c rangedComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{\"Type\": \"ranged\"}")
	return buffer.Bytes(), nil
}

func (c *rangedComponent) UnmarshalJSON(data []byte) error {
	return nil
}

type wieldComponent struct{}

func (c wieldComponent) action(ai hasAi, world *worldmap.Map) Action {
	if itemUser, ok := ai.(usesItems); ok && itemUser.wieldItem() {
		return NoAction{}
	}
	return nil
}

func (c wieldComponent) shouldHappen(state string) float64 {
	if state == "normal" {
		return 0.75
	}
	return 0
}

func (c wieldComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{\"Type\": \"wield\"}")
	return buffer.Bytes(), nil
}

func (c *wieldComponent) UnmarshalJSON(data []byte) error {
	return nil
}

type wearComponent struct{}

func (c wearComponent) action(ai hasAi, world *worldmap.Map) Action {
	if itemUser, ok := ai.(usesItems); ok && itemUser.wearArmour() {
		return NoAction{}
	}
	return nil
}

func (c wearComponent) shouldHappen(state string) float64 {
	if state == "normal" {
		return 0.75
	}
	return 0
}

func (c wearComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{\"Type\": \"wear\"}")
	return buffer.Bytes(), nil
}

func (c *wearComponent) UnmarshalJSON(data []byte) error {
	return nil
}

func unmarshalSenses(cs []map[string]interface{}) []senses {
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

func unmarshalActions(cs []map[string]interface{}) []hasAction {
	components := make([]hasAction, 0)
	for _, c := range cs {
		componentJSON, err := json.Marshal(c)
		check(err)
		var component hasAction
		switch c["Type"] {
		case "findMount":
			var findMount findMountComponent
			err := json.Unmarshal(componentJSON, &findMount)
			check(err)
			component = findMount
		case "flee":
			var flee fleeComponent
			err := json.Unmarshal(componentJSON, &flee)
			check(err)
			component = flee
		case "consume":
			var consume consumeComponent
			err := json.Unmarshal(componentJSON, &consume)
			check(err)
			component = consume
		case "waypoint":
			var waypoint waypointComponent
			err := json.Unmarshal(componentJSON, &waypoint)
			check(err)
			component = waypoint
		case "chase":
			var chase chaseComponent
			err := json.Unmarshal(componentJSON, &chase)
			check(err)
			component = chase
		case "cover":
			var cover coverComponent
			err := json.Unmarshal(componentJSON, &cover)
			check(err)
			component = cover
		case "items":
			var items itemsComponent
			err := json.Unmarshal(componentJSON, &items)
			check(err)
			component = items
		case "door":
			var door doorComponent
			err := json.Unmarshal(componentJSON, &door)
			check(err)
			component = door
		case "ranged":
			var ranged rangedComponent
			err := json.Unmarshal(componentJSON, &ranged)
			check(err)
			component = ranged
		case "wield":
			var wield wieldComponent
			err := json.Unmarshal(componentJSON, &wield)
			check(err)
			component = wield
		case "wear":
			var wear wearComponent
			err := json.Unmarshal(componentJSON, &wear)
			check(err)
			component = wear
		}
		components = append(components, component)
	}
	return components
}
