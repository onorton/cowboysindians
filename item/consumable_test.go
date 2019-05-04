package item

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/onorton/cowboysindians/icon"
)

type consumableMarshallingPair struct {
	consumable Item
	result     string
}

var consumableMarshallingTests = []consumableMarshallingPair{
	{Item{"beer", "townsman", icon.NewIcon(98, 2), 0.01, 20, nil, nil, nil, nil, nil, nil, &consumableComponent{map[string][]Effect{"hp": []Effect{*NewInstantEffect(1)}, "thirst": []Effect{*NewInstantEffect(-10)}}}}, "{\"Name\":\"beer\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":98,\"Colour\":2},\"Weight\":0.01,\"Value\":20,\"Cover\":null,\"Description\":null,\"Corpse\":null,\"AmmoType\":null,\"Armour\":null,\"Weapon\":null,\"Consumable\":{\"Effects\":{\"hp\":[{\"Effect\":1,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}],\"thirst\":[{\"Effect\":-10,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}]}}}"},
	{Item{"standard ration", "bandit", icon.NewIcon(42, 4), 0.1, 40, nil, nil, nil, nil, nil, nil, &consumableComponent{map[string][]Effect{"hp": []Effect{*NewInstantEffect(10)}, "hunger": []Effect{*NewInstantEffect(-10)}}}}, "{\"Name\":\"standard ration\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":42,\"Colour\":4},\"Weight\":0.1,\"Value\":40,\"Cover\":null,\"Description\":null,\"Corpse\":null,\"AmmoType\":null,\"Armour\":null,\"Weapon\":null,\"Consumable\":{\"Effects\":{\"hp\":[{\"Effect\":10,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}],\"hunger\":[{\"Effect\":-10,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}]}}}"},
	{Item{"healing potion", "townsman", icon.NewIcon(112, 4), 0.1, 100, nil, nil, nil, nil, nil, nil, &consumableComponent{map[string][]Effect{"hp": []Effect{*NewInstantEffect(20)}}}}, "{\"Name\":\"healing potion\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":112,\"Colour\":4},\"Weight\":0.1,\"Value\":100,\"Cover\":null,\"Description\":null,\"Corpse\":null,\"AmmoType\":null,\"Armour\":null,\"Weapon\":null,\"Consumable\":{\"Effects\":{\"hp\":[{\"Effect\":20,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}]}}}"},
}

type consumableUnmarshallingPair struct {
	consumableJson string
	consumable     Item
}

var consumableUnmarshallingTests = []consumableUnmarshallingPair{
	{"{\"Name\":\"beer\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":98,\"Colour\":2},\"Weight\":0.01,\"Value\":20,\"Cover\":null,\"Description\":null,\"Corpse\":null,\"AmmoType\":null,\"Armour\":null,\"Weapon\":null,\"Consumable\":{\"Effects\":{\"hp\":[{\"Effect\":1,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}],\"thirst\":[{\"Effect\":-10,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}]}}}", Item{"beer", "townsman", icon.NewIcon(98, 2), 0.01, 20, nil, nil, nil, nil, nil, nil, &consumableComponent{map[string][]Effect{"hp": []Effect{*NewInstantEffect(1)}, "thirst": []Effect{*NewInstantEffect(-10)}}}}},
	{"{\"Name\":\"standard ration\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":42,\"Colour\":4},\"Weight\":0.1,\"Value\":40,\"Cover\":null,\"Description\":null,\"Corpse\":null,\"AmmoType\":null,\"Armour\":null,\"Weapon\":null,\"Consumable\":{\"Effects\":{\"hp\":[{\"Effect\":10,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}],\"hunger\":[{\"Effect\":-10,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}]}}}", Item{"standard ration", "bandit", icon.NewIcon(42, 4), 0.1, 40, nil, nil, nil, nil, nil, nil, &consumableComponent{map[string][]Effect{"hp": []Effect{*NewInstantEffect(10)}, "hunger": []Effect{*NewInstantEffect(-10)}}}}},
	{"{\"Name\":\"healing potion\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":112,\"Colour\":4},\"Weight\":0.1,\"Value\":100,\"Cover\":null,\"Description\":null,\"Corpse\":null,\"AmmoType\":null,\"Armour\":null,\"Weapon\":null,\"Consumable\":{\"Effects\":{\"hp\":[{\"Effect\":20,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}]}}}", Item{"healing potion", "townsman", icon.NewIcon(112, 4), 0.1, 100, nil, nil, nil, nil, nil, nil, &consumableComponent{map[string][]Effect{"hp": []Effect{*NewInstantEffect(20)}}}}},
}

func TestConsumableMarshalling(t *testing.T) {

	for _, pair := range consumableMarshallingTests {

		result, err := json.Marshal(&(pair.consumable))
		if err != nil {
			t.Error("Failed when marshalling", pair.consumable, err)
		}
		if string(result) != pair.result {
			t.Error(
				"For", pair.consumable,
				"expected", pair.result,
				"got", string(result),
			)
		}
	}
}

func TestConsumableUnmarshalling(t *testing.T) {

	for _, pair := range consumableUnmarshallingTests {
		consumable := Item{}

		if err := json.Unmarshal([]byte(pair.consumableJson), &consumable); err != nil {
			t.Error("Failed when unmarshalling", pair.consumableJson, err)
		}
		if consumable.name != pair.consumable.name {
			t.Error(
				"For", "Name",
				"expected", pair.consumable.name,
				"got", consumable.name,
			)
		}

		if consumable.owner != pair.consumable.owner {
			t.Error(
				"For", "Owner",
				"expected", pair.consumable.owner,
				"got", consumable.owner,
			)
		}

		if consumable.ic != pair.consumable.ic {
			t.Error(
				"For", "Icon",
				"expected", pair.consumable.ic,
				"got", consumable.ic,
			)
		}

		if consumable.w != pair.consumable.w {
			t.Error(
				"For", "Weight",
				"expected", pair.consumable.w,
				"got", consumable.w,
			)
		}

		if consumable.v != pair.consumable.v {
			t.Error(
				"For", "Value",
				"expected", pair.consumable.v,
				"got", consumable.v,
			)
		}

		if !reflect.DeepEqual(consumable.consumable.Effects, pair.consumable.consumable.Effects) {
			t.Error(
				"For", "Effects",
				"expected", pair.consumable.consumable.Effects,
				"got", consumable.consumable.Effects,
			)
		}
	}

}
