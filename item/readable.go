package item

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/onorton/cowboysindians/icon"
)

type readableAttributes struct {
	Icon        icon.Icon
	Components  map[string]interface{}
	Weight      float64
	Value       int
	Probability float64
}

var readableData map[string]readableAttributes
var readableProbabilities map[string]float64

func fetchReadableData() {
	data, err := ioutil.ReadFile("data/readable.json")
	check(err)
	var rD map[string]readableAttributes
	err = json.Unmarshal(data, &rD)
	check(err)
	readableData = rD

	readableProbabilities = make(map[string]float64)
	for name, attributes := range readableData {
		readableProbabilities[name] = attributes.Probability
	}
}

func NewReadable(name string, values map[string]string) *Item {
	item := readableData[name]

	itm := &Item{name, "", item.Icon, item.Weight, item.Value, UnmarshalComponents(item.Components)}
	description := itm.components["readable"].(ReadableComponent).Description
	for key, value := range values {
		description = strings.Replace(description, "["+key+"]", value, -1)
	}
	itm.components["readable"] = ReadableComponent{description}
	return itm
}

func GenerateReadable() *Item {
	return NewReadable(Choose(readableProbabilities), map[string]string{})
}

type ReadableComponent struct {
	Description string
}
