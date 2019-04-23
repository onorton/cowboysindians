package item

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math/rand"

	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/ui"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
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

type ItemList []Item

func (itemList *ItemList) UnmarshalJSON(data []byte) error {
	var rawItems []map[string]interface{}

	if err := json.Unmarshal(data, &rawItems); err != nil {
		return err
	}

	items := []Item{}

	//convert raw items back into byte data and unmarshal individually,
	for _, rawItem := range rawItems {
		itemJson, err := json.Marshal(rawItem)
		check(err)

		switch rawItem["Type"] {

		case "ammo":
			ammo := &Ammo{}
			err = json.Unmarshal(itemJson, ammo)
			check(err)
			items = append(items, ammo)
		case "armour":
			armour := &Armour{}
			err = json.Unmarshal(itemJson, armour)
			check(err)
			items = append(items, armour)
		case "consumable":
			consumable := &Consumable{}
			err = json.Unmarshal(itemJson, consumable)
			check(err)
			items = append(items, consumable)
		case "corpse":
			corpse := &Corpse{}
			err = json.Unmarshal(itemJson, corpse)
			check(err)
			items = append(items, corpse)
		case "normal":
			item := &NormalItem{}
			err = json.Unmarshal(itemJson, item)
			check(err)
			items = append(items, item)
		case "readable":
			readable := &Readable{}
			err = json.Unmarshal(itemJson, readable)
			check(err)
			items = append(items, readable)
		case "weapon":
			weapon := &Weapon{}
			err = json.Unmarshal(itemJson, weapon)
			check(err)
			items = append(items, weapon)
		}
	}
	*itemList = ItemList(items)
	return nil
}

func GenerateItem() Item {

	itemType := Choose(typeProbabilities)

	var itm Item = nil
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
	effect      int
	effectOnMax int
	duration    int
}

func NewEffect(effect, effectOnMax, duration int) *Effect {
	return &Effect{effect, effectOnMax, duration}
}

func NewInstantEffect(effect int) *Effect {
	return &Effect{effect, 0, 1}
}

func (e *Effect) Update(value, max int) (int, int) {
	if e.duration != 0 {
		if e.duration > 0 {
			e.duration--
		}
		return value + e.effect, max + e.effectOnMax
	} else {
		// Return maximum to original value if applicable
		return value, max - e.effectOnMax
	}
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

	effectOnMaxValue, err := json.Marshal(e.effectOnMax)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"EffectOnMax\":%s,", effectOnMaxValue))

	durationValue, err := json.Marshal(e.duration)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Duration\":%s", durationValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (e *Effect) UnmarshalJSON(data []byte) error {

	type effectJson struct {
		Effect      int
		EffectOnMax int
		Duration    int
	}

	var v effectJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	e.effect = v.Effect
	e.effectOnMax = v.EffectOnMax
	e.duration = v.Duration

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

type Item interface {
	GetKey() rune
	GetName() string
	Owned(string) bool
	TransferOwner(string)
	Render() ui.Element
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
	GetWeight() float64
	GetValue() int
	GivesCover() bool
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
