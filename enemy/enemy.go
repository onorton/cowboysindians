package enemy

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
	"github.com/onorton/cowboysindians/ui"
	"github.com/onorton/cowboysindians/worldmap"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type EnemyAttributes struct {
	Icon        icon.Icon
	Initiative  int
	Hp          int
	Ac          int
	Str         int
	Dex         int
	Encumbrance int
	Inventory   []item.ItemDefinition
}

var enemyData map[string]EnemyAttributes = fetchEnemyData()

func fetchEnemyData() map[string]EnemyAttributes {
	data, err := ioutil.ReadFile("data/enemy.json")
	check(err)
	var eD map[string]EnemyAttributes
	err = json.Unmarshal(data, &eD)
	check(err)
	return eD
}

func NewEnemy(name string, x, y int, world *worldmap.Map) *Enemy {
	enemy := enemyData[name]
	e := &Enemy{name, x, y, enemy.Icon, enemy.Initiative, enemy.Hp, enemy.Hp, enemy.Ac, enemy.Str, enemy.Dex, enemy.Encumbrance, false, nil, nil, make([]item.Item, 0), world}
	for _, itemDefinition := range enemy.Inventory {
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
			e.pickupItem(itm)
		}
	}
	return e
}
func (e *Enemy) Render() ui.Element {
	return e.icon.Render()
}

