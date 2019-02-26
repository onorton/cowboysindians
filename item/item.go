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
	Cover       bool
	Probability float64
}

type ItemDefinition struct {
	Category string
	Name     string
	Amount   int
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

type NormalItem struct {
	name  string
	ic    icon.Icon
	w     float64
	cover bool
}

func NewNormalItem(name string) Item {
	item := normalItemData[name]
	var itm Item = &NormalItem{name, item.Icon, item.Weight, item.Cover}
	return itm
}

func SelectItem(probabilites map[string]float64) string {
	max := 0.0

	for _, probability := range probabilites {
		inverse := 1.0 / probability
		if inverse > max {
			max = inverse
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

func GenerateNormalItem() Item {
	return NewNormalItem(SelectItem(normalItemProbabilities))
}

func (item *NormalItem) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	nameValue, err := json.Marshal(item.name)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Name\":%s,", nameValue))

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

	coverValue, err := json.Marshal(item.cover)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Cover\":%s", coverValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (item *NormalItem) UnmarshalJSON(data []byte) error {

	type itemJson struct {
		Name   string
		Icon   icon.Icon
		Weight float64
		Cover  bool
	}
	var v itemJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	item.name = v.Name
	item.ic = v.Icon
	item.w = v.Weight
	item.cover = v.Cover

	return nil
}

type ItemList []Item

func (itemList *ItemList) UnmarshalJSON(data []byte) error {
	var rawItems []interface{}

	if err := json.Unmarshal(data, &rawItems); err != nil {
		return err
	}

	items := []Item{}

	//convert raw items back into byte data and unmarshal individually,
	for _, rawItem := range rawItems {
		itemJson, err := json.Marshal(rawItem)
		check(err)

		//Tying unmarshalling for each item type

		//armour
		armour := &Armour{}
		err = json.Unmarshal(itemJson, armour)
		if err == nil {
			items = append(items, armour)
			continue
		}

		//consumable
		consumable := &Consumable{}
		err = json.Unmarshal(itemJson, consumable)
		if err == nil {
			items = append(items, consumable)
			continue
		}

		//weapon
		weapon := &Weapon{}
		err = json.Unmarshal(itemJson, weapon)
		if err == nil {
			items = append(items, weapon)
			continue
		}

		//ammo
		ammo := &Ammo{}
		err = json.Unmarshal(itemJson, ammo)
		if err == nil {
			items = append(items, ammo)
			continue
		}

		//Must be a plain ordinary item
		item := &NormalItem{}
		err = json.Unmarshal(itemJson, item)
		if err == nil {
			items = append(items, item)
			continue
		} else {
			return err
		}
	}

	*itemList = ItemList(items)

	return nil
}

func (item *NormalItem) Render() ui.Element {
	return item.ic.Render()
}

func (item *NormalItem) GetName() string {
	return item.name
}

func (item *NormalItem) GetKey() rune {
	h := fnv.New32()
	h.Write([]byte(item.name))
	key := rune(33 + h.Sum32()%93)
	if key == '*' {
		key++
	}
	return key
}

func (item *NormalItem) GetWeight() float64 {
	return item.w
}

func (item *NormalItem) GivesCover() bool {
	return item.cover
}

func GenerateItem() Item {
	// Pick random type (all same probability)

	n := rand.Intn(5)

	var itm Item = nil
	switch n {
	case 0:
		itm = GenerateAmmo()
	case 1:
		itm = GenerateArmour()
	case 2:
		itm = GenerateConsumable()
	case 3:
		itm = GenerateNormalItem()
	case 4:
		itm = GenerateWeapon()
	}

	return itm
}

func LoadAllData() {
	fetchAmmoData()
	fetchArmourData()
	fetchConsumableData()
	fetchItemData()
	fetchWeaponData()
}

type Item interface {
	GetKey() rune
	GetName() string
	Render() ui.Element
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
	GetWeight() float64
	GivesCover() bool
}
