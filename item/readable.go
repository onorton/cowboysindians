package item

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/onorton/cowboysindians/icon"
)

type readableAttributes struct {
	Icon        icon.Icon
	Weight      float64
	Value       int
	Description string
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

func NewReadable(name string, values map[string]string) Item {
	item := readableData[name]
	description := item.Description
	for key, value := range values {
		description = strings.Replace(description, "["+key+"]", value, -1)
	}
	var itm Item = &NormalItem{baseItem{name, "", item.Icon, item.Weight, item.Value}, false, &description}
	return itm
}

func GenerateReadable() Item {
	return NewReadable(Choose(readableProbabilities), map[string]string{})
}

func (item *NormalItem) Read() string {
	return *(item.description)
}
