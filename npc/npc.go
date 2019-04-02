package npc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"

	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/mount"
	"github.com/onorton/cowboysindians/ui"
	"github.com/onorton/cowboysindians/worldmap"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type NpcAttributes struct {
	Icon          icon.Icon
	Initiative    int
	Hp            int
	Ac            int
	Str           int
	Dex           int
	Encumbrance   int
	Money         int
	Inventory     []item.ItemDefinition
	ShopInventory map[string]int
	DialogueType  dialogueType
}

type nameData struct {
	FirstNames []string
	LastNames  []string
}

var npcData map[string]NpcAttributes = fetchNpcData()
var names nameData = fetchNameData()

func fetchNpcData() map[string]NpcAttributes {
	data, err := ioutil.ReadFile("data/npc.json")
	check(err)
	var eD map[string]NpcAttributes
	err = json.Unmarshal(data, &eD)
	check(err)
	return eD
}

func fetchNameData() nameData {
	data, err := ioutil.ReadFile("data/names.json")
	check(err)
	var nD nameData
	err = json.Unmarshal(data, &nD)
	check(err)
	return nD
}

func generateName() string {
	firstName := names.FirstNames[rand.Intn(len(names.FirstNames))]
	lastName := names.LastNames[rand.Intn(len(names.LastNames))]

	return firstName + " " + lastName
}

func NewNpc(npcType string, x, y int, world *worldmap.Map) *Npc {
	n := npcData[npcType]
	location := worldmap.Coordinates{x, y}
	name := generateName()
	npc := &Npc{name, worldmap.Coordinates{x, y}, n.Icon, n.Initiative, n.Hp, n.Hp, n.Ac, n.Str, n.Dex, n.Encumbrance, false, n.Money, nil, nil, make([]item.Item, 0), "", nil, world, worldmap.NewRandomWaypoint(world, location), &basicDialogue{false}}
	npc.initialiseInventory(n.Inventory)
	return npc
}

func NewShopkeeper(npcType string, x, y int, world *worldmap.Map, b worldmap.Building) *Npc {
	n := npcData[npcType]
	var dialogue dialogue
	if n.DialogueType == Basic {
		dialogue = &basicDialogue{false}
	} else {
		dialogue = &shopkeeperDialogue{false, world, b}
	}
	location := worldmap.Coordinates{x, y}
	name := generateName()
	npc := &Npc{name, worldmap.Coordinates{x, y}, n.Icon, n.Initiative, n.Hp, n.Hp, n.Ac, n.Str, n.Dex, n.Encumbrance, false, n.Money, nil, nil, make([]item.Item, 0), "", nil, world, worldmap.NewWithinBuilding(world, b, location), dialogue}
	for c, count := range n.ShopInventory {

		for i := 0; i < count; i++ {
			switch c {
			case "Ammo":
				npc.PickupItem(item.GenerateAmmo())
			case "Armour":
				npc.PickupItem(item.GenerateArmour())
			case "Consumable":
				npc.PickupItem(item.GenerateConsumable())
			case "Item":
				npc.PickupItem(item.GenerateItem())
			case "Weapon":
				npc.PickupItem(item.GenerateWeapon())
			}
		}
	}

	npc.initialiseInventory(n.Inventory)

	return npc
}

func (npc *Npc) initialiseInventory(inventory []item.ItemDefinition) {
	for _, itemDefinition := range inventory {
		for i := 0; i < itemDefinition.Amount; i++ {
			var itm item.Item = nil
			switch itemDefinition.Category {
			case "Ammo":
				itm = item.NewAmmo(itemDefinition.Name)
			case "Armour":
				itm = item.NewArmour(itemDefinition.Name)
			case "Consumable":
				itm = item.NewConsumable(itemDefinition.Name)
			case "Item":
				itm = item.NewNormalItem(itemDefinition.Name)
			case "Weapon":
				itm = item.NewWeapon(itemDefinition.Name)
			}
			npc.PickupItem(itm)
		}
	}
}

