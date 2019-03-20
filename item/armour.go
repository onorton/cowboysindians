package item

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"

	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/ui"
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
	name  string
	ic    icon.Icon
	bonus int
	w     float64
	v     int
}

func NewArmour(name string) *Armour {
	armour := armourData[name]
	return &Armour{name, armour.Icon, armour.Bonus, armour.Weight, armour.Value}
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
	armour.ic = v.Icon
	armour.bonus = *(v.Bonus)
	armour.w = v.Weight
	armour.v = v.Value

	return nil
}

func (armour *Armour) GetName() string {
	return armour.name
}
func (armour *Armour) Render() ui.Element {
	return armour.ic.Render()
}

func (armour *Armour) GetACBonus() int {
	return armour.bonus
}

func (armour *Armour) GetKey() rune {
	h := fnv.New32()
	h.Write([]byte(armour.name))
	key := rune(33 + h.Sum32()%93)
	if key == '*' {
		key++
	}
	return key
}

func (armour *Armour) GetWeight() float64 {
	return armour.w
}

func (armour *Armour) GivesCover() bool {
	return false
}
