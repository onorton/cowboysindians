package item

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"

	"github.com/onorton/cowboysindians/icon"
)

type ArmourAttributes struct {
	Icon   icon.Icon
	Bonus  int
	Weight float64
}

var armourData map[string]ArmourAttributes

func fetchArmourData() {
	data, err := ioutil.ReadFile("data/armour.json")
	check(err)
	var aD map[string]ArmourAttributes
	err = json.Unmarshal(data, &aD)
	check(err)
	armourData = aD
}

type Armour struct {
	name  string
	ic    icon.Icon
	bonus int
	w     float64
}

func NewArmour(name string) *Armour {
	armour := armourData[name]
	return &Armour{name, armour.Icon, armour.Bonus, armour.Weight}
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

	buffer.WriteString(fmt.Sprintf("\"Weight\":%s", weightValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (armour *Armour) UnmarshalJSON(data []byte) error {

	type armourJson struct {
		Name   string
		Icon   icon.Icon
		Bonus  int
		Weight float64
	}
	var v armourJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	armour.name = v.Name
	armour.ic = v.Icon
	armour.bonus = v.Bonus
	armour.w = v.Weight

	return nil
}

func (armour *Armour) GetName() string {
	return armour.name
}
func (armour *Armour) Render(x, y int) {

	armour.ic.Render(x, y)
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
