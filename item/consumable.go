package item

import (
	"encoding/json"
	"io/ioutil"

	"github.com/onorton/cowboysindians/icon"
)

type ConsumableAttributes struct {
	Icon        icon.Icon
	Components  map[string]interface{}
	Weight      float64
	Value       int
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

type ConsumableComponent struct {
	Effects Effects
}

func NewConsumable(name string) *Item {
	consumable := consumableData[name]
	return &Item{name, "", consumable.Icon, consumable.Weight, consumable.Value, UnmarshalComponents(consumable.Components)}
}

func GenerateConsumable() *Item {
	return NewConsumable(Choose(consumableProbabilities))
}
