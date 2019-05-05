package item

import (
	"encoding/json"
	"testing"

	"github.com/onorton/cowboysindians/icon"
)

type marshallingPair struct {
	item   Item
	result string
}

var marshallingTests = []marshallingPair{
	{Item{"gem", "bandit", icon.NewIcon(42, 4), 2, 2000, map[string]tag{}, nil, nil, nil, nil, nil}, "{\"Name\":\"gem\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":42,\"Colour\":4},\"Weight\":2,\"Value\":2000,\"Components\":{},\"Description\":null,\"AmmoType\":null,\"Armour\":null,\"Weapon\":null,\"Consumable\":null}"},
	{Item{"stick", "townsman", icon.NewIcon(30, 7), 5, 2, map[string]tag{}, nil, nil, nil, nil, nil}, "{\"Name\":\"stick\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":30,\"Colour\":7},\"Weight\":5,\"Value\":2,\"Components\":{},\"Description\":null,\"AmmoType\":null,\"Armour\":null,\"Weapon\":null,\"Consumable\":null}"},
	{Item{"bowl", "bandit", icon.NewIcon(66, 10), 3, 10, map[string]tag{}, nil, nil, nil, nil, nil}, "{\"Name\":\"bowl\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":66,\"Colour\":10},\"Weight\":3,\"Value\":10,\"Components\":{},\"Description\":null,\"AmmoType\":null,\"Armour\":null,\"Weapon\":null,\"Consumable\":null}"},
	{Item{"barrel", "townsman", icon.NewIcon(111, 0), 30, 200, map[string]tag{"cover": tag{}}, nil, nil, nil, nil, nil}, "{\"Name\":\"barrel\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":111,\"Colour\":0},\"Weight\":30,\"Value\":200,\"Components\":{\"cover\":{}},\"Description\":null,\"AmmoType\":null,\"Armour\":null,\"Weapon\":null,\"Consumable\":null}"},
}

type unmarshallingPair struct {
	itemJson string
	item     Item
}

var unmarshallingTests = []unmarshallingPair{
	{"{\"Name\":\"gem\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":42,\"Colour\":4},\"Weight\":2,\"Value\":2000,\"Components\":{},\"Consumable\":null}", Item{"gem", "bandit", icon.NewIcon(42, 4), 2, 2000, map[string]tag{}, nil, nil, nil, nil, nil}},
	{"{\"Name\":\"stick\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":30,\"Colour\":7},\"Weight\":5,\"Value\":2,\"Components\":{},\"Consumable\":null}", Item{"stick", "townsman", icon.NewIcon(30, 7), 5, 2, map[string]tag{}, nil, nil, nil, nil, nil}},
	{"{\"Name\":\"bowl\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":66,\"Colour\":10},\"Weight\":3,\"Value\":10,\"Components\":{},\"Consumable\":null}", Item{"bowl", "bandit", icon.NewIcon(66, 10), 3, 10, map[string]tag{}, nil, nil, nil, nil, nil}},
	{"{\"Name\":\"barrel\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":111,\"Colour\":0},\"Weight\":30,\"Value\":200,\"Components\":{\"cover\":{}},\"Consumable\":null}", Item{"barrel", "townsman", icon.NewIcon(111, 0), 30, 200, map[string]tag{"cover": tag{}}, nil, nil, nil, nil, nil}},
}

func TestMarshalling(t *testing.T) {

	for _, pair := range marshallingTests {

		result, err := json.Marshal(&(pair.item))
		if err != nil {
			t.Error("Failed when marshalling", pair.item, err)
		}
		if string(result) != pair.result {
			t.Error(
				"For", pair.item,
				"expected", pair.result,
				"got", string(result),
			)
		}
	}
}

func TestUnmarshalling(t *testing.T) {

	for _, pair := range unmarshallingTests {
		item := Item{}

		if err := json.Unmarshal([]byte(pair.itemJson), &item); err != nil {
			t.Error("Failed when unmarshalling", pair.itemJson, err)
		}
		if item.name != pair.item.name {
			t.Error(
				"For", "Name",
				"expected", pair.item.name,
				"got", item.name,
			)
		}

		if item.owner != pair.item.owner {
			t.Error(
				"For", "Owner",
				"expected", pair.item.owner,
				"got", item.owner,
			)
		}

		if item.ic != pair.item.ic {
			t.Error(
				"For", "Icon",
				"expected", pair.item.ic,
				"got", item.ic,
			)
		}

		if item.w != pair.item.w {
			t.Error(
				"For", "Weight",
				"expected", pair.item.w,
				"got", item.w,
			)
		}

		if item.v != pair.item.v {
			t.Error(
				"For", "Value",
				"expected", pair.item.v,
				"got", item.v,
			)
		}

		if item.HasComponent("cover") != pair.item.HasComponent("cover") {
			t.Error(
				"For", "Gives cover",
				"expected", pair.item.HasComponent("cover"),
				"got", item.HasComponent("cover"),
			)
		}

		if item.HasComponent("corpse") != pair.item.HasComponent("corpse") {
			t.Error(
				"For", "Corpse",
				"expected", pair.item.HasComponent("corpse"),
				"got", item.HasComponent("corpse"),
			)
		}
	}

}