func (npc *Npc) Render() ui.Element {
	if npc.mount != nil {
		return icon.MergeIcons(npc.icon, npc.mount.GetIcon())
	}
	return npc.icon.Render()
}

func (npc *Npc) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	keys := []string{"Name", "Location", "Icon", "Initiative", "Hp", "MaxHp", "AC", "Str", "Dex", "Encumbrance", "Crouching", "Money", "Weapon", "Armour", "Inventory", "MountID", "WaypointSystem", "Dialogue"}

	mountID := ""
	if npc.mount != nil {
		mountID = npc.mount.GetID()
	}

	npcValues := map[string]interface{}{
		"Name":           npc.name,
		"Location":       npc.location,
		"Icon":           npc.icon,
		"Initiative":     npc.initiative,
		"Hp":             npc.hp,
		"MaxHp":          npc.maxHp,
		"AC":             npc.ac,
		"Str":            npc.str,
		"Dex":            npc.dex,
		"Encumbrance":    npc.encumbrance,
		"Crouching":      npc.crouching,
		"Money":          npc.money,
		"Weapon":         npc.weapon,
		"Armour":         npc.armour,
		"Inventory":      npc.inventory,
		"MountID":        mountID,
		"WaypointSystem": npc.waypoint,
		"Dialogue":       npc.dialogue,
	}

	length := len(npcValues)
	count := 0

	for _, key := range keys {
		jsonValue, err := json.Marshal(npcValues[key])

		if err != nil {
			return nil, err
		}

		buffer.WriteString(fmt.Sprintf("\"%s\":%s", key, jsonValue))
		count++
		if count < length {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (npc *Npc) Talk() bool {
	return npc.dialogue.interact()
}

func (npc *Npc) UnmarshalJSON(data []byte) error {

	type npcJson struct {
		Name           string
		Location       worldmap.Coordinates
		Icon           icon.Icon
		Initiative     int
		Hp             int
		MaxHp          int
		AC             int
		Str            int
		Dex            int
		Encumbrance    int
		Crouching      bool
		Money          int
		Weapon         *item.Weapon
		Armour         *item.Armour
		Inventory      item.ItemList
		MountID        string
		WaypointSystem map[string]interface{}
		Dialogue       map[string]interface{}
	}
	var v npcJson

	json.Unmarshal(data, &v)

	npc.name = v.Name
	npc.location = v.Location
	npc.icon = v.Icon
	npc.initiative = v.Initiative
	npc.hp = v.Hp
	npc.maxHp = v.MaxHp
	npc.ac = v.AC
	npc.str = v.Str
	npc.dex = v.Dex
	npc.encumbrance = v.Encumbrance
	npc.crouching = v.Crouching
	npc.money = v.Money
	npc.weapon = v.Weapon
	npc.armour = v.Armour
	npc.inventory = v.Inventory
	npc.mountID = v.MountID
	npc.waypoint = worldmap.UnmarshalWaypointSystem(v.WaypointSystem)
	npc.dialogue = unmarshalDialogue(v.Dialogue)

	return nil
}

func (npc *Npc) GetCoordinates() (int, int) {
	return npc.location.X, npc.location.Y
}
func (npc *Npc) SetCoordinates(x int, y int) {
	npc.location = worldmap.Coordinates{x, y}
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

func (npc *Npc) generateMap(aiMap [][]int) [][]int {
	width, height := len(aiMap[0]), len(aiMap)
	prev := make([][]int, height)
	for i, _ := range prev {
		prev[i] = make([]int, width)
	}
	// While map changes, update
	for !compareMaps(aiMap, prev) {
		prev = copyMap(aiMap)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				wX, wY := npc.location.X+x-npc.GetVisionDistance(), npc.location.Y+y-npc.GetVisionDistance()
				if !npc.world.IsValid(wX, wY) || !npc.world.IsPassable(wX, wY) {
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
	return npc.generateMap(aiMap)
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
				if m, ok := npc.world.GetCreature(wX, wY).(*mount.Mount); ok && m != nil {
					aiMap[y][x] = 0
				} else {
					aiMap[y][x] = width * width
				}
			}
		}
	}
	return npc.generateMap(aiMap)
}

func (npc *Npc) GetInitiative() int {
	return npc.initiative
}

func (npc *Npc) MeleeAttack(c worldmap.Creature) {
	npc.attack(c, worldmap.GetBonus(npc.str), worldmap.GetBonus(npc.str))
}
func (npc *Npc) attack(c worldmap.Creature, hitBonus, damageBonus int) {

	hits := c.AttackHits(rand.Intn(20) + hitBonus + 1)
	if hits {
		if npc.weapon != nil {
			c.TakeDamage(npc.weapon.GetDamage() + damageBonus)
		} else {
			c.TakeDamage(damageBonus)
		}
	}
	if c.GetAlignment() == worldmap.Player {
		if hits {
			message.Enqueue(fmt.Sprintf("The %s hit you.", npc.name))
		} else {
			message.Enqueue(fmt.Sprintf("The %s missed you.", npc.name))
		}
	}

}

func (npc *Npc) AttackHits(roll int) bool {
	return roll > npc.ac
}
func (npc *Npc) TakeDamage(damage int) {
	npc.hp -= damage
}

func (npc *Npc) IsDead() bool {
	return npc.hp <= 0
}

func (npc *Npc) wieldItem() bool {
	changed := false
	for i, itm := range npc.inventory {
		if w, ok := itm.(*item.Weapon); ok {
			if npc.weapon == nil {
				npc.weapon = w
				npc.inventory = append(npc.inventory[:i], npc.inventory[i+1:]...)
				changed = true

			} else if w.GetMaxDamage() > npc.weapon.GetMaxDamage() {
				npc.inventory[i] = npc.weapon
				npc.weapon = w
				changed = true
			}

		}

	}
	return changed
}

func (npc *Npc) wearArmour() bool {
	changed := false
	for i, itm := range npc.inventory {
		if a, ok := itm.(*item.Armour); ok {
			if npc.armour == nil {
				npc.armour = a
				npc.inventory = append(npc.inventory[:i], npc.inventory[i+1:]...)
				changed = true

			} else if a.GetACBonus() > npc.armour.GetACBonus() {
				npc.inventory[i] = npc.weapon
				npc.armour = a
				changed = true
			}

		}

	}
	return changed
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

func (npc *Npc) FindAction() {
	waypoint := npc.waypoint.NextWaypoint(npc.location)
	aiMap := npc.getWaypointMap(waypoint)
	mountMap := npc.getMountMap()

	current := aiMap[npc.GetVisionDistance()][npc.GetVisionDistance()]
	possibleLocations := make([]worldmap.Coordinates, 0)

	// Find adjacent locations closer to the goal
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			nX := npc.location.X + i
			nY := npc.location.Y + j
			if aiMap[nY-npc.location.Y+npc.GetVisionDistance()][nX-npc.location.X+npc.GetVisionDistance()] < current {
				// Add if not occupied
				if npc.world.IsValid(nX, nY) && !npc.world.IsOccupied(nX, nY) {
					possibleLocations = append(possibleLocations, worldmap.Coordinates{nX, nY})
				}
			}
		}
	}
	// If mounted, can move first before executing another action
	if npc.mount != nil && !npc.mount.Moved() {
		if len(possibleLocations) > 0 {
			if npc.overEncumbered() {
				for _, itm := range npc.inventory {
					if itm.GetWeight() > 1 {
						npc.dropItem(itm)
						return
					}
				}
			} else {
				l := possibleLocations[rand.Intn(len(possibleLocations))]
				npc.mount.Move()
				npc.location = l
				// Can choose new action again
				npc.FindAction()
				return
			}
		}
	}

	// If at half health heal up
	if npc.hp <= npc.maxHp/2 {
		for i, itm := range npc.inventory {
			if con, ok := itm.(*item.Consumable); ok && con.GetEffect("hp") > 0 {
				npc.heal(con.GetEffect("hp"))
				npc.inventory = append(npc.inventory[:i], npc.inventory[i+1:]...)
				return
			}
		}
	}

	// If adjacent to mount, attempt to mount it
	if npc.mount == nil {
		for i := -1; i <= 1; i++ {
			for j := -1; j <= 1; j++ {
				x, y := npc.location.X+j, npc.location.Y+i
				if npc.world.IsValid(x, y) && mountMap[npc.GetVisionDistance()+i][npc.GetVisionDistance()+j] == 0 {
					m, _ := npc.world.GetCreature(x, y).(*mount.Mount)
					m.AddRider(npc)
					npc.world.DeleteCreature(m)
					npc.mount = m
					npc.crouching = false
					npc.location = worldmap.Coordinates{x, y}
					return
				}
			}
		}
	}

	// If adjacent to closed door attempt to open it
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			x, y := npc.location.X+j, npc.location.Y+i
			if npc.world.IsValid(x, y) && npc.world.IsDoor(x, y) && !npc.world.IsPassable(x, y) {
				npc.world.ToggleDoor(x, y, true)
				return
			}
		}
	}

	if len(possibleLocations) > 0 {
		if npc.overEncumbered() {
			for _, itm := range npc.inventory {
				if itm.GetWeight() > 1 {
					npc.dropItem(itm)
					return
				}
			}
		} else if npc.mount == nil || (npc.mount != nil && !npc.mount.Moved()) {
			l := possibleLocations[rand.Intn(len(possibleLocations))]
			npc.location = l
			return
		}
	} else {
		items := npc.world.GetItems(npc.location.X, npc.location.Y)
		for _, item := range items {
			npc.PickupItem(item)
		}
	}
	return
}

