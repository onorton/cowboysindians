package item

import (
	"encoding/json"
	"testing"

	"github.com/onorton/cowboysindians/icon"
)

type readableMarshallingPair struct {
	readable Readable
	result   string
}

var readableMarshallingTests = []readableMarshallingPair{
	{Readable{baseItem{"signpost", "", icon.NewIcon(80, 4), 20, 1000}, "\"Welcome to Deadwood!\""}, "{\"Type\":\"readable\",\"Name\":\"signpost\",\"Owner\":\"\",\"Icon\":{\"Icon\":80,\"Colour\":4},\"Weight\":20,\"Value\":1000,\"Description\":\"\\\"Welcome to Deadwood!\\\"\"}"},
	{Readable{baseItem{"book", "townsman", icon.NewIcon(98, 6), 1, 1000}, "This book has words in it."}, "{\"Type\":\"readable\",\"Name\":\"book\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":98,\"Colour\":6},\"Weight\":1,\"Value\":1000,\"Description\":\"This book has words in it.\"}"},
}

type readableUnmarshallingPair struct {
	readableJson string
	readable     Readable
}

var readableUnmarshallingTests = []readableUnmarshallingPair{
	{"{\"Type\":\"readable\",\"Name\":\"signpost\",\"Owner\":\"\",\"Icon\":{\"Icon\":80,\"Colour\":4},\"Weight\":20,\"Value\":1000,\"Description\":\"\\\"Welcome to Deadwood!\\\"\"}", Readable{baseItem{"signpost", "", icon.NewIcon(80, 4), 20, 1000}, "\"Welcome to Deadwood!\""}},
	{"{\"Type\":\"readable\",\"Name\":\"book\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":98,\"Colour\":6},\"Weight\":1,\"Value\":1000,\"Description\":\"This book has words in it.\"}", Readable{baseItem{"book", "townsman", icon.NewIcon(98, 6), 1, 1000}, "This book has words in it."}},
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
		readable := Readable{}

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

		if readable.description != pair.readable.description {
			t.Error(
				"For", "Description",
				"expected", pair.readable.description,
				"got", readable.description,
			)
		}
	}

}
