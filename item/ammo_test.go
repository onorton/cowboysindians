package item

import (
	"encoding/json"
	"testing"

	"github.com/onorton/cowboysindians/icon"
)

type ammoMarshallingPair struct {
	ammo   Ammo
	result string
}

var ammoMarshallingTests = []ammoMarshallingPair{
	{Ammo{baseItem{"shotgun shell", "bandit", icon.NewIcon(44, 2), 0.2, 20}, 2}, "{\"Name\":\"shotgun shell\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":44,\"Colour\":2},\"Type\":2,\"Weight\":0.2,\"Value\":20}"},
	{Ammo{baseItem{"pistol bullet", "bandit", icon.NewIcon(44, 3), 0.01, 10}, 1}, "{\"Name\":\"pistol bullet\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":44,\"Colour\":3},\"Type\":1,\"Weight\":0.01,\"Value\":10}"},
}

type ammoUnmarshallingPair struct {
	ammoJson string
	ammo     Ammo
}

var ammoUnmarshallingTests = []ammoUnmarshallingPair{
	{"{\"Name\":\"shotgun shell\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":44,\"Colour\":2},\"Type\":2,\"Weight\":0.2,\"Value\":20}", Ammo{baseItem{"shotgun shell", "bandit", icon.NewIcon(44, 2), 0.2, 20}, 2}},
	{"{\"Name\":\"pistol bullet\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":44,\"Colour\":3},\"Type\":1,\"Weight\":0.01,\"Value\":10}", Ammo{baseItem{"pistol bullet", "bandit", icon.NewIcon(44, 3), 0.01, 10}, 1}},
}

func TestAmmoMarshalling(t *testing.T) {

	for _, pair := range ammoMarshallingTests {

		result, err := json.Marshal(&(pair.ammo))
		if err != nil {
			t.Error("Failed when marshalling", pair.ammo, err)
		}
		if string(result) != pair.result {
			t.Error(
				"For", pair.ammo,
				"expected", pair.result,
				"got", string(result),
			)
		}
	}
}

func TestAmmoUnmarshalling(t *testing.T) {

	for _, pair := range ammoUnmarshallingTests {
		ammo := Ammo{}

		if err := json.Unmarshal([]byte(pair.ammoJson), &ammo); err != nil {
			t.Error("Failed when unmarshalling", pair.ammoJson, err)
		}
		if ammo.name != pair.ammo.name {
			t.Error(
				"For", "Name",
				"expected", pair.ammo.name,
				"got", ammo.name,
			)
		}

		if ammo.owner != pair.ammo.owner {
			t.Error(
				"For", "Owner",
				"expected", pair.ammo.owner,
				"got", ammo.owner,
			)
		}

		if ammo.ic != pair.ammo.ic {
			t.Error(
				"For", "Icon",
				"expected", pair.ammo.ic,
				"got", ammo.ic,
			)
		}

		if ammo.w != pair.ammo.w {
			t.Error(
				"For", "Weight",
				"expected", pair.ammo.w,
				"got", ammo.w,
			)
		}

		if ammo.t != pair.ammo.t {
			t.Error(
				"For", "Type",
				"expected", pair.ammo.t,
				"got", ammo.t,
			)
		}
		if ammo.v != pair.ammo.v {
			t.Error(
				"For", "Value",
				"expected", pair.ammo.v,
				"got", ammo.v,
			)
		}
	}

}