func (npc *Npc) overEncumbered() bool {
	weight := 0.0
	for _, item := range npc.inventory {
		weight += item.GetWeight()
	}
	return weight > float64(npc.encumbrance)
}
func (npc *Npc) dropItem(item item.Item) {
	npc.world.PlaceItem(npc.location.X, npc.location.Y, item)
	if npc.world.IsVisible(npc.world.GetPlayer(), npc.location.X, npc.location.Y) {
		message.Enqueue(fmt.Sprintf("The %s dropped a %s.", npc.name, item.GetName()))
	}

}

func (npc *Npc) Update() (int, int) {
	// Needs to be fixed
	x, y := npc.location.X, npc.location.Y
	p := npc.world.GetPlayer()
	pX, pY := p.GetCoordinates()
	if npc.world.InConversationRange(npc, p) {
		npc.dialogue.initialGreeting()
	} else if npc.world.IsVisible(npc, pX, pY) {
		npc.dialogue.resetSeen()
	}
	npc.FindAction()
	if npc.mount != nil {
		npc.mount.ResetMoved()
		if npc.mount.IsDead() {
			npc.mount = nil
		}
	}
	newX, newY := npc.location.X, npc.location.Y
	npc.location = worldmap.Coordinates{x, y}
	return newX, newY
}

