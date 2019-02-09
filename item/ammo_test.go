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
	{Ammo{"shotgun shell", icon.NewIcon(44, 2), 1, 0.2}, "{\"Name\":\"shotgun shell\",\"Icon\":{\"Icon\":44,\"Colour\":2},\"Type\":1,\"Weight\":0.2}"},
	{Ammo{"pistol bullet", icon.NewIcon(44, 3), 0, 0.01}, "{\"Name\":\"pistol bullet\",\"Icon\":{\"Icon\":44,\"Colour\":3},\"Type\":0,\"Weight\":0.01}"},
}

type ammoUnmarshallingPair struct {
	ammoJson string
	ammo     Ammo
}

var ammoUnmarshallingTests = []ammoUnmarshallingPair{
	{"{\"Name\":\"shotgun shell\",\"Icon\":{\"Icon\":44,\"Colour\":2},\"Type\":1,\"Weight\":0.2}", Ammo{"shotgun shell", icon.NewIcon(44, 2), 1, 0.2}},
	{"{\"Name\":\"pistol bullet\",\"Icon\":{\"Icon\":44,\"Colour\":3},\"Type\":0,\"Weight\":0.01}", Ammo{"pistol bullet", icon.NewIcon(44, 3), 0, 0.01}},
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
	}

}
