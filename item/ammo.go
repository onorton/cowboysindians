package item

import (
	"encoding/json"
	"io/ioutil"

	"github.com/onorton/cowboysindians/icon"
)

type AmmoAttributes struct {
	Icon        icon.Icon
	Type        WeaponType
	Weight      float64
	Value       int
	Probability float64
}

var ammoData map[string]AmmoAttributes
var ammoProbabilities map[string]float64

func fetchAmmoData() {
	data, err := ioutil.ReadFile("data/ammo.json")
	check(err)
	var aD map[string]AmmoAttributes
	err = json.Unmarshal(data, &aD)
	check(err)
	ammoData = aD

	ammoProbabilities = make(map[string]float64)
	for name, attributes := range ammoData {
		ammoProbabilities[name] = attributes.Probability
	}
}

func NewAmmo(name string) *Item {
	ammo := ammoData[name]
	return &Item{name, "", ammo.Icon, ammo.Weight, ammo.Value, nil, nil, nil, ammo.Type, nil, nil, nil}
}

func GenerateAmmo() *Item {
	return NewAmmo(Choose(ammoProbabilities))
}