func (e *Enemy) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	keys := []string{"Name", "X", "Y", "Icon", "Initiative", "Hp", "MaxHp", "AC", "Str", "Dex", "Encumbrance", "Crouching", "Weapon", "Armour", "Inventory"}

	enemyValues := map[string]interface{}{
		"Name":        e.name,
		"X":           e.x,
		"Y":           e.y,
		"Icon":        e.icon,
		"Initiative":  e.initiative,
		"Hp":          e.hp,
		"MaxHp":       e.maxHp,
		"AC":          e.ac,
		"Str":         e.str,
		"Dex":         e.dex,
		"Encumbrance": e.encumbrance,
		"Crouching":   e.crouching,
		"Weapon":      e.weapon,
		"Armour":      e.armour,
		"Inventory":   e.inventory,
	}

	length := len(enemyValues)
	count := 0

	for _, key := range keys {
		jsonValue, err := json.Marshal(enemyValues[key])
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

func (e *Enemy) UnmarshalJSON(data []byte) error {

	type enemyJson struct {
		Name        string
		X           int
		Y           int
		Icon        icon.Icon
		Initiative  int
		Hp          int
		MaxHp       int
		AC          int
		Str         int
		Dex         int
		Encumbrance int
		Crouching   bool
		Weapon      *item.Weapon
		Armour      *item.Armour
		Inventory   item.ItemList
	}
	var v enemyJson

	json.Unmarshal(data, &v)

	e.name = v.Name
	e.x = v.X
	e.y = v.Y
	e.icon = v.Icon
	e.initiative = v.Initiative
	e.hp = v.Hp
	e.maxHp = v.MaxHp
	e.ac = v.AC
	e.str = v.Str
	e.dex = v.Dex
	e.encumbrance = v.Encumbrance
	e.crouching = v.Crouching
	e.weapon = v.Weapon
	e.armour = v.Armour
	e.inventory = v.Inventory

	return nil
}

func (e *Enemy) GetCoordinates() (int, int) {
	return e.x, e.y
}
func (e *Enemy) SetCoordinates(x int, y int) {

	e.x = x
	e.y = y
}

func generateMap(aiMap [][]int, m *worldmap.Map) [][]int {
	width, height := len(aiMap[0]), len(aiMap)
	prev := make([][]int, height)
	for i, _ := range prev {
		prev[i] = make([]int, width)
	}
	// While map changes, update
	for !compareMaps(aiMap, prev) {
		prev = aiMap
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				if !m.IsPassable(x, y) {
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

func (e *Enemy) getChaseMap() [][]int {
	height, width := e.world.GetHeight(), e.world.GetWidth()
	aiMap := make([][]int, height)

	// Initialise Dijkstra map with goals.
	// Max is size of grid.
	for y := 0; y < height; y++ {
		aiMap[y] = make([]int, width)
		for x := 0; x < width; x++ {
			if e.world.IsVisible(e, x, y) && e.world.HasPlayer(x, y) {
				aiMap[y][x] = 0
			} else {
				aiMap[y][x] = height * width
			}
		}
	}
	return generateMap(aiMap, e.world)

}

func (e *Enemy) getItemMap() [][]int {
	height, width := e.world.GetHeight(), e.world.GetWidth()
	aiMap := make([][]int, height)

	// Initialise Dijkstra map with goals.
	// Max is size of grid.
	for y := 0; y < height; y++ {
		aiMap[y] = make([]int, width)
		for x := 0; x < width; x++ {
			if e.world.IsVisible(e, x, y) && e.world.HasItems(x, y) {
				aiMap[y][x] = 0
			} else {
				aiMap[y][x] = height * width
			}
		}
	}
	return generateMap(aiMap, e.world)
}

func (e *Enemy) getCoverMap() [][]int {
	height, width := e.world.GetHeight(), e.world.GetWidth()
	aiMap := make([][]int, height)

	player := e.world.GetPlayer()
	pX, pY := player.GetCoordinates()
	// Initialise Dijkstra map with goals.
	// Max is size of grid.
	for y := 0; y < height; y++ {
		aiMap[y] = make([]int, width)
		for x := 0; x < width; x++ {
			// Enemy must be able to see player in order to know it would be behind cover
			if e.world.IsVisible(e, x, y) && e.world.IsVisible(e, pX, pY) && e.world.BehindCover(x, y, player) {
				aiMap[y][x] = 0
			} else {
				aiMap[y][x] = height * width
			}
		}
	}
	return generateMap(aiMap, e.world)
}

func (e *Enemy) GetInitiative() int {
	return e.initiative
}

func (e *Enemy) MeleeAttack(c worldmap.Creature) {
	e.attack(c, worldmap.GetBonus(e.str), worldmap.GetBonus(e.str))
}
func (e *Enemy) attack(c worldmap.Creature, hitBonus, damageBonus int) {

	hits := c.AttackHits(rand.Intn(20) + hitBonus + 1)
	if hits {
		if e.weapon != nil {
			c.TakeDamage(e.weapon.GetDamage() + damageBonus)
		} else {
			c.TakeDamage(damageBonus)
		}
	}
	if c.GetAlignment() == worldmap.Player {
		if hits {
			message.Enqueue(fmt.Sprintf("The %s hit you.", e.name))
		} else {
			message.Enqueue(fmt.Sprintf("The %s missed you.", e.name))
		}
	}

}

func (e *Enemy) AttackHits(roll int) bool {
	return roll > e.ac
}
func (e *Enemy) TakeDamage(damage int) {
	e.hp -= damage
}

func (e *Enemy) IsDead() bool {
	return e.hp <= 0
}

func (e *Enemy) wieldItem() bool {
	changed := false
	for i, itm := range e.inventory {
		if w, ok := itm.(*item.Weapon); ok {
			if e.weapon == nil {
				e.weapon = w
				e.inventory = append(e.inventory[:i], e.inventory[i+1:]...)
				changed = true

			} else if w.GetMaxDamage() > e.weapon.GetMaxDamage() {
				e.inventory[i] = e.weapon
				e.weapon = w
				changed = true
			}

		}

	}
	return changed
}

func (e *Enemy) wearArmour() bool {
	changed := false
	for i, itm := range e.inventory {
		if a, ok := itm.(*item.Armour); ok {
			if e.armour == nil {
				e.armour = a
				e.inventory = append(e.inventory[:i], e.inventory[i+1:]...)
				changed = true

			} else if a.GetACBonus() > e.armour.GetACBonus() {
				e.inventory[i] = e.weapon
				e.armour = a
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

type Coordinate struct {
	x int
	y int
}

func (e *Enemy) Update() (int, int) {
	// If at half health heal up
	if e.hp <= e.maxHp/2 {
		for i, itm := range e.inventory {
			if con, ok := itm.(*item.Consumable); ok && con.GetEffect("hp") > 0 {
				e.heal(con.GetEffect("hp"))
				e.inventory = append(e.inventory[:i], e.inventory[i+1:]...)
				return e.x, e.y
			}
		}
	}

	coverMap := e.getCoverMap()
	aiMap := addMaps([][][]int{e.getChaseMap(), e.getItemMap(), coverMap}, []float64{0.5, 0.2, 0.3})

	// If moving into or out of cover toggle crouch
	if coverMap[e.y][e.x] == 0 && !e.crouching {
		e.crouching = true
		return e.x, e.y
	} else if coverMap[e.y][e.x] > 0 && e.crouching {
		e.crouching = false
		return e.x, e.y
	}

	// Try and wield best weapon
	if e.wieldItem() {
		return e.x, e.y
	}
	// Try and wear best armour
	if e.wearArmour() {
		return e.x, e.y
	}

	current := aiMap[e.y][e.x]
	possibleLocations := make([]Coordinate, 0)
	// Find adjacent locations closer to the goal
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			nX := e.x + i
			nY := e.y + j
			if nX >= 0 && nX < len(aiMap[0]) && nY >= 0 && nY < len(aiMap) && aiMap[nY][nX] < current {
				// Add if not occupied by another enemy
				if e.world.HasPlayer(nX, nY) || !e.world.IsOccupied(nX, nY) {
					possibleLocations = append(possibleLocations, Coordinate{nX, nY})
				}
			}
		}
	}

	target := e.world.GetPlayer()
	tX, tY := target.GetCoordinates()

	if distance := math.Sqrt(math.Pow(float64(e.x-tX), 2) + math.Pow(float64(e.y-tY), 2)); e.ranged() && distance < float64(e.weapon.GetRange()) && e.world.IsVisible(e, tX, tY) {
		// if weapon loaded, shoot at target else if enemy has ammo, load weapon
		if e.weaponLoaded() {
			e.weapon.Fire()
			coverPenalty := 0
			if e.world.TargetBehindCover(e, target) {
				coverPenalty = 5
			}
			e.attack(target, worldmap.GetBonus(e.dex)-coverPenalty, 0)
			return e.x, e.y
		} else if e.hasAmmo() {
			for !e.weaponFullyLoaded() && e.hasAmmo() {
				e.getAmmo()
				e.weapon.Load()
			}
			return e.x, e.y
		}

	}

	if len(possibleLocations) > 0 {
		if e.overEncumbered() {
			for _, itm := range e.inventory {
				if itm.GetWeight() > 1 {
					e.dropItem(itm)
				}
			}
		} else {
			l := possibleLocations[rand.Intn(len(possibleLocations))]
			return l.x, l.y
		}
	} else {
		items := e.world.GetItems(e.x, e.y)
		for _, item := range items {
			e.pickupItem(item)
		}
	}

	return e.x, e.y

}

func (e *Enemy) overEncumbered() bool {
	weight := 0.0
	for _, item := range e.inventory {
		weight += item.GetWeight()
	}
	return weight > float64(e.encumbrance)
}
func (e *Enemy) dropItem(item item.Item) {
	e.world.PlaceItem(e.x, e.y, item)
	if e.world.IsVisible(e.world.GetPlayer(), e.x, e.y) {
		message.Enqueue(fmt.Sprintf("The %s dropped a %s.", e.name, item.GetName()))
	}

}

func (e *Enemy) EmptyInventory() {
	itemTypes := make(map[string]int)
	for _, item := range e.inventory {
		e.world.PlaceItem(e.x, e.y, item)
		itemTypes[item.GetName()]++
	}

	if e.weapon != nil {
		e.world.PlaceItem(e.x, e.y, e.weapon)
		itemTypes[e.weapon.GetName()]++
		e.weapon = nil
	}
	if e.armour != nil {
		e.world.PlaceItem(e.x, e.y, e.armour)
		itemTypes[e.armour.GetName()]++
		e.armour = nil
	}

	if e.world.IsVisible(e.world.GetPlayer(), e.x, e.y) {
		for name, count := range itemTypes {
			message.Enqueue(fmt.Sprintf("The %s dropped %d %ss.", e.name, count, name))
		}
	}

}

func (e *Enemy) getAmmo() *item.Ammo {
	for i, itm := range e.inventory {
		if ammo, ok := itm.(*item.Ammo); ok && e.weapon.AmmoTypeMatches(ammo) {
			e.inventory = append(e.inventory[:i], e.inventory[i+1:]...)
			return ammo
		}
	}
	return nil
}

func (e *Enemy) pickupItem(item item.Item) {
	e.inventory = append(e.inventory, item)
}

func (e *Enemy) ranged() bool {
	if e.weapon != nil {
		return e.weapon.GetRange() > 0
	}
	return false
}

// Check whether enemy is carrying a fully loaded weapon
func (e *Enemy) weaponFullyLoaded() bool {
	return e.weapon.IsFullyLoaded()
}

// Check whether enemy has ammo for particular wielded weapon
func (e *Enemy) hasAmmo() bool {
	for _, itm := range e.inventory {
		if a, ok := itm.(*item.Ammo); ok && e.weapon.AmmoTypeMatches(a) {
			return true
		}
	}
	return false
}

func (e *Enemy) weaponLoaded() bool {
	if e.weapon != nil && e.weapon.NeedsAmmo() {
		return !e.weapon.IsUnloaded()
	}
	return true

}

func (e *Enemy) heal(amount int) {
	originalHp := e.hp
	e.hp = int(math.Min(float64(originalHp+amount), float64(e.maxHp)))
}

func (e *Enemy) GetName() string {
	return e.name
}

func (e *Enemy) GetAlignment() worldmap.Alignment {
	return worldmap.Enemy
}

func (e *Enemy) IsCrouching() bool {
	return e.crouching
}

func (e *Enemy) SetMap(world *worldmap.Map) {
	e.world = world
}

func (e *Enemy) GetMount() worldmap.Creature {
	return nil
}

type Enemy struct {
	name        string
	x           int
	y           int
	icon        icon.Icon
	initiative  int
	hp          int
	maxHp       int
	ac          int
	str         int
	dex         int
	encumbrance int
	crouching   bool
	weapon      *item.Weapon
	armour      *item.Armour
	inventory   []item.Item
	world       *worldmap.Map
}
