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

type Ammo struct {
	name string
	ic   icon.Icon
	t    WeaponType
	w    float64
	v    int
}

func NewAmmo(name string) Item {
	ammo := ammoData[name]
	var itm Item = &Ammo{name, ammo.Icon, ammo.Type, ammo.Weight, ammo.Value}
	return itm
}

func GenerateAmmo() Item {
	return NewAmmo(Choose(ammoProbabilities))
}

func (ammo *Ammo) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	nameValue, err := json.Marshal(ammo.name)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Name\":%s,", nameValue))

	iconValue, err := json.Marshal(ammo.ic)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Icon\":%s,", iconValue))

	typeValue, err := json.Marshal(ammo.t)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Type\":%s,", typeValue))

	weightValue, err := json.Marshal(ammo.w)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Weight\":%s,", weightValue))
	buffer.WriteString(fmt.Sprintf("\"Value\":%d", ammo.v))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (ammo *Ammo) UnmarshalJSON(data []byte) error {

	type ammoJson struct {
		Name   string
		Icon   icon.Icon
		Type   *WeaponType
		Weight float64
		Value  int
	}
	var v ammoJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if v.Type == nil {
		return fmt.Errorf("The Type field is required")
	}

	ammo.name = v.Name
	ammo.ic = v.Icon
	ammo.t = *(v.Type)
	ammo.w = v.Weight
	ammo.v = v.Value

	return nil
}

func (ammo *Ammo) GetName() string {
	return ammo.name
}
func (ammo *Ammo) Render() ui.Element {
	return ammo.ic.Render()
}

func (ammo *Ammo) GetKey() rune {
	h := fnv.New32()
	h.Write([]byte(ammo.name))
	key := rune(33 + h.Sum32()%93)
	if key == '*' {
		key++
	}
	return key
}

func (ammo *Ammo) GetWeight() float64 {
	return ammo.w
}

func (ammo *Ammo) GetValue() int {
	return ammo.v
}

func (ammo *Ammo) GivesCover() bool {
	return false
}
