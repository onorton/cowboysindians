package item

import (
	"encoding/json"
	"io/ioutil"

	"github.com/onorton/cowboysindians/icon"
)

type ArmourAttributes struct {
	Icon        icon.Icon
	Bonus       int
	Weight      float64
	Value       int
	Probability float64
}

var armourData map[string]ArmourAttributes
var armourProbabilities map[string]float64

func fetchArmourData() {
	data, err := ioutil.ReadFile("data/armour.json")
	check(err)
	var aD map[string]ArmourAttributes
	err = json.Unmarshal(data, &aD)
	check(err)
	armourData = aD

	armourProbabilities = make(map[string]float64)
	for name, attributes := range armourData {
		armourProbabilities[name] = attributes.Probability
	}
}

type armourComponent struct {
	Bonus int
}

func NewArmour(name string) *NormalItem {
	armour := armourData[name]
	return &NormalItem{baseItem{name, "", armour.Icon, armour.Weight, armour.Value}, false, nil, false, NoAmmo, &armourComponent{armour.Bonus}, nil, nil}
}

func GenerateArmour() Item {
	return NewArmour(Choose(armourProbabilities))
}
