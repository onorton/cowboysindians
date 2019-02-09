package item

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"strconv"
	"strings"

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

func (item *NormalItem) Serialize() string {
	if item == nil {
		return ""
	}

	iconJson, err := json.Marshal(item.ic)
	check(err)

	return fmt.Sprintf("Item{%s %f %s}", item.name, item.w, iconJson)
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

func Deserialize(itemString string) Item {

	if len(itemString) == 0 {
		return nil
	}

	itemString = itemString[1 : len(itemString)-2]
	item := new(NormalItem)
	itemAttributes := strings.SplitN(itemString, " ", 3)
	item.name = itemAttributes[0]
	item.w, _ = strconv.ParseFloat(itemAttributes[1], 64)

	err := json.Unmarshal([]byte(itemAttributes[2]), &(item.ic))
	check(err)

	var itm Item = item
	return itm
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
	Serialize() string
	GetWeight() float64
}
