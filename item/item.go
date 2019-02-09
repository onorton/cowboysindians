package item

import (
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

type ItemAttributes struct {
	Icon   icon.Icon
	Weight float64
}

type ItemDefinition struct {
	Category string
	Name     string
	Amount   int
}

var itemData map[string]ItemAttributes = fetchItemData()

func fetchItemData() map[string]ItemAttributes {
	data, err := ioutil.ReadFile("data/item.json")
	check(err)
	var eD map[string]ItemAttributes
	err = json.Unmarshal(data, &eD)
	check(err)
	return eD
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

type Item interface {
	GetKey() rune
	GetName() string
	Render(int, int)
	Serialize() string
	GetWeight() float64
}
