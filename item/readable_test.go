package item

import (
	"encoding/json"
	"testing"

	"github.com/onorton/cowboysindians/icon"
)

type readableMarshallingPair struct {
	readable NormalItem
	result   string
}

var signpostDescription string = "\"Welcome to Deadwood!\""
var bookDescription string = "This book has words in it."

var readableMarshallingTests = []readableMarshallingPair{
	{NormalItem{baseItem{"signpost", "", icon.NewIcon(80, 4), 20, 1000}, false, &signpostDescription, false, NoAmmo, nil, nil}, "{\"Type\":\"normal\",\"Name\":\"signpost\",\"Owner\":\"\",\"Icon\":{\"Icon\":80,\"Colour\":4},\"Weight\":20,\"Value\":1000,\"Cover\":false,\"Description\":\"\\\"Welcome to Deadwood!\\\"\",\"Corpse\":false,\"AmmoType\":0,\"Armour\":null,\"Weapon\":null}"},
	{NormalItem{baseItem{"book", "townsman", icon.NewIcon(98, 6), 1, 1000}, false, &bookDescription, false, NoAmmo, nil, nil}, "{\"Type\":\"normal\",\"Name\":\"book\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":98,\"Colour\":6},\"Weight\":1,\"Value\":1000,\"Cover\":false,\"Description\":\"This book has words in it.\",\"Corpse\":false,\"AmmoType\":0,\"Armour\":null,\"Weapon\":null}"},
}

type readableUnmarshallingPair struct {
	readableJson string
	readable     NormalItem
}

var readableUnmarshallingTests = []readableUnmarshallingPair{
	{"{\"Type\":\"normal\",\"Name\":\"signpost\",\"Owner\":\"\",\"Icon\":{\"Icon\":80,\"Colour\":4},\"Weight\":20,\"Value\":1000,\"Description\":\"\\\"Welcome to Deadwood!\\\"\",\"Corpse\":false}", NormalItem{baseItem{"signpost", "", icon.NewIcon(80, 4), 20, 1000}, false, &signpostDescription, false, NoAmmo, nil, nil}},
	{"{\"Type\":\"normal\",\"Name\":\"book\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":98,\"Colour\":6},\"Weight\":1,\"Value\":1000,\"Description\":\"This book has words in it.\",\"Corpse\":false}", NormalItem{baseItem{"book", "townsman", icon.NewIcon(98, 6), 1, 1000}, false, &bookDescription, false, NoAmmo, nil, nil}},
}

func TestReadableMarshalling(t *testing.T) {

	for _, pair := range readableMarshallingTests {

		result, err := json.Marshal(&(pair.readable))
		if err != nil {
			t.Error("Failed when marshalling", pair.readable, err)
		}
		if string(result) != pair.result {
			t.Error(
				"For", pair.readable,
				"expected", pair.result,
				"got", string(result),
			)
		}
	}
}

func TestReadableUnmarshalling(t *testing.T) {

	for _, pair := range readableUnmarshallingTests {
		readable := NormalItem{}

		if err := json.Unmarshal([]byte(pair.readableJson), &readable); err != nil {
			t.Error("Failed when unmarshalling", pair.readableJson, err)
		}
		if readable.name != pair.readable.name {
			t.Error(
				"For", "Name",
				"expected", pair.readable.name,
				"got", readable.name,
			)
		}

		if readable.owner != pair.readable.owner {
			t.Error(
				"For", "Owner",
				"expected", pair.readable.owner,
				"got", readable.owner,
			)
		}

		if readable.ic != pair.readable.ic {
			t.Error(
				"For", "Icon",
				"expected", pair.readable.ic,
				"got", readable.ic,
			)
		}

		if readable.w != pair.readable.w {
			t.Error(
				"For", "Weight",
				"expected", pair.readable.w,
				"got", readable.w,
			)
		}

		if readable.v != pair.readable.v {
			t.Error(
				"For", "Value",
				"expected", pair.readable.v,
				"got", readable.v,
			)
		}

		if readable.cover != pair.readable.cover {
			t.Error(
				"For", "Gives Cover",
				"expected", pair.readable.cover,
				"got", readable.cover,
			)
		}

		if *(readable.description) != *(pair.readable.description) {
			t.Error(
				"For", "Description",
				"expected", *(pair.readable.description),
				"got", *(readable.description),
			)
		}

		if readable.corpse != pair.readable.corpse {
			t.Error(
				"For", "Corpse",
				"expected", pair.readable.corpse,
				"got", readable.corpse,
			)
		}
	}

}
