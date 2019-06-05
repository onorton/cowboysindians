package item

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"math/rand"

	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/ui"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type itemAttributes struct {
	Icon        icon.Icon
	Components  map[string]interface{}
	Weight      float64
	Value       int
	Probability float64
}

var normalItemData map[string]itemAttributes
var normalItemProbabilities map[string]float64

func fetchItemData() {
	data, err := ioutil.ReadFile("data/item.json")
	check(err)
	var iD map[string]itemAttributes
	err = json.Unmarshal(data, &iD)
	check(err)
	normalItemData = iD

	normalItemProbabilities = make(map[string]float64)
	for name, attributes := range normalItemData {
		normalItemProbabilities[name] = attributes.Probability
	}
}

type Item struct {
	name       string
	owner      string
	ic         icon.Icon
	w          float64
	v          int
	components map[string]component
}

var typeProbabilities = map[string]float64{
	"armour":      0.1,
	"ammo":        0.45,
	"consumable":  0.2,
	"normal item": 0.1,
	"weapon":      0.1,
	"readable":    0.05,
}

type ItemChoice struct {
	Items       map[string]int
	Probability float64
}

// Generates a new item based on the name
func NewItem(itemType string) *Item {
	if _, ok := ammoData[itemType]; ok {
		return NewAmmo(itemType)
	} else if _, ok := armourData[itemType]; ok {
		return NewArmour(itemType)
	} else if _, ok := consumableData[itemType]; ok {
		return NewConsumable(itemType)
	} else if _, ok := normalItemData[itemType]; ok {
		return NewNormalItem(itemType)
	} else if _, ok := readableData[itemType]; ok {
		return NewReadable(itemType, map[string]string{})
	} else if _, ok := weaponData[itemType]; ok {
		return NewWeapon(itemType)
	}
	return nil
}

func RandomItemName(names []string) string {
	probabilities := map[string]float64{}
	for _, name := range names {
		probability := 0.0
		if _, ok := ammoData[name]; ok {
			probability = ammoData[name].Probability
		} else if _, ok := armourData[name]; ok {
			probability = armourData[name].Probability
		} else if _, ok := consumableData[name]; ok {
			probability = consumableData[name].Probability
		} else if _, ok := normalItemData[name]; ok {
			probability = normalItemData[name].Probability
		} else if _, ok := readableData[name]; ok {
			probability = readableData[name].Probability
		} else if _, ok := weaponData[name]; ok {
			probability = weaponData[name].Probability
		}
		probabilities[name] = probability
	}
	return Choose(probabilities)
}

func Choose(probabilities map[string]float64) string {
	max := 0.0

	for _, probability := range probabilities {
		if probability > 0 {
			inverse := 1.0 / probability
			if inverse > max {
				max = inverse
			}
		}
	}
	items := make([]string, 0)

	for name, probability := range probabilities {
		count := int(probability * max)
		for i := 0; i < count; i++ {
			items = append(items, name)
		}
	}

	n := rand.Intn(len(items))
	return items[n]
}

func GenerateItem() *Item {

	itemType := Choose(typeProbabilities)

	var itm *Item = nil
	switch itemType {
	case "ammo":
		itm = GenerateAmmo()
	case "armour":
		itm = GenerateArmour()
	case "consumable":
		itm = GenerateConsumable()
	case "normal item":
		itm = GenerateNormalItem()
	case "weapon":
		itm = GenerateWeapon()
	case "readable":
		itm = GenerateReadable()
	}

	return itm
}

func LoadAllData() {
	fetchAmmoData()
	fetchArmourData()
	fetchConsumableData()
	fetchCorpseData()
	fetchItemData()
	fetchWeaponData()
	fetchReadableData()
}

type Key struct {
}

func (item *Item) Render() ui.Element {
	return item.ic.Render()
}

func (item *Item) GetName() string {
	return item.name
}

func (item *Item) GetKey() rune {
	h := fnv.New32()
	h.Write([]byte(item.name))
	key := rune(33 + h.Sum32()%93)
	if key == '*' {
		key++
	}
	return key
}

func (item *Item) GetWeight() float64 {
	return item.w
}

func (item *Item) GetValue() int {
	return item.v
}

