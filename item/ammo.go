package item

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type Ammo struct {
	baseItem
	t WeaponType
}

func NewAmmo(name string) Item {
	ammo := ammoData[name]
	var itm Item = &Ammo{baseItem{name, "", ammo.Icon, ammo.Weight, ammo.Value}, ammo.Type}
	return itm
}

func GenerateAmmo() Item {
	return NewAmmo(Choose(ammoProbabilities))
}

func (ammo *Ammo) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	buffer.WriteString("\"Type\":\"ammo\",")

	nameValue, err := json.Marshal(ammo.name)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Name\":%s,", nameValue))

	ownerValue, err := json.Marshal(ammo.owner)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Owner\":%s,", ownerValue))

	iconValue, err := json.Marshal(ammo.ic)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Icon\":%s,", iconValue))

	ammoTypeValue, err := json.Marshal(ammo.t)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"AmmoType\":%s,", ammoTypeValue))

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
		Name     string
		Owner    string
		Icon     icon.Icon
		AmmoType WeaponType
		Weight   float64
		Value    int
	}
	var v ammoJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	ammo.name = v.Name
	ammo.owner = v.Owner
	ammo.ic = v.Icon
	ammo.t = v.AmmoType
	ammo.w = v.Weight
	ammo.v = v.Value

	return nil
}

func (ammo *Ammo) Owned(id string) bool {
	if ammo.owner == "" {
		return true
	}
	return ammo.owner == id
}

func (ammo *Ammo) TransferOwner(newOwner string) {
	// Only assign owner if item not owned
	if ammo.owner == "" {
		ammo.owner = newOwner
	}
}

func (ammo *Ammo) GivesCover() bool {
	return false
}
