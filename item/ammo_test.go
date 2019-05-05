package item

import (
	"encoding/json"
	"testing"

	"github.com/onorton/cowboysindians/icon"
)

type ammoMarshallingPair struct {
	ammo   Item
	result string
}

var shotgun WeaponType = Shotgun
var pistol WeaponType = Pistol

var ammoMarshallingTests = []ammoMarshallingPair{
	{Item{"shotgun shell", "bandit", icon.NewIcon(44, 2), 0.2, 20, map[string]component{}, nil, &shotgun, nil, nil}, "{\"Name\":\"shotgun shell\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":44,\"Colour\":2},\"Weight\":0.2,\"Value\":20,\"Components\":{},\"Description\":null,\"AmmoType\":2,\"Armour\":null,\"Weapon\":null}"},
	{Item{"pistol bullet", "bandit", icon.NewIcon(44, 3), 0.01, 10, map[string]component{}, nil, &pistol, nil, nil}, "{\"Name\":\"pistol bullet\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":44,\"Colour\":3},\"Weight\":0.01,\"Value\":10,\"Components\":{},\"Description\":null,\"AmmoType\":1,\"Armour\":null,\"Weapon\":null}"},
}

type ammoUnmarshallingPair struct {
	ammoJson string
	ammo     Item
}

var ammoUnmarshallingTests = []ammoUnmarshallingPair{
	{"{\"Name\":\"shotgun shell\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":44,\"Colour\":2},\"Weight\":0.2,\"Value\":20,\"Description\":null,\"Components\":{},\"AmmoType\":2,\"Armour\":null,\"Weapon\":null}", Item{"shotgun shell", "bandit", icon.NewIcon(44, 2), 0.2, 20, map[string]component{}, nil, &shotgun, nil, nil}},
	{"{\"Name\":\"pistol bullet\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":44,\"Colour\":3},\"Weight\":0.01,\"Value\":10,\"Description\":null,\"Components\":{},\"AmmoType\":1,\"Armour\":null,\"Weapon\":null}", Item{"pistol bullet", "bandit", icon.NewIcon(44, 3), 0.01, 10, map[string]component{}, nil, &pistol, nil, nil}},
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
		ammo := Item{}

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

		if *(ammo.ammoType) != *(pair.ammo.ammoType) {
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

		if ammo.HasComponent("cover") != pair.ammo.HasComponent("cover") {
			t.Error(
				"For", "Gives cover",
				"expected", pair.ammo.HasComponent("cover"),
				"got", ammo.HasComponent("cover"),
			)
		}

		if ammo.HasComponent("corpse") != pair.ammo.HasComponent("corpse") {
			t.Error(
				"For", "Corpse",
				"expected", pair.ammo.HasComponent("corpse"),
				"got", ammo.HasComponent("corpse"),
			)
		}

	}

}
