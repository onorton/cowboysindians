package item

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type Armour struct {
	baseItem
	bonus int
}

func NewArmour(name string) *Armour {
	armour := armourData[name]
	return &Armour{baseItem{name, "", armour.Icon, armour.Weight, armour.Value}, armour.Bonus}
}

func GenerateArmour() Item {
	return NewArmour(Choose(armourProbabilities))
}

func (armour *Armour) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	nameValue, err := json.Marshal(armour.name)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Name\":%s,", nameValue))

	ownerValue, err := json.Marshal(armour.owner)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Owner\":%s,", ownerValue))

	iconValue, err := json.Marshal(armour.ic)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Icon\":%s,", iconValue))

	bonusValue, err := json.Marshal(armour.bonus)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Bonus\":%s,", bonusValue))

	weightValue, err := json.Marshal(armour.w)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Weight\":%s,", weightValue))
	buffer.WriteString(fmt.Sprintf("\"Value\":%d", armour.v))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (armour *Armour) UnmarshalJSON(data []byte) error {

	type armourJson struct {
		Name   string
		Owner  string
		Icon   icon.Icon
		Bonus  *int
		Weight float64
		Value  int
	}
	var v armourJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if v.Bonus == nil {
		return fmt.Errorf("The Bonus field is required")
	}

	armour.name = v.Name
	armour.owner = v.Owner
	armour.ic = v.Icon
	armour.bonus = *(v.Bonus)
	armour.w = v.Weight
	armour.v = v.Value

	return nil
}

func (armour *Armour) Owned(id string) bool {
	if armour.owner == "" {
		return true
	}
	return armour.owner == id
}

func (armour *Armour) TransferOwner(newOwner string) {
	// Only assign owner if item not owned
	if armour.owner == "" {
		armour.owner = newOwner
	}
}

func (armour *Armour) GetACBonus() int {
	return armour.bonus
}

func (armour *Armour) GivesCover() bool {
	return false
}
