package item

import (
	"encoding/json"
	"testing"

	"github.com/onorton/cowboysindians/icon"
)

type armourMarshallingPair struct {
	armour Item
	result string
}

var armourMarshallingTests = []armourMarshallingPair{
	{Item{"leather jacket", "bandit", icon.NewIcon(91, 100), 2, 1000, map[string]component{"armour": ArmourComponent{1}}}, "{\"Name\":\"leather jacket\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":100},\"Weight\":2,\"Value\":1000,\"Components\":{\"armour\":{\"Bonus\":1}}}"},
	{Item{"metal breastplate", "bandit", icon.NewIcon(91, 50), 5, 2000, map[string]component{"armour": ArmourComponent{3}}}, "{\"Name\":\"metal breastplate\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":50},\"Weight\":5,\"Value\":2000,\"Components\":{\"armour\":{\"Bonus\":3}}}"},
	{Item{"reinforced leather jacket", "bandit", icon.NewIcon(91, 70), 3, 1500, map[string]component{"armour": ArmourComponent{2}}}, "{\"Name\":\"reinforced leather jacket\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":70},\"Weight\":3,\"Value\":1500,\"Components\":{\"armour\":{\"Bonus\":2}}}"},
}

type armourUnmarshallingPair struct {
	armourJson string
	armour     Item
}

var armourUnmarshallingTests = []armourUnmarshallingPair{
	{"{\"Name\":\"leather jacket\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":100},\"Weight\":2,\"Value\":1000,\"Components\":{\"armour\":{\"Bonus\":1}}}", Item{"leather jacket", "bandit", icon.NewIcon(91, 100), 2, 1000, map[string]component{"armour": ArmourComponent{1}}}},
	{"{\"Name\":\"metal breastplate\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":50},\"Weight\":5,\"Value\":2000,\"Components\":{\"armour\":{\"Bonus\":3}}}", Item{"metal breastplate", "bandit", icon.NewIcon(91, 50), 5, 2000, map[string]component{"armour": ArmourComponent{3}}}},
	{"{\"Name\":\"reinforced leather jacket\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":70},\"Weight\":3,\"Value\":1500,\"Components\":{\"armour\":{\"Bonus\":2}}}", Item{"reinforced leather jacket", "bandit", icon.NewIcon(91, 70), 3, 1500, map[string]component{"armour": ArmourComponent{2}}}},
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
		armour := Item{}

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

		if armour.Component("armour") != pair.armour.Component("armour") {
			t.Error(
				"For", "Armour",
				"expected", pair.armour.Component("armour"),
				"got", armour.Component("armour"),
			)
		}

		if armour.HasComponent("cover") != pair.armour.HasComponent("cover") {
			t.Error(
				"For", "Gives cover",
				"expected", pair.armour.HasComponent("cover"),
				"got", armour.HasComponent("cover"),
			)
		}

		if armour.HasComponent("corpse") != pair.armour.HasComponent("corpse") {
			t.Error(
				"For", "Corpse",
				"expected", pair.armour.HasComponent("corpse"),
				"got", armour.HasComponent("corpse"),
			)
		}

	}

}
