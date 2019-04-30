package item

import (
	"encoding/json"
	"testing"

	"github.com/onorton/cowboysindians/icon"
)

type ammoMarshallingPair struct {
	ammo   NormalItem
	result string
}

var ammoMarshallingTests = []ammoMarshallingPair{
	{NormalItem{baseItem{"shotgun shell", "bandit", icon.NewIcon(44, 2), 0.2, 20}, false, nil, false, Shotgun}, "{\"Type\":\"normal\",\"Name\":\"shotgun shell\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":44,\"Colour\":2},\"Weight\":0.2,\"Value\":20,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":2}"},
	{NormalItem{baseItem{"pistol bullet", "bandit", icon.NewIcon(44, 3), 0.01, 10}, false, nil, false, Pistol}, "{\"Type\":\"normal\",\"Name\":\"pistol bullet\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":44,\"Colour\":3},\"Weight\":0.01,\"Value\":10,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":1}"},
}

type ammoUnmarshallingPair struct {
	ammoJson string
	ammo     NormalItem
}

var ammoUnmarshallingTests = []ammoUnmarshallingPair{
	{"{\"Type\":\"normal\",\"Name\":\"shotgun shell\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":44,\"Colour\":2},\"Weight\":0.2,\"Value\":20,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":2}", NormalItem{baseItem{"shotgun shell", "bandit", icon.NewIcon(44, 2), 0.2, 20}, false, nil, false, Shotgun}},
	{"{\"Type\":\"normal\",\"Name\":\"pistol bullet\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":44,\"Colour\":3},\"Weight\":0.01,\"Value\":10,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":1}", NormalItem{baseItem{"pistol bullet", "bandit", icon.NewIcon(44, 3), 0.01, 10}, false, nil, false, Pistol}},
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
		ammo := NormalItem{}

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

		if ammo.ammoType != pair.ammo.ammoType {
			t.Error(
				"For", "Ammo type",
				"expected", pair.ammo.ammoType,
				"got", ammo.ammoType,
			)
		}
		if ammo.v != pair.ammo.v {
			t.Error(
				"For", "Value",
				"expected", pair.ammo.v,
				"got", ammo.v,
			)
		}

		if ammo.cover != pair.ammo.cover {
			t.Error(
				"For", "Gives cover",
				"expected", pair.ammo.cover,
				"got", ammo.cover,
			)
		}

		if ammo.corpse != pair.ammo.corpse {
			t.Error(
				"For", "Corpse",
				"expected", pair.ammo.corpse,
				"got", ammo.corpse,
			)
		}
	}

}
