package item

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/onorton/cowboysindians/icon"
)

type itemAttributes struct {
	Icon        icon.Icon
	Weight      float64
	Value       int
	Cover       bool
	Probability float64
}

var normalItemData map[string]itemAttributes
var normalItemProbabilities map[string]float64

func fetchItemData() {
	data, err := ioutil.ReadFile("data/item.json")
	check(err)
	var iD map[string]itemAttributes
	err = json.Unmarshal(data, &iD)
	check(err)
	normalItemData = iD

	normalItemProbabilities = make(map[string]float64)
	for name, attributes := range normalItemData {
		normalItemProbabilities[name] = attributes.Probability
	}
}

type NormalItem struct {
	baseItem
	cover bool
}

func NewNormalItem(name string) Item {
	item := normalItemData[name]
	var itm Item = &NormalItem{baseItem{name, "", item.Icon, item.Weight, item.Value}, item.Cover}
	return itm
}

func Money(amount int) Item {
	return &NormalItem{baseItem{"money", "", icon.NewIcon('$', 4), 0, amount}, false}
}

func GenerateNormalItem() Item {
	return NewNormalItem(Choose(normalItemProbabilities))
}

func (item *NormalItem) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	nameValue, err := json.Marshal(item.name)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Name\":%s,", nameValue))

	ownerValue, err := json.Marshal(item.owner)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Owner\":%s,", ownerValue))

	iconValue, err := json.Marshal(item.ic)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Icon\":%s,", iconValue))

	weightValue, err := json.Marshal(item.w)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Weight\":%s,", weightValue))
	buffer.WriteString(fmt.Sprintf("\"Value\":%d,", item.v))

	coverValue, err := json.Marshal(item.cover)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Cover\":%s", coverValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (item *NormalItem) UnmarshalJSON(data []byte) error {

	type itemJson struct {
		Name   string
		Owner  string
		Icon   icon.Icon
		Weight float64
		Value  int
		Cover  bool
	}
	var v itemJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	item.name = v.Name
	item.owner = v.Owner
	item.ic = v.Icon
	item.w = v.Weight
	item.v = v.Value
	item.cover = v.Cover

	return nil
}

func (item *NormalItem) Owned(id string) bool {
	if item.owner == "" {
		return true
	}
	return item.owner == id
}

func (item *NormalItem) TransferOwner(newOwner string) {
	// Only assign owner if item not owned
	if item.owner == "" {
		item.owner = newOwner
	}
}

func (item *NormalItem) GivesCover() bool {
	return item.cover
}
