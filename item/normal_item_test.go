package item

import (
	"encoding/json"
	"testing"

	"github.com/onorton/cowboysindians/icon"
)

type marshallingPair struct {
	item   NormalItem
	result string
}

var marshallingTests = []marshallingPair{
	{NormalItem{baseItem{"gem", "bandit", icon.NewIcon(42, 4), 2, 2000}, false, nil, false, NoAmmo, nil, nil}, "{\"Type\":\"normal\",\"Name\":\"gem\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":42,\"Colour\":4},\"Weight\":2,\"Value\":2000,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":null,\"Weapon\":null}"},
	{NormalItem{baseItem{"stick", "townsman", icon.NewIcon(30, 7), 5, 2}, false, nil, false, NoAmmo, nil, nil}, "{\"Type\":\"normal\",\"Name\":\"stick\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":30,\"Colour\":7},\"Weight\":5,\"Value\":2,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":null,\"Weapon\":null}"},
	{NormalItem{baseItem{"bowl", "bandit", icon.NewIcon(66, 10), 3, 10}, false, nil, false, NoAmmo, nil, nil}, "{\"Type\":\"normal\",\"Name\":\"bowl\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":66,\"Colour\":10},\"Weight\":3,\"Value\":10,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":null,\"Weapon\":null}"},
	{NormalItem{baseItem{"barrel", "townsman", icon.NewIcon(111, 0), 30, 200}, true, nil, false, NoAmmo, nil, nil}, "{\"Type\":\"normal\",\"Name\":\"barrel\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":111,\"Colour\":0},\"Weight\":30,\"Value\":200,\"Cover\":true,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":null,\"Weapon\":null}"},
}

type unmarshallingPair struct {
	itemJson string
	item     NormalItem
}

var unmarshallingTests = []unmarshallingPair{
	{"{\"Type\":\"normal\",\"Name\":\"gem\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":42,\"Colour\":4},\"Weight\":2,\"Value\":2000,\"Cover\":false}", NormalItem{baseItem{"gem", "bandit", icon.NewIcon(42, 4), 2, 2000}, false, nil, false, NoAmmo, nil, nil}},
	{"{\"Type\":\"normal\",\"Name\":\"stick\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":30,\"Colour\":7},\"Weight\":5,\"Value\":2,\"Cover\":false}", NormalItem{baseItem{"stick", "townsman", icon.NewIcon(30, 7), 5, 2}, false, nil, false, NoAmmo, nil, nil}},
	{"{\"Type\":\"normal\",\"Name\":\"bowl\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":66,\"Colour\":10},\"Weight\":3,\"Value\":10,\"Cover\":false}", NormalItem{baseItem{"bowl", "bandit", icon.NewIcon(66, 10), 3, 10}, false, nil, false, NoAmmo, nil, nil}},
	{"{\"Type\":\"normal\",\"Name\":\"barrel\",\"Owner\":\"townsman\",\"Icon\":{\"Icon\":111,\"Colour\":0},\"Weight\":30,\"Value\":200,\"Cover\":true}", NormalItem{baseItem{"barrel", "townsman", icon.NewIcon(111, 0), 30, 200}, true, nil, false, NoAmmo, nil, nil}},
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
		item := NormalItem{}

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

		if item.cover != pair.item.cover {
			t.Error(
				"For", "Gives cover",
				"expected", pair.item.cover,
				"got", item.cover,
			)
		}

		if item.corpse != pair.item.corpse {
			t.Error(
				"For", "Corpse",
				"expected", pair.item.corpse,
				"got", item.corpse,
			)
		}
	}

}