func (npc *Npc) EmptyInventory() {
	itemTypes := make(map[string]int)
	for _, item := range npc.inventory {
		npc.world.PlaceItem(npc.location.X, npc.location.Y, item)
		itemTypes[item.GetName()]++
	}

	if npc.weapon != nil {
		npc.world.PlaceItem(npc.location.X, npc.location.Y, npc.weapon)
		itemTypes[npc.weapon.GetName()]++
		npc.weapon = nil
	}
	if npc.armour != nil {
		npc.world.PlaceItem(npc.location.X, npc.location.Y, npc.armour)
		itemTypes[npc.armour.GetName()]++
		npc.armour = nil
	}

	if npc.money > 0 {
		npc.world.PlaceItem(npc.location.X, npc.location.Y, item.Money(npc.money))
		itemTypes["money"]++
	}

	if npc.world.IsVisible(npc.world.GetPlayer(), npc.location.X, npc.location.Y) {
		for name, count := range itemTypes {
			if count == 1 {
				message.Enqueue(fmt.Sprintf("The %s dropped 1 %s.", npc.name, name))
			} else {
				message.Enqueue(fmt.Sprintf("The %s dropped %d %ss.", npc.name, count, name))
			}
		}
	}

}

func (npc *Npc) getAmmo() *item.Ammo {
	for i, itm := range npc.inventory {
		if ammo, ok := itm.(*item.Ammo); ok && npc.weapon.AmmoTypeMatches(ammo) {
			npc.inventory = append(npc.inventory[:i], npc.inventory[i+1:]...)
			return ammo
		}
	}
	return nil
}

