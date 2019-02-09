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
	{Consumable{"beer", icon.NewIcon(98, 2), 0.01, map[string]int{"hp": 1, "thirst": 10}}, "{\"Name\":\"beer\",\"Icon\":{\"Icon\":98,\"Colour\":2},\"Weight\":0.01,\"Effects\":{\"hp\":1,\"thirst\":10}}"},
	{Consumable{"standard ration", icon.NewIcon(42, 4), 0.1, map[string]int{"hp": 10, "hunger": 10}}, "{\"Name\":\"standard ration\",\"Icon\":{\"Icon\":42,\"Colour\":4},\"Weight\":0.1,\"Effects\":{\"hp\":10,\"hunger\":10}}"},
	{Consumable{"healing potion", icon.NewIcon(112, 4), 0.1, map[string]int{"hp": 20}}, "{\"Name\":\"healing potion\",\"Icon\":{\"Icon\":112,\"Colour\":4},\"Weight\":0.1,\"Effects\":{\"hp\":20}}"},
}

type consumableUnmarshallingPair struct {
	consumableJson string
	consumable     Consumable
}

var consumableUnmarshallingTests = []consumableUnmarshallingPair{
	{"{\"Name\":\"beer\",\"Icon\":{\"Icon\":98,\"Colour\":2},\"Weight\":0.01,\"Effects\":{\"hp\":1,\"thirst\":10}}", Consumable{"beer", icon.NewIcon(98, 2), 0.01, map[string]int{"hp": 1, "thirst": 10}}},
	{"{\"Name\":\"standard ration\",\"Icon\":{\"Icon\":42,\"Colour\":4},\"Weight\":0.1,\"Effects\":{\"hp\":10,\"hunger\":10}}", Consumable{"standard ration", icon.NewIcon(42, 4), 0.1, map[string]int{"hp": 10, "hunger": 10}}},
	{"{\"Name\":\"healing potion\",\"Icon\":{\"Icon\":112,\"Colour\":4},\"Weight\":0.1,\"Effects\":{\"hp\":20}}", Consumable{"healing potion", icon.NewIcon(112, 4), 0.1, map[string]int{"hp": 20}}},
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

		if !reflect.DeepEqual(consumable.effects, pair.consumable.effects) {
			t.Error(
				"For", "Effects",
				"expected", pair.consumable.effects,
				"got", consumable.effects,
			)
		}
	}

}
