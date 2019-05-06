package item

import (
	"encoding/json"
	"testing"

	"github.com/onorton/cowboysindians/icon"
)

type readableMarshallingPair struct {
	readable Item
	result   string
}

var readableMarshallingTests = []readableMarshallingPair{
	{Item{"signpost", "", icon.NewIcon(80, 4), 20, 1000, map[string]component{"readable": ReadableComponent{"\"Welcome to Deadwood!\""}}}, "{\"Name\":\"signpost\",\"Owner\":\"\",\"Icon\":{\"Icon\":80,\"Colour\":4},\"Weight\":20,\"Value\":1000,\"Components\":{\"readable\":{\"Description\":\"\\\"Welcome to Deadwood!\\\"\"}}}"},
	{Item{"book", "townsman", icon.NewIcon(98, 6), 1, 1000, map[string]component{"readable": ReadableComponent{"This book has words in it."}}}, "{\"Name\":\"book\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":98,\"Colour\":6},\"Weight\":1,\"Value\":1000,\"Components\":{\"readable\":{\"Description\":\"This book has words in it.\"}}}"},
}

type readableUnmarshallingPair struct {
	readableJson string
	readable     Item
}

var readableUnmarshallingTests = []readableUnmarshallingPair{
	{"{\"Name\":\"signpost\",\"Owner\":\"\",\"Icon\":{\"Icon\":80,\"Colour\":4},\"Weight\":20,\"Value\":1000,\"Components\":{\"readable\":{\"Description\":\"\\\"Welcome to Deadwood!\\\"\"}}}", Item{"signpost", "", icon.NewIcon(80, 4), 20, 1000, map[string]component{"readable": ReadableComponent{"\"Welcome to Deadwood!\""}}}},
	{"{\"Name\":\"book\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":98,\"Colour\":6},\"Weight\":1,\"Value\":1000,\"Components\":{\"readable\":{\"Description\":\"This book has words in it.\"}}}", Item{"book", "townsman", icon.NewIcon(98, 6), 1, 1000, map[string]component{"readable": ReadableComponent{"This book has words in it."}}}},
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
		readable := Item{}

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

		if readable.HasComponent("cover") != pair.readable.HasComponent("cover") {
			t.Error(
				"For", "Gives cover",
				"expected", pair.readable.HasComponent("cover"),
				"got", readable.HasComponent("cover"),
			)
		}

		if readable.HasComponent("corpse") != pair.readable.HasComponent("corpse") {
			t.Error(
				"For", "Corpse",
				"expected", pair.readable.HasComponent("corpse"),
				"got", readable.HasComponent("corpse"),
			)
		}

		if readable.Component("readable").(ReadableComponent).Description != pair.readable.Component("readable").(ReadableComponent).Description {
			t.Error(
				"For", "Description",
				"expected", pair.readable.Component("readable").(ReadableComponent).Description,
				"got", readable.Component("readable").(ReadableComponent).Description,
			)
		}
	}

}