func (npc *Npc) PickupItem(item item.Item) {
	npc.inventory = append(npc.inventory, item)
}

func (npc *Npc) ranged() bool {
	if npc.weapon != nil {
		return npc.weapon.GetRange() > 0
	}
	return false
}

// Check whether npc is carrying a fully loaded weapon
func (npc *Npc) weaponFullyLoaded() bool {
	return npc.weapon.IsFullyLoaded()
}

// Check whether npc has ammo for particular wielded weapon
func (npc *Npc) hasAmmo() bool {
	for _, itm := range npc.inventory {
		if a, ok := itm.(*item.Ammo); ok && npc.weapon.AmmoTypeMatches(a) {
			return true
		}
	}
	return false
}

func (npc *Npc) weaponLoaded() bool {
	if npc.weapon != nil && npc.weapon.NeedsAmmo() {
		return !npc.weapon.IsUnloaded()
	}
	return true

}

func (npc *Npc) heal(amount int) {
	originalHp := npc.hp
	npc.hp = int(math.Min(float64(originalHp+amount), float64(npc.maxHp)))
}

func (npc *Npc) GetName() string {
	return npc.name
}

func (npc *Npc) GetAlignment() worldmap.Alignment {
	return worldmap.Neutral
}

func (npc *Npc) IsCrouching() bool {
	return npc.crouching
}

func (npc *Npc) SetMap(world *worldmap.Map) {
	npc.world = world

	switch w := npc.waypoint.(type) {
	case *worldmap.RandomWaypoint:
		w.SetMap(world)
	case *worldmap.Patrol:
	case *worldmap.WithinBuilding:
		w.SetMap(world)
	}

	if d, ok := npc.dialogue.(*shopkeeperDialogue); ok {
		d.setMap(world)
	}

}

func (npc *Npc) GetMount() worldmap.Creature {
	return npc.mount
}

func (npc *Npc) GetVisionDistance() int {
	return 20
}

func (npc *Npc) GetItems() map[rune]([]item.Item) {
	items := make(map[rune]([]item.Item))
	for _, itm := range npc.inventory {
		existing := items[itm.GetKey()]
		if existing == nil {
			existing = make([]item.Item, 0)
		}
		existing = append(existing, itm)
		items[itm.GetKey()] = existing
	}
	return items
}

func (npc *Npc) RemoveItem(itm item.Item) {
	for i, item := range npc.inventory {
		if itm.GetName() == item.GetName() {
			npc.inventory = append(npc.inventory[:i], npc.inventory[i+1:]...)
			return
		}
	}
}

func (npc *Npc) LoadMount(mounts []*mount.Mount) {
	for _, m := range mounts {
		if npc.mountID == m.GetID() {
			m.AddRider(npc)
			npc.mount = m
		}
	}
}

func (npc Npc) CanBuy(itm item.Item) bool {
	return itm.GetValue() <= npc.money
}

func (npc *Npc) AddMoney(amount int) {
	npc.money += amount
}

func (npc *Npc) RemoveMoney(amount int) {
	npc.money += amount
}

type Npc struct {
	name        string
	location    worldmap.Coordinates
	icon        icon.Icon
	initiative  int
	hp          int
	maxHp       int
	ac          int
	str         int
	dex         int
	encumbrance int
	crouching   bool
	money       int
	weapon      *item.Weapon
	armour      *item.Armour
	inventory   []item.Item
	mountID     string
	mount       *mount.Mount
	world       *worldmap.Map
	waypoint    worldmap.WaypointSystem
	dialogue    dialogue
}
