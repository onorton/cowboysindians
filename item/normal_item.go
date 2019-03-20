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
	name  string
	ic    icon.Icon
	w     float64
	v     int
	cover bool
}

func NewNormalItem(name string) Item {
	item := normalItemData[name]
	var itm Item = &NormalItem{name, item.Icon, item.Weight, item.Value, item.Cover}
	return itm
}

func Money(amount int) Item {
	return &NormalItem{"money", icon.NewIcon('$', 4), 0, amount, false}
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
	item.ic = v.Icon
	item.w = v.Weight
	item.v = v.Value
	item.cover = v.Cover

	return nil
}

func (item *NormalItem) Render() ui.Element {
	return item.ic.Render()
}

func (item *NormalItem) GetName() string {
	return item.name
}

func (item *NormalItem) GetKey() rune {
	h := fnv.New32()
	h.Write([]byte(item.name))
	key := rune(33 + h.Sum32()%93)
	if key == '*' {
		key++
	}
	return key
}

func (item *NormalItem) GetWeight() float64 {
	return item.w
}

func (item *NormalItem) GetValue() int {
	return item.v
}

func (item *NormalItem) GivesCover() bool {
	return item.cover
}
