package item

import (
	"encoding/json"
	"testing"

	"github.com/onorton/cowboysindians/icon"
)

type armourMarshallingPair struct {
	armour Armour
	result string
}

var armourMarshallingTests = []armourMarshallingPair{
	{Armour{"leather jacket", icon.NewIcon(91, 100), 1, 2}, "{\"Name\":\"leather jacket\",\"Icon\":{\"Icon\":91,\"Colour\":100},\"Bonus\":1,\"Weight\":2}"},
	{Armour{"metal breastplate", icon.NewIcon(91, 50), 3, 5}, "{\"Name\":\"metal breastplate\",\"Icon\":{\"Icon\":91,\"Colour\":50},\"Bonus\":3,\"Weight\":5}"},
	{Armour{"reinforced leather jacket", icon.NewIcon(91, 70), 2, 3}, "{\"Name\":\"reinforced leather jacket\",\"Icon\":{\"Icon\":91,\"Colour\":70},\"Bonus\":2,\"Weight\":3}"},
}

type armourUnmarshallingPair struct {
	armourJson string
	armour     Armour
}

var armourUnmarshallingTests = []armourUnmarshallingPair{
	{"{\"Name\":\"leather jacket\",\"Icon\":{\"Icon\":91,\"Colour\":100},\"Bonus\":1,\"Weight\":2}", Armour{"leather jacket", icon.NewIcon(91, 100), 1, 2}},
	{"{\"Name\":\"metal breastplate\",\"Icon\":{\"Icon\":91,\"Colour\":50},\"Bonus\":3,\"Weight\":5}", Armour{"metal breastplate", icon.NewIcon(91, 50), 3, 5}},
	{"{\"Name\":\"reinforced leather jacket\",\"Icon\":{\"Icon\":91,\"Colour\":70},\"Bonus\":2,\"Weight\":3}", Armour{"reinforced leather jacket", icon.NewIcon(91, 70), 2, 3}},
}

func TestArmourMarshalling(t *testing.T) {

	for _, pair := range armourMarshallingTests {

		result, err := json.Marshal(&(pair.armour))
		if err != nil {
			t.Error("Failed when marshalling", pair.armour, err)
		}
		if string(result) != pair.result {
			t.Error(
				"For", pair.armour,
				"expected", pair.result,
				"got", string(result),
			)
		}
	}
}

func TestArmourUnmarshalling(t *testing.T) {

	for _, pair := range armourUnmarshallingTests {
		armour := Armour{}

		if err := json.Unmarshal([]byte(pair.armourJson), &armour); err != nil {
			t.Error("Failed when unmarshalling", pair.armourJson, err)
		}
		if armour.name != pair.armour.name {
			t.Error(
				"For", "Name",
				"expected", pair.armour.name,
				"got", armour.name,
			)
		}

		if armour.ic != pair.armour.ic {
			t.Error(
				"For", "Icon",
				"expected", pair.armour.ic,
				"got", armour.ic,
			)
		}

		if armour.w != pair.armour.w {
			t.Error(
				"For", "Weight",
				"expected", pair.armour.w,
				"got", armour.w,
			)
		}

		if armour.bonus != pair.armour.bonus {
			t.Error(
				"For", "Bonus",
				"expected", pair.armour.bonus,
				"got", armour.bonus,
			)
		}
	}

}
