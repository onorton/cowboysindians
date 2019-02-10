package item

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"

	"github.com/onorton/cowboysindians/icon"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type itemAttributes struct {
	Icon   icon.Icon
	Weight float64
}

type ItemDefinition struct {
	Category string
	Name     string
	Amount   int
}

var itemData map[string]itemAttributes

func fetchItemData() {
	data, err := ioutil.ReadFile("data/item.json")
	check(err)
	var iD map[string]itemAttributes
	err = json.Unmarshal(data, &iD)
	check(err)
	itemData = iD
}

type NormalItem struct {
	name string
	ic   icon.Icon
	w    float64
}

func NewItem(name string) Item {
	item := itemData[name]
	var itm Item = &NormalItem{name, item.Icon, item.Weight}
	return itm
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

	buffer.WriteString(fmt.Sprintf("\"Weight\":%s", weightValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (item *NormalItem) UnmarshalJSON(data []byte) error {

	type itemJson struct {
		Name   string
		Icon   icon.Icon
		Weight float64
	}
	var v itemJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	item.name = v.Name
	item.ic = v.Icon
	item.w = v.Weight

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

func (item *NormalItem) GetName() string {
	return item.name
}
func (item *NormalItem) Render(x, y int) {

	item.ic.Render(x, y)
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
	Render(int, int)
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
	GetWeight() float64
}
