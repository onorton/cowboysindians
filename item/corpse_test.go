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
	{Item{"bandit's head", "some bandit", icon.NewIcon(37, 5), 10.5, 1000, map[string]component{"corpse": tag{}}}, "{\"Name\":\"bandit's head\",\"Owner\":\"some bandit\",\"Icon\":{\"Icon\":37,\"Colour\":5},\"Weight\":10.5,\"Value\":1000,\"Components\":{\"corpse\":{}}}"},
	{Item{"bandit's body", "another bandit", icon.NewIcon(37, 5), 140.0, 500, map[string]component{"corpse": tag{}, "cover": tag{}}}, "{\"Name\":\"bandit's body\",\"Owner\":\"another bandit\",\"Icon\":{\"Icon\":37,\"Colour\":5},\"Weight\":140,\"Value\":500,\"Components\":{\"corpse\":{},\"cover\":{}}}"},
}

type corpseUnmarshallingPair struct {
	corpseJson string
	corpse     Item
}

var corpseUnmarshallingTests = []corpseUnmarshallingPair{
	{"{\"Name\":\"bandit's head\",\"Owner\":\"some bandit\",\"Icon\":{\"Icon\":37,\"Colour\":5},\"Weight\":10.5,\"Value\":1000,\"Components\":{\"corpse\":{}}}", Item{"bandit's head", "some bandit", icon.NewIcon(37, 5), 10.5, 1000, map[string]component{"corpse": tag{}}}},
	{"{\"Name\":\"bandit's body\",\"Owner\":\"another bandit\",\"Icon\":{\"Icon\":37,\"Colour\":5},\"Weight\":140,\"Value\":500,\"Components\":{\"corpse\":{},\"cover\":{}}}", Item{"bandit's body", "another bandit", icon.NewIcon(37, 5), 140.0, 500, map[string]component{"corpse": tag{}, "cover": tag{}}}},
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

		if corpse.HasComponent("cover") != pair.corpse.HasComponent("cover") {
			t.Error(
				"For", "Gives cover",
				"expected", pair.corpse.HasComponent("cover"),
				"got", corpse.HasComponent("cover"),
			)
		}

		if corpse.HasComponent("corpse") != pair.corpse.HasComponent("corpse") {
			t.Error(
				"For", "Corpse",
				"expected", pair.corpse.HasComponent("corpse"),
				"got", corpse.HasComponent("corpse"),
			)
		}

	}

}
