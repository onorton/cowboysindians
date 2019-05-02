package item

import (
	"encoding/json"
	"io/ioutil"

	"github.com/onorton/cowboysindians/icon"
)

type ConsumableAttributes struct {
	Icon        icon.Icon
	Weight      float64
	Value       int
	Effects     map[string][]Effect
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

type consumableComponent struct {
	Effects map[string][]Effect
}

func NewConsumable(name string) Item {
	consumable := consumableData[name]
	var itm Item = &NormalItem{baseItem{name, "", consumable.Icon, consumable.Weight, consumable.Value}, false, nil, false, NoAmmo, nil, nil, &consumableComponent{consumable.Effects}}
	return itm
}

func GenerateConsumable() Item {
	return NewConsumable(Choose(consumableProbabilities))
}
