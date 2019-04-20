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
	{Armour{baseItem{"leather jacket", "bandit", icon.NewIcon(91, 100), 2, 1000}, 1}, "{\"Type\":\"armour\",\"Name\":\"leather jacket\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":100},\"Bonus\":1,\"Weight\":2,\"Value\":1000}"},
	{Armour{baseItem{"metal breastplate", "bandit", icon.NewIcon(91, 50), 5, 2000}, 3}, "{\"Type\":\"armour\",\"Name\":\"metal breastplate\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":50},\"Bonus\":3,\"Weight\":5,\"Value\":2000}"},
	{Armour{baseItem{"reinforced leather jacket", "bandit", icon.NewIcon(91, 70), 3, 1500}, 2}, "{\"Type\":\"armour\",\"Name\":\"reinforced leather jacket\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":70},\"Bonus\":2,\"Weight\":3,\"Value\":1500}"},
}

type armourUnmarshallingPair struct {
	armourJson string
	armour     Armour
}

var armourUnmarshallingTests = []armourUnmarshallingPair{
	{"{\"Type\":\"armour\",\"Name\":\"leather jacket\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":100},\"Bonus\":1,\"Weight\":2,\"Value\":1000}", Armour{baseItem{"leather jacket", "bandit", icon.NewIcon(91, 100), 2, 1000}, 1}},
	{"{\"Type\":\"armour\",\"Name\":\"metal breastplate\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":50},\"Bonus\":3,\"Weight\":5,\"Value\":2000}", Armour{baseItem{"metal breastplate", "bandit", icon.NewIcon(91, 50), 5, 2000}, 3}},
	{"{\"Type\":\"armour\",\"Name\":\"reinforced leather jacket\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":70},\"Bonus\":2,\"Weight\":3,\"Value\":1500}", Armour{baseItem{"reinforced leather jacket", "bandit", icon.NewIcon(91, 70), 3, 1500}, 2}},
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

		if armour.v != pair.armour.v {
			t.Error(
				"For", "Value",
				"expected", pair.armour.v,
				"got", armour.v,
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
