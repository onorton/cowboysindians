package item

import (
	"encoding/json"
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/icon"
	"hash/fnv"
	"io/ioutil"
	"strconv"
	"strings"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type ItemAttributes struct {
	Colour termbox.Attribute
	Icon   rune
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
	var itm Item = &NormalItem{name, icon.NewIcon(item.Icon, item.Colour), item.Weight}
	return itm
}

func (item *NormalItem) Serialize() string {
	if item == nil {
		return ""
	}
	return fmt.Sprintf("Item{%s %f %s}", item.name, item.w, item.ic.Serialize())
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
	item.ic = icon.Deserialize(itemAttributes[2])
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
