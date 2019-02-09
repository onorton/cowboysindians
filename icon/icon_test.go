package icon

import (
	"encoding/json"
	"testing"

	termbox "github.com/nsf/termbox-go"
)

type marshallingPair struct {
	icon   Icon
	result string
}

var marshallingTests = []marshallingPair{
	{Icon{98, 5}, "{\"Icon\":98,\"Colour\":5}"},
	{Icon{67, 0}, "{\"Icon\":67,\"Colour\":0}"},
	{Icon{76, 10}, "{\"Icon\":76,\"Colour\":10}"},
	{Icon{35, 50}, "{\"Icon\":35,\"Colour\":50}"},
	{Icon{126, 255}, "{\"Icon\":126,\"Colour\":255}"},
}

type unmarshallingPair struct {
	iconJson string
	icon     Icon
}

var unmarshallingTests = []unmarshallingPair{
	{"{\"Icon\":98,\"Colour\":5}", Icon{98, 5}},
	{"{\"Icon\":67,\"Colour\":0}", Icon{67, 0}},
	{"{\"Icon\":76,\"Colour\":10}", Icon{76, 10}},
	{"{\"Icon\":35,\"Colour\":50}", Icon{35, 50}},
	{"{\"Icon\":126,\"Colour\":255}", Icon{126, 255}},
}

func Testmarshalling(t *testing.T) {

	for _, pair := range marshallingTests {

		result, err := json.Marshal(pair.icon)
		if err != nil {
			t.Error("Failed when marshalling", pair.icon, err)
		}
		if string(result) != pair.result {
			t.Error(
				"For", pair.icon,
				"expected", pair.result,
				"got", string(result),
			)
		}
	}
}

func TestUnmarshalling(t *testing.T) {

	for _, pair := range unmarshallingTests {
		i := Icon{}

		if err := json.Unmarshal([]byte(pair.iconJson), &i); err != nil {
			t.Error("Failed when unmarshalling", pair.iconJson, err)
		}
		if i.icon != rune(pair.icon.icon) {
			t.Error(
				"For", "Icon",
				"expected", pair.icon.icon,
				"got", i.icon,
			)
		}

		if i.colour != termbox.Attribute(pair.icon.colour) {
			t.Error(
				"For", "Colour",
				"expected", pair.icon.colour,
				"got", i.colour,
			)
		}
	}

}