func NewNormalItem(name string) *Item {
	item := normalItemData[name]

	return &Item{name, "", item.Icon, item.Weight, item.Value, UnmarshalComponents(item.Components)}
}

type BreakableComponent struct {
	Chance float64
}

func (bc BreakableComponent) Broken() bool {
	return rand.Float64() < bc.Chance
}

type KeyComponent struct {
	Key    int32
	Chance float64
}

func (kc KeyComponent) Works() bool {
	return rand.Float64() < kc.Chance
}

func NewKey(keyValue int32) *Item {
	key := NewNormalItem("key")
	key.components["key"] = KeyComponent{keyValue, 1}
	return key
}

func Money(amount int) *Item {
	return &Item{"money", "", icon.NewIcon('$', 4), 0, amount, map[string]component{}}
}

func GenerateNormalItem() *Item {
	return NewNormalItem(Choose(normalItemProbabilities))
}

func (item *Item) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	nameValue, err := json.Marshal(item.name)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Name\":%s,", nameValue))

	ownerValue, err := json.Marshal(item.owner)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Owner\":%s,", ownerValue))

	iconValue, err := json.Marshal(item.ic)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Icon\":%s,", iconValue))

	weightValue, err := json.Marshal(item.w)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Weight\":%s,", weightValue))
	buffer.WriteString(fmt.Sprintf("\"Value\":%d,", item.v))

	componentsValue, err := json.Marshal(item.components)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Components\":%s", componentsValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (item *Item) UnmarshalJSON(data []byte) error {

	type itemJson struct {
		Name       string
		Owner      string
		Icon       icon.Icon
		Weight     float64
		Value      int
		Components map[string]interface{}
	}
	var v itemJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	item.name = v.Name
	item.owner = v.Owner
	item.ic = v.Icon
	item.w = v.Weight
	item.v = v.Value
	item.components = UnmarshalComponents(v.Components)

	return nil
}

func UnmarshalComponents(cs map[string]interface{}) map[string]component {
	components := map[string]component{}
	for key, c := range cs {
		componentJson, err := json.Marshal(c)
		check(err)
		var component component = nil
		switch key {
		case "cover":
			component = tag{}
		case "corpse":
			component = tag{}
		case "consumable":
			var consumable ConsumableComponent
			err := json.Unmarshal(componentJson, &consumable)
			check(err)
			component = consumable
		case "weapon":
			var weapon WeaponComponent
			err := json.Unmarshal(componentJson, &weapon)
			check(err)
			component = weapon
		case "readable":
			var readable ReadableComponent
			err := json.Unmarshal(componentJson, &readable)
			check(err)
			component = readable
		case "ammo":
			var ammo AmmoComponent
			err := json.Unmarshal(componentJson, &ammo)
			check(err)
			component = ammo
		case "armour":
			var armour ArmourComponent
			err := json.Unmarshal(componentJson, &armour)
			check(err)
			component = armour
		case "key":
			var key KeyComponent
			err := json.Unmarshal(componentJson, &key)
			check(err)
			component = key
		case "usable":
			component = tag{}
		case "breakable":
			var breakable BreakableComponent
			err := json.Unmarshal(componentJson, &breakable)
			check(err)
			component = breakable
		}
		components[key] = component
	}
	return components
}

func (item *Item) Owner() string {
	return item.owner
}

func (item *Item) Owned(id string) bool {
	if item.owner == "" || item.HasComponent("corpse") {
		return true
	}
	return item.owner == id
}

func (item *Item) TransferOwner(newOwner string) {
	if item.HasComponent("corpse") {
		return
	}

	// Only assign owner if item not owned
	if item.owner == "" {
		item.owner = newOwner
	}
}

func (item *Item) TryBreaking() bool {
	if item.HasComponent("breakable") {
		if item.Component("breakable").(BreakableComponent).Broken() {
			item.name = fmt.Sprintf("broken %s", item.name)
			delete(item.components, "usable")
			delete(item.components, "breakable")
			item.v = item.v / 100
			return true
		}
	}
	return false
}

func (item *Item) HasComponent(key string) bool {
	_, ok := item.components[key]
	return ok
}

func (item *Item) Component(key string) component {
	if item.HasComponent(key) {
		return item.components[key]
	}
	return nil
}

type tag struct{}

type component interface{}
