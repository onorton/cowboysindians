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
	baseItem
	cover       bool
	description *string
	corpse      bool
	ammoType    WeaponType
	armour      *armourComponent
	weapon      *weaponComponent
	consumable  *consumableComponent
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

type Effect struct {
	effect     int
	onMax      bool
	duration   int
	activated  bool
	compounded bool
}

func NewEffect(effect, duration int, onMax bool) *Effect {
	return &Effect{effect, onMax, duration, false, false}
}

func NewInstantEffect(effect int) *Effect {
	return &Effect{effect, false, 1, false, false}
}

func NewOngoingEffect(effect int) *Effect {
	return &Effect{effect, false, -1, false, true}
}

func (e *Effect) Update(value, max int) (int, int) {
	if e.duration != 0 {
		if e.duration > 0 {
			e.duration--
		}

		if !e.activated || e.compounded {
			e.activated = true
			if e.onMax {
				return value, max + e.effect
			} else {
				return value + e.effect, max
			}
		}
	} else if e.onMax {
		// Return maximum to original value if applicable
		return value, max - e.effect
	}
	return value, max
}

func (e *Effect) Expired() bool {
	return e.duration == 0
}

func (e *Effect) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	effectValue, err := json.Marshal(e.effect)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Effect\":%s,", effectValue))

	onMaxValue, err := json.Marshal(e.onMax)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"OnMax\":%s,", onMaxValue))

	durationValue, err := json.Marshal(e.duration)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Duration\":%s,", durationValue))

	activatedValue, err := json.Marshal(e.activated)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Activated\":%s,", activatedValue))

	compoundedValue, err := json.Marshal(e.compounded)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Compounded\":%s", compoundedValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (e *Effect) UnmarshalJSON(data []byte) error {

	type effectJson struct {
		Effect     int
		OnMax      bool
		Duration   int
		Activated  bool
		Compounded bool
	}

	var v effectJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	e.effect = v.Effect
	e.onMax = v.OnMax
	e.duration = v.Duration
	e.activated = v.Activated
	e.compounded = v.Compounded

	return nil
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

func (item *baseItem) Render() ui.Element {
	return item.ic.Render()
}

func (item *baseItem) GetName() string {
	return item.name
}

func (item *baseItem) GetKey() rune {
	h := fnv.New32()
	h.Write([]byte(item.name))
	key := rune(33 + h.Sum32()%93)
	if key == '*' {
		key++
	}
	return key
}

func (item *baseItem) GetWeight() float64 {
	return item.w
}

func (item *baseItem) GetValue() int {
	return item.v
}

type baseItem struct {
	name  string
	owner string
	ic    icon.Icon
	w     float64
	v     int
}

func NewNormalItem(name string) *Item {
	item := normalItemData[name]
	return &Item{baseItem{name, "", item.Icon, item.Weight, item.Value}, item.Cover, nil, false, NoAmmo, nil, nil, nil}
}

func Money(amount int) *Item {
	return &Item{baseItem{"money", "", icon.NewIcon('$', 4), 0, amount}, false, nil, false, NoAmmo, nil, nil, nil}
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

	coverValue, err := json.Marshal(item.cover)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Cover\":%s,", coverValue))

	descriptionValue, err := json.Marshal(item.description)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Description\":%s,", descriptionValue))

	corpseValue, err := json.Marshal(item.corpse)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Corpse\":%s,", corpseValue))

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

	buffer.WriteString(fmt.Sprintf("\"Weapon\":%s,", weaponValue))

	consumableValue, err := json.Marshal(item.consumable)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Consumable\":%s", consumableValue))
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
		Cover       bool
		Description *string
		Corpse      bool
		AmmoType    WeaponType
		Armour      *armourComponent
		Weapon      *weaponComponent
		Consumable  *consumableComponent
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
	item.cover = v.Cover
	item.description = v.Description
	item.corpse = v.Corpse
	item.ammoType = v.AmmoType
	item.armour = v.Armour
	item.weapon = v.Weapon
	item.consumable = v.Consumable

	return nil
}

func (item *Item) Owner() string {
	return item.owner
}

func (item *Item) Owned(id string) bool {
	if item.owner == "" || item.corpse {
		return true
	}
	return item.owner == id
}

func (item *Item) TransferOwner(newOwner string) {
	if item.corpse {
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

func (item *Item) IsCorpse() bool {
	return item.corpse
}

func (item *Item) IsAmmo() bool {
	return item.ammoType != NoAmmo
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

func (item *Item) IsConsumable() bool {
	return item.consumable != nil
}

func (item *Item) Effects(attr string) []Effect {
	if item.consumable != nil {
		return item.consumable.Effects[attr]
	}
	return []Effect{}
}

func (item *Item) GivesCover() bool {
	return item.cover
}
