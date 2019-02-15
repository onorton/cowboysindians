package item

import (
	"encoding/json"
	"testing"

	"github.com/onorton/cowboysindians/icon"
)

type marshallingPair struct {
	item   NormalItem
	result string
}

var marshallingTests = []marshallingPair{
	{NormalItem{"gem", icon.NewIcon(42, 4), 2, false}, "{\"Name\":\"gem\",\"Icon\":{\"Icon\":42,\"Colour\":4},\"Weight\":2,\"Cover\":false}"},
	{NormalItem{"stick", icon.NewIcon(30, 7), 5, false}, "{\"Name\":\"stick\",\"Icon\":{\"Icon\":30,\"Colour\":7},\"Weight\":5,\"Cover\":false}"},
	{NormalItem{"bowl", icon.NewIcon(66, 10), 3, false}, "{\"Name\":\"bowl\",\"Icon\":{\"Icon\":66,\"Colour\":10},\"Weight\":3,\"Cover\":false}"},
	{NormalItem{"barrel", icon.NewIcon(111, 0), 30, true}, "{\"Name\":\"barrel\",\"Icon\":{\"Icon\":111,\"Colour\":0},\"Weight\":30,\"Cover\":true}"},
}

type unmarshallingPair struct {
	itemJson string
	item     NormalItem
}

var unmarshallingTests = []unmarshallingPair{
	{"{\"Name\":\"gem\",\"Icon\":{\"Icon\":42,\"Colour\":4},\"Weight\":2,\"Cover\":false}", NormalItem{"gem", icon.NewIcon(42, 4), 2, false}},
	{"{\"Name\":\"stick\",\"Icon\":{\"Icon\":30,\"Colour\":7},\"Weight\":5,\"Cover\":false}", NormalItem{"stick", icon.NewIcon(30, 7), 5, false}},
	{"{\"Name\":\"bowl\",\"Icon\":{\"Icon\":66,\"Colour\":10},\"Weight\":3,\"Cover\":false}", NormalItem{"bowl", icon.NewIcon(66, 10), 3, false}},
	{"{\"Name\":\"barrel\",\"Icon\":{\"Icon\":111,\"Colour\":0},\"Weight\":30,\"Cover\":true}", NormalItem{"barrel", icon.NewIcon(111, 0), 30, true}},
}

func TestMarshalling(t *testing.T) {

	for _, pair := range marshallingTests {

		result, err := json.Marshal(&(pair.item))
		if err != nil {
			t.Error("Failed when marshalling", pair.item, err)
		}
		if string(result) != pair.result {
			t.Error(
				"For", pair.item,
				"expected", pair.result,
				"got", string(result),
			)
		}
	}
}

func TestUnmarshalling(t *testing.T) {

	for _, pair := range unmarshallingTests {
		item := NormalItem{}

		if err := json.Unmarshal([]byte(pair.itemJson), &item); err != nil {
			t.Error("Failed when unmarshalling", pair.itemJson, err)
		}
		if item.name != pair.item.name {
			t.Error(
				"For", "Name",
				"expected", pair.item.name,
				"got", item.name,
			)
		}

		if item.ic != pair.item.ic {
			t.Error(
				"For", "Icon",
				"expected", pair.item.ic,
				"got", item.ic,
			)
		}

		if item.w != pair.item.w {
			t.Error(
				"For", "Weight",
				"expected", pair.item.w,
				"got", item.w,
			)
		}
	}

}
