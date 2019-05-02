package item

import (
	"encoding/json"
	"testing"

	"github.com/onorton/cowboysindians/icon"
)

type armourMarshallingPair struct {
	armour NormalItem
	result string
}

var armourMarshallingTests = []armourMarshallingPair{
	{NormalItem{baseItem{"leather jacket", "bandit", icon.NewIcon(91, 100), 2, 1000}, false, nil, false, NoAmmo, &armourComponent{1}, nil, nil}, "{\"Name\":\"leather jacket\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":100},\"Weight\":2,\"Value\":1000,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":{\"Bonus\":1},\"Weapon\":null,\"Consumable\":null}"},
	{NormalItem{baseItem{"metal breastplate", "bandit", icon.NewIcon(91, 50), 5, 2000}, false, nil, false, NoAmmo, &armourComponent{3}, nil, nil}, "{\"Name\":\"metal breastplate\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":50},\"Weight\":5,\"Value\":2000,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":{\"Bonus\":3},\"Weapon\":null,\"Consumable\":null}"},
	{NormalItem{baseItem{"reinforced leather jacket", "bandit", icon.NewIcon(91, 70), 3, 1500}, false, nil, false, NoAmmo, &armourComponent{2}, nil, nil}, "{\"Name\":\"reinforced leather jacket\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":70},\"Weight\":3,\"Value\":1500,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":{\"Bonus\":2},\"Weapon\":null,\"Consumable\":null}"},
}

type armourUnmarshallingPair struct {
	armourJson string
	armour     NormalItem
}

var armourUnmarshallingTests = []armourUnmarshallingPair{
	{"{\"Name\":\"leather jacket\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":100},\"Weight\":2,\"Value\":1000,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":{\"Bonus\":1},\"Weapon\":null,\"Consumable\":null}", NormalItem{baseItem{"leather jacket", "bandit", icon.NewIcon(91, 100), 2, 1000}, false, nil, false, NoAmmo, &armourComponent{1}, nil, nil}},
	{"{\"Name\":\"metal breastplate\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":50},\"Weight\":5,\"Value\":2000,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":{\"Bonus\":3},\"Weapon\":null,\"Consumable\":null}", NormalItem{baseItem{"metal breastplate", "bandit", icon.NewIcon(91, 50), 5, 2000}, false, nil, false, NoAmmo, &armourComponent{3}, nil, nil}},
	{"{\"Name\":\"reinforced leather jacket\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":91,\"Colour\":70},\"Weight\":3,\"Value\":1500,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":{\"Bonus\":2},\"Weapon\":null,\"Consumable\":null}", NormalItem{baseItem{"reinforced leather jacket", "bandit", icon.NewIcon(91, 70), 3, 1500}, false, nil, false, NoAmmo, &armourComponent{2}, nil, nil}},
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
		armour := NormalItem{}

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

		if armour.cover != pair.armour.cover {
			t.Error(
				"For", "Gives cover",
				"expected", pair.armour.cover,
				"got", armour.cover,
			)
		}

		if armour.corpse != pair.armour.corpse {
			t.Error(
				"For", "Corpse",
				"expected", pair.armour.corpse,
				"got", armour.corpse,
			)
		}
	}

}
