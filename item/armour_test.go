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
	{Item{"leather jacket", "bandit", icon.NewIcon(91, 100), 2, 1000, map[string]component{}, nil, nil, &armourComponent{1}, nil}, "{\"Name\":\"leather jacket\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":100},\"Weight\":2,\"Value\":1000,\"Components\":{},\"Description\":null,\"AmmoType\":null,\"Armour\":{\"Bonus\":1},\"Weapon\":null}"},
	{Item{"metal breastplate", "bandit", icon.NewIcon(91, 50), 5, 2000, map[string]component{}, nil, nil, &armourComponent{3}, nil}, "{\"Name\":\"metal breastplate\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":50},\"Weight\":5,\"Value\":2000,\"Components\":{},\"Description\":null,\"AmmoType\":null,\"Armour\":{\"Bonus\":3},\"Weapon\":null}"},
	{Item{"reinforced leather jacket", "bandit", icon.NewIcon(91, 70), 3, 1500, map[string]component{}, nil, nil, &armourComponent{2}, nil}, "{\"Name\":\"reinforced leather jacket\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":70},\"Weight\":3,\"Value\":1500,\"Components\":{},\"Description\":null,\"AmmoType\":null,\"Armour\":{\"Bonus\":2},\"Weapon\":null}"},
}

type armourUnmarshallingPair struct {
	armourJson string
	armour     Item
}

var armourUnmarshallingTests = []armourUnmarshallingPair{
	{"{\"Name\":\"leather jacket\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":100},\"Weight\":2,\"Value\":1000,\"Components\":{},\"Description\":null,\"AmmoType\":null,\"Armour\":{\"Bonus\":1},\"Weapon\":null}", Item{"leather jacket", "bandit", icon.NewIcon(91, 100), 2, 1000, map[string]component{}, nil, nil, &armourComponent{1}, nil}},
	{"{\"Name\":\"metal breastplate\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":50},\"Weight\":5,\"Value\":2000,\"Components\":{},\"Description\":null,\"AmmoType\":null,\"Armour\":{\"Bonus\":3},\"Weapon\":null}", Item{"metal breastplate", "bandit", icon.NewIcon(91, 50), 5, 2000, map[string]component{}, nil, nil, &armourComponent{3}, nil}},
	{"{\"Name\":\"reinforced leather jacket\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":70},\"Weight\":3,\"Value\":1500,\"Components\":{},\"Description\":null,\"AmmoType\":null,\"Armour\":{\"Bonus\":2},\"Weapon\":null}", Item{"reinforced leather jacket", "bandit", icon.NewIcon(91, 70), 3, 1500, map[string]component{}, nil, nil, &armourComponent{2}, nil}},
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

		if *(armour.armour) != *(pair.armour.armour) {
			t.Error(
				"For", "Armour",
				"expected", *(pair.armour.armour),
				"got", *(armour.armour),
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
