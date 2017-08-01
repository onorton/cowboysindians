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

type ConsumableAttributes struct {
	Colour termbox.Attribute
	Icon   rune
	Weight float64
	Amount int
}

var consumableData map[string]ConsumableAttributes = fetchConsumableData()

func fetchConsumableData() map[string]ConsumableAttributes {
	data, err := ioutil.ReadFile("data/consumable.json")
	check(err)
	var eD map[string]ConsumableAttributes
	err = json.Unmarshal(data, &eD)
	check(err)
	return eD
}

type Consumable struct {
	name   string
	ic     icon.Icon
	w      float64
	amount int
}

func NewConsumable(name string) Item {
	consumable := consumableData[name]
	var itm Item = &Consumable{name, icon.NewIcon(consumable.Icon, consumable.Colour), consumable.Weight, consumable.Amount}
	return itm
}

func (consumable *Consumable) Serialize() string {
	if consumable == nil {
		return ""
	}
	return fmt.Sprintf("Consumable{%s %f %s}", consumable.name, consumable.w, consumable.ic.Serialize())
}

func DeserializeConsumable(consumableString string) Item {

	if len(consumableString) == 0 {
		return nil
	}
	consumableString = consumableString[1 : len(consumableString)-2]
	consumable := new(Consumable)
	consumableAttributes := strings.SplitN(consumableString, " ", 3)
	consumable.name = consumableAttributes[0]
	consumable.w, _ = strconv.ParseFloat(consumableAttributes[1], 64)
	consumable.ic = icon.Deserialize(consumableAttributes[2])
	var itm Item = consumable
	return itm
}

func (consumable *Consumable) GetName() string {
	return consumable.name
}
func (consumable *Consumable) Render(x, y int) {

	consumable.ic.Render(x, y)
}

func (consumable *Consumable) GetKey() rune {
	h := fnv.New32()
	h.Write([]byte(consumable.name))
	key := rune(33 + h.Sum32()%93)
	if key == '*' {
		key++
	}
	return key
}

func (consumable *Consumable) GetWeight() float64 {
	return consumable.w
}

func (consumable *Consumable) GetAmount() int {
	return consumable.amount
}
