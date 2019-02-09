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

type ConsumableAttributes struct {
	Icon    icon.Icon
	Weight  float64
	Effects map[string]int
}

var consumableData map[string]ConsumableAttributes

func fetchConsumableData() {
	data, err := ioutil.ReadFile("data/consumable.json")
	check(err)
	var cD map[string]ConsumableAttributes
	err = json.Unmarshal(data, &cD)
	check(err)
	consumableData = cD
}

type Consumable struct {
	name    string
	ic      icon.Icon
	w       float64
	effects map[string]int
}

func NewConsumable(name string) Item {
	consumable := consumableData[name]
	var itm Item = &Consumable{name, consumable.Icon, consumable.Weight, consumable.Effects}
	return itm
}

func (consumable *Consumable) Serialize() string {
	if consumable == nil {
		return ""
	}
	effects := "["
	for effect, amount := range consumable.effects {
		effects += fmt.Sprintf("%s:%d ", effect, amount)
	}
	effects += "]"

	iconJson, err := json.Marshal(consumable.ic)
	check(err)

	return fmt.Sprintf("Consumable{%s %f %s %s}", strings.Replace(consumable.name, " ", "_", -1), consumable.w, iconJson, effects)
}

func DeserializeConsumable(consumableString string) Item {

	if len(consumableString) == 0 {
		return nil
	}
	consumableString = consumableString[1 : len(consumableString)-2]
	consumable := new(Consumable)
	attributesEffects := strings.SplitN(consumableString, "[", 2)
	consumableAttributes := strings.SplitN(attributesEffects[0], " ", 3)

	consumable.name = strings.Replace(consumableAttributes[0], "_", " ", -1)
	consumable.w, _ = strconv.ParseFloat(consumableAttributes[1], 64)

	err := json.Unmarshal([]byte(consumableAttributes[2]), &(consumable.ic))
	check(err)

	effects := strings.Split(attributesEffects[1], " ")
	effects = effects[:len(effects)-1]
	consumable.effects = make(map[string]int)
	for _, effect := range effects {
		nameAmount := strings.Split(effect, ":")
		consumable.effects[nameAmount[0]], _ = strconv.Atoi(nameAmount[1])
	}
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

func (consumable *Consumable) GetEffect(e string) int {
	return consumable.effects[e]
}
