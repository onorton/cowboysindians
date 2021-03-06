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

var ammoMarshallingTests = []ammoMarshallingPair{
	{Item{"shotgun shell", "bandit", icon.NewIcon(44, 2), 0.2, 20, map[string]component{"ammo": AmmoComponent{Shotgun}}}, "{\"Name\":\"shotgun shell\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":44,\"Colour\":2},\"Weight\":0.2,\"Value\":20,\"Components\":{\"ammo\":{\"AmmoType\":2}}}"},
	{Item{"pistol bullet", "bandit", icon.NewIcon(44, 3), 0.01, 10, map[string]component{"ammo": AmmoComponent{Pistol}}}, "{\"Name\":\"pistol bullet\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":44,\"Colour\":3},\"Weight\":0.01,\"Value\":10,\"Components\":{\"ammo\":{\"AmmoType\":1}}}"},
}

type ammoUnmarshallingPair struct {
	ammoJson string
	ammo     Item
}

var ammoUnmarshallingTests = []ammoUnmarshallingPair{
	{"{\"Name\":\"shotgun shell\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":44,\"Colour\":2},\"Weight\":0.2,\"Value\":20,\"Components\":{\"ammo\":{\"AmmoType\":2}}}", Item{"shotgun shell", "bandit", icon.NewIcon(44, 2), 0.2, 20, map[string]component{"ammo": AmmoComponent{Shotgun}}}},
	{"{\"Name\":\"pistol bullet\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":44,\"Colour\":3},\"Weight\":0.01,\"Value\":10,\"Components\":{\"ammo\":{\"AmmoType\":1}}}", Item{"pistol bullet", "bandit", icon.NewIcon(44, 3), 0.01, 10, map[string]component{"ammo": AmmoComponent{Pistol}}}},
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

		if ammo.Component("ammo").(AmmoComponent).AmmoType != pair.ammo.Component("ammo").(AmmoComponent).AmmoType {
			t.Error(
				"For", "Ammo type",
				"expected", pair.ammo.Component("ammo").(AmmoComponent).AmmoType,
				"got", ammo.Component("ammo").(AmmoComponent).AmmoType,
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
