package item

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type Readable struct {
	baseItem
	description string
}

func NewReadable(name string, values map[string]string) Item {
	item := readableData[name]
	description := item.Description
	for key, value := range values {
		description = strings.Replace(description, "["+key+"]", value, -1)
	}
	var itm Item = &Readable{baseItem{name, "", item.Icon, item.Weight, item.Value}, description}
	return itm
}

func GenerateReadable() Item {
	return NewReadable(Choose(readableProbabilities), map[string]string{})
}

func (readable *Readable) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	buffer.WriteString("\"Type\":\"readable\",")

	nameValue, err := json.Marshal(readable.name)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Name\":%s,", nameValue))

	ownerValue, err := json.Marshal(readable.owner)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Owner\":%s,", ownerValue))

	iconValue, err := json.Marshal(readable.ic)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Icon\":%s,", iconValue))

	weightValue, err := json.Marshal(readable.w)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Weight\":%s,", weightValue))
	buffer.WriteString(fmt.Sprintf("\"Value\":%d,", readable.v))

	descriptionValue, err := json.Marshal(readable.description)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Description\":%s", descriptionValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (readable *Readable) UnmarshalJSON(data []byte) error {

	type readableJson struct {
		Name        string
		Owner       string
		Icon        icon.Icon
		Weight      float64
		Value       int
		Description string
	}
	var v readableJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	readable.name = v.Name
	readable.owner = v.Owner
	readable.ic = v.Icon
	readable.w = v.Weight
	readable.v = v.Value
	readable.description = v.Description

	return nil
}

func (readable *Readable) Owned(id string) bool {
	if readable.owner == "" {
		return true
	}
	return readable.owner == id
}

func (readable *Readable) TransferOwner(newOwner string) {
	// Only assign owner if item not owned
	if readable.owner == "" {
		readable.owner = newOwner
	}
}

func (readable *Readable) Read() string {
	return readable.description
}

func (readable *Readable) GivesCover() bool {
	return false
}
