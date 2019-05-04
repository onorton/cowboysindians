package item

import (
	"encoding/json"
	"testing"

	"github.com/onorton/cowboysindians/icon"
)

type corpseMarshallingPair struct {
	corpse Item
	result string
}

var corpseMarshallingTests = []corpseMarshallingPair{
	{Item{"bandit's head", "some bandit", icon.NewIcon(37, 5), 10.5, 1000, nil, nil, &tag{}, nil, nil, nil, nil}, "{\"Name\":\"bandit's head\",\"Owner\":\"some bandit\",\"Icon\":{\"Icon\":37,\"Colour\":5},\"Weight\":10.5,\"Value\":1000,\"Cover\":null,\"Description\":null,\"Corpse\":{},\"AmmoType\":null,\"Armour\":null,\"Weapon\":null,\"Consumable\":null}"},
	{Item{"bandit's body", "another bandit", icon.NewIcon(37, 5), 140.0, 500, &tag{}, nil, &tag{}, nil, nil, nil, nil}, "{\"Name\":\"bandit's body\",\"Owner\":\"another bandit\",\"Icon\":{\"Icon\":37,\"Colour\":5},\"Weight\":140,\"Value\":500,\"Cover\":{},\"Description\":null,\"Corpse\":{},\"AmmoType\":null,\"Armour\":null,\"Weapon\":null,\"Consumable\":null}"},
}

type corpseUnmarshallingPair struct {
	corpseJson string
	corpse     Item
}

var corpseUnmarshallingTests = []corpseUnmarshallingPair{
	{"{\"Name\":\"bandit's head\",\"Owner\":\"some bandit\",\"Icon\":{\"Icon\":37,\"Colour\":5},\"Weight\":10.5,\"Value\":1000,\"Cover\":null,\"Corpse\":{},\"Armour\":null,\"Weapon\":null,\"Consumable\":null}", Item{"bandit's head", "some bandit", icon.NewIcon(37, 5), 10.5, 1000, nil, nil, &tag{}, nil, nil, nil, nil}},
	{"{\"Name\":\"bandit's body\",\"Owner\":\"another bandit\",\"Icon\":{\"Icon\":37,\"Colour\":5},\"Weight\":140,\"Value\":500,\"Cover\":{},\"Corpse\":{},\"Armour\":null,\"Weapon\":null,\"Consumable\":null}", Item{"bandit's body", "another bandit", icon.NewIcon(37, 5), 140.0, 500, &tag{}, nil, &tag{}, nil, nil, nil, nil}},
}

func TestCorpseMarshalling(t *testing.T) {

	for _, pair := range corpseMarshallingTests {

		result, err := json.Marshal(&(pair.corpse))
		if err != nil {
			t.Error("Failed when marshalling", pair.corpse, err)
		}
		if string(result) != pair.result {
			t.Error(
				"For", pair.corpse,
				"expected", pair.result,
				"got", string(result),
			)
		}
	}
}

func TestCorpseUnmarshalling(t *testing.T) {

	for _, pair := range corpseUnmarshallingTests {
		corpse := Item{}

		if err := json.Unmarshal([]byte(pair.corpseJson), &corpse); err != nil {
			t.Error("Failed when unmarshalling", pair.corpseJson, err)
		}

		if corpse.name != pair.corpse.name {
			t.Error(
				"For", "Name",
				"expected", pair.corpse.name,
				"got", corpse.name,
			)
		}

		if corpse.owner != pair.corpse.owner {
			t.Error(
				"For", "Owner",
				"expected", pair.corpse.owner,
				"got", corpse.owner,
			)
		}

		if corpse.ic != pair.corpse.ic {
			t.Error(
				"For", "Icon",
				"expected", pair.corpse.ic,
				"got", corpse.ic,
			)
		}

		if corpse.w != pair.corpse.w {
			t.Error(
				"For", "Weight",
				"expected", pair.corpse.w,
				"got", corpse.w,
			)
		}

		if corpse.v != pair.corpse.v {
			t.Error(
				"For", "Value",
				"expected", pair.corpse.v,
				"got", corpse.v,
			)
		}

		if (corpse.cover == nil && pair.corpse.cover != nil) || (corpse.cover != nil && pair.corpse.cover == nil) {
			t.Error(
				"For", "Gives cover",
				"expected", pair.corpse.cover,
				"got", corpse.cover,
			)
		}

		if corpse.cover != nil && pair.corpse.cover != nil && *(corpse.cover) != *(pair.corpse.cover) {
			t.Error(
				"For", "Gives cover",
				"expected", *(pair.corpse.cover),
				"got", *(corpse.cover),
			)
		}

		if (corpse.corpse == nil && pair.corpse.corpse != nil) || (corpse.corpse != nil && pair.corpse.corpse == nil) {
			t.Error(
				"For", "Corpse",
				"expected", pair.corpse.corpse,
				"got", corpse.corpse,
			)
		}

		if corpse.corpse != nil && pair.corpse.corpse != nil && *(corpse.corpse) != *(pair.corpse.corpse) {
			t.Error(
				"For", "Corpse",
				"expected", *(pair.corpse.corpse),
				"got", *(corpse.corpse),
			)
		}

	}

}
