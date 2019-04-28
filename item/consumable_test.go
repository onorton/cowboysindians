package item

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/onorton/cowboysindians/icon"
)

type consumableMarshallingPair struct {
	consumable Consumable
	result     string
}

var consumableMarshallingTests = []consumableMarshallingPair{
	{Consumable{baseItem{"beer", "townsman", icon.NewIcon(98, 2), 0.01, 20}, map[string][]Effect{"hp": []Effect{*NewInstantEffect(1)}, "thirst": []Effect{*NewInstantEffect(-10)}}}, "{\"Type\":\"consumable\",\"Name\":\"beer\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":98,\"Colour\":2},\"Weight\":0.01,\"Value\":20,\"Effects\":{\"hp\":[{\"Effect\":1,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}],\"thirst\":[{\"Effect\":-10,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}]}}"},
	{Consumable{baseItem{"standard ration", "bandit", icon.NewIcon(42, 4), 0.1, 40}, map[string][]Effect{"hp": []Effect{*NewInstantEffect(10)}, "hunger": []Effect{*NewInstantEffect(-10)}}}, "{\"Type\":\"consumable\",\"Name\":\"standard ration\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":42,\"Colour\":4},\"Weight\":0.1,\"Value\":40,\"Effects\":{\"hp\":[{\"Effect\":10,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}],\"hunger\":[{\"Effect\":-10,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}]}}"},
	{Consumable{baseItem{"healing potion", "townsman", icon.NewIcon(112, 4), 0.1, 100}, map[string][]Effect{"hp": []Effect{*NewInstantEffect(20)}}}, "{\"Type\":\"consumable\",\"Name\":\"healing potion\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":112,\"Colour\":4},\"Weight\":0.1,\"Value\":100,\"Effects\":{\"hp\":[{\"Effect\":20,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}]}}"},
}

type consumableUnmarshallingPair struct {
	consumableJson string
	consumable     Consumable
}

var consumableUnmarshallingTests = []consumableUnmarshallingPair{
	{"{\"Type\":\"consumable\",\"Name\":\"beer\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":98,\"Colour\":2},\"Weight\":0.01,\"Value\":20,\"Effects\":{\"hp\":[{\"Effect\":1,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}],\"thirst\":[{\"Effect\":-10,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}]}}", Consumable{baseItem{"beer", "townsman", icon.NewIcon(98, 2), 0.01, 20}, map[string][]Effect{"hp": []Effect{*NewInstantEffect(1)}, "thirst": []Effect{*NewInstantEffect(-10)}}}},
	{"{\"Type\":\"consumable\",\"Name\":\"standard ration\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":42,\"Colour\":4},\"Weight\":0.1,\"Value\":40,\"Effects\":{\"hp\":[{\"Effect\":10,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}],\"hunger\":[{\"Effect\":-10,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}]}}", Consumable{baseItem{"standard ration", "bandit", icon.NewIcon(42, 4), 0.1, 40}, map[string][]Effect{"hp": []Effect{*NewInstantEffect(10)}, "hunger": []Effect{*NewInstantEffect(-10)}}}},
	{"{\"Type\":\"consumable\",\"Name\":\"healing potion\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":112,\"Colour\":4},\"Weight\":0.1,\"Value\":100,\"Effects\":{\"hp\":[{\"Effect\":20,\"OnMax\":false,\"Duration\":1,\"Activated\":false,\"Compounded\":false}]}}", Consumable{baseItem{"healing potion", "townsman", icon.NewIcon(112, 4), 0.1, 100}, map[string][]Effect{"hp": []Effect{*NewInstantEffect(20)}}}},
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
		consumable := Consumable{}

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

		if !reflect.DeepEqual(consumable.effects, pair.consumable.effects) {
			t.Error(
				"For", "Effects",
				"expected", pair.consumable.effects,
				"got", consumable.effects,
			)
		}
	}

}
