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
	Weight      float64
	Value       int
	Cover       bool
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
	name        string
	owner       string
	ic          icon.Icon
	w           float64
	v           int
	components  map[string]component
	description *string
	ammoType    *WeaponType
	armour      *armourComponent
	weapon      *weaponComponent
}

var typeProbabilities = map[string]float64{
	"armour":      0.1,
	"ammo":        0.45,
	"consumable":  0.2,
	"normal item": 0.1,
	"weapon":      0.1,
	"readable":    0.05,
}

type ItemDefinition struct {
	Category string
	Name     string
	Amount   int
}

func Choose(probabilites map[string]float64) string {
	max := 0.0

	for _, probability := range probabilites {
		if probability > 0 {
			inverse := 1.0 / probability
			if inverse > max {
				max = inverse
			}
		}
	}
	items := make([]string, 0)

	for name, probability := range probabilites {
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
	components := map[string]component{}
	if item.Cover {
		components["cover"] = tag{}
	}
	return &Item{name, "", item.Icon, item.Weight, item.Value, components, nil, nil, nil, nil}
}

func Money(amount int) *Item {
	return &Item{"money", "", icon.NewIcon('$', 4), 0, amount, map[string]component{}, nil, nil, nil, nil}
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

	buffer.WriteString(fmt.Sprintf("\"Components\":%s,", componentsValue))

	descriptionValue, err := json.Marshal(item.description)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Description\":%s,", descriptionValue))

	ammoValue, err := json.Marshal(item.ammoType)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"AmmoType\":%s,", ammoValue))

	armourValue, err := json.Marshal(item.armour)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Armour\":%s,", armourValue))

	weaponValue, err := json.Marshal(item.weapon)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Weapon\":%s", weaponValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (item *Item) UnmarshalJSON(data []byte) error {

	type itemJson struct {
		Name        string
		Owner       string
		Icon        icon.Icon
		Weight      float64
		Value       int
		Components  map[string]interface{}
		Description *string
		AmmoType    *WeaponType
		Armour      *armourComponent
		Weapon      *weaponComponent
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
	item.description = v.Description
	item.ammoType = v.AmmoType
	item.armour = v.Armour
	item.weapon = v.Weapon

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

func (item *Item) IsReadable() bool {
	return item.description != nil
}

func (item *Item) IsAmmo() bool {
	return item.ammoType != nil
}

func (item *Item) IsArmour() bool {
	return item.armour != nil
}

func (item *Item) ACBonus() int {
	if item.armour != nil {
		return item.armour.Bonus
	}
	return 0
}

func (item *Item) IsWeapon() bool {
	return item.weapon != nil
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
