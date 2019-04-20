package item

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/onorton/cowboysindians/icon"
)

type ConsumableAttributes struct {
	Icon        icon.Icon
	Weight      float64
	Value       int
	Effects     map[string]int
	Probability float64
}

var consumableData map[string]ConsumableAttributes
var consumableProbabilities map[string]float64

func fetchConsumableData() {
	data, err := ioutil.ReadFile("data/consumable.json")
	check(err)
	var cD map[string]ConsumableAttributes
	err = json.Unmarshal(data, &cD)
	check(err)
	consumableData = cD

	consumableProbabilities = make(map[string]float64)
	for name, attributes := range consumableData {
		consumableProbabilities[name] = attributes.Probability
	}
}

type Consumable struct {
	baseItem
	effects map[string]int
}

func NewConsumable(name string) Item {
	consumable := consumableData[name]
	var itm Item = &Consumable{baseItem{name, "", consumable.Icon, consumable.Weight, consumable.Value}, consumable.Effects}
	return itm
}

func GenerateConsumable() Item {
	return NewConsumable(Choose(consumableProbabilities))
}

func (item *Consumable) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	buffer.WriteString("\"Type\":\"consumable\",")

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

	effectsValue, err := json.Marshal(item.effects)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Effects\":%s", effectsValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (consumable *Consumable) UnmarshalJSON(data []byte) error {

	type consumableJson struct {
		Name    string
		Owner   string
		Icon    icon.Icon
		Weight  float64
		Value   int
		Effects map[string]int
	}
	var v consumableJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	consumable.name = v.Name
	consumable.owner = v.Owner
	consumable.ic = v.Icon
	consumable.w = v.Weight
	consumable.v = v.Value
	consumable.effects = v.Effects

	return nil
}

func (consumable *Consumable) Owned(id string) bool {
	if consumable.owner == "" {
		return true
	}
	return consumable.owner == id
}

func (consumable *Consumable) TransferOwner(newOwner string) {
	// Only assign owner if item not owned
	if consumable.owner == "" {
		consumable.owner = newOwner
	}
}

func (consumable *Consumable) GetEffect(e string) int {
	return consumable.effects[e]
}

func (consumable *Consumable) GivesCover() bool {
	return false
}
