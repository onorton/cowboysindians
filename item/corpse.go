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

type corpseAttributes struct {
	Icon        icon.Icon
	Weight      float64
	Value       int
	Cover       bool
	Probability float64
}

var corpseData map[string]corpseAttributes

func fetchCorpseData() {
	data, err := ioutil.ReadFile("data/corpse.json")
	check(err)
	var cD map[string]corpseAttributes
	err = json.Unmarshal(data, &cD)
	check(err)
	corpseData = cD
}

type Corpse struct {
	name  string
	owner string
	ic    icon.Icon
	w     float64
	v     int
}

func NewCorpse(corpseType string, owner string, ownerName string, ownerIcon icon.Icon) Item {
	corpse := corpseData[corpseType]
	name := fmt.Sprintf("%s's %s", ownerName, corpseType)
	var itm Item = &Corpse{name, owner, icon.NewCorpseIcon(ownerIcon), corpse.Weight, corpse.Value}
	return itm
}

func (corpse *Corpse) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	nameValue, err := json.Marshal(corpse.name)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Name\":%s,", nameValue))

	ownerValue, err := json.Marshal(corpse.owner)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Owner\":%s,", ownerValue))

	iconValue, err := json.Marshal(corpse.ic)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Icon\":%s,", iconValue))

	weightValue, err := json.Marshal(corpse.w)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Weight\":%s,", weightValue))
	buffer.WriteString(fmt.Sprintf("\"Value\":%d", corpse.v))

	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (corpse *Corpse) UnmarshalJSON(data []byte) error {

	type corpseJson struct {
		Name   string
		Owner  string
		Icon   icon.Icon
		Weight float64
		Value  int
	}
	var v corpseJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	corpse.name = v.Name
	corpse.owner = v.Owner
	corpse.ic = v.Icon
	corpse.w = v.Weight
	corpse.v = v.Value

	return nil
}

func (corpse *Corpse) Render() ui.Element {
	return corpse.ic.Render()
}

func (corpse *Corpse) GetName() string {
	return corpse.name
}

func (corpse *Corpse) GetKey() rune {
	h := fnv.New32()
	h.Write([]byte(corpse.name))
	key := rune(33 + h.Sum32()%93)
	if key == '*' {
		key++
	}
	return key
}

func (corpse *Corpse) GetWeight() float64 {
	return corpse.w
}

func (corpse *Corpse) GetValue() int {
	return corpse.v
}

func (corpse *Corpse) Owner() string {
	return corpse.owner
}

func (corpse *Corpse) GivesCover() bool {
	return false
}
