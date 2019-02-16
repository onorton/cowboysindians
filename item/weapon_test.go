package item

import (
	"encoding/json"
	"testing"

	"github.com/onorton/cowboysindians/icon"
)

type weaponMarshallingPair struct {
	weapon Weapon
	result string
}

var weaponMarshallingTests = []weaponMarshallingPair{
	{Weapon{"shotgun", icon.NewIcon(115, 3), 4, 1, 20, &WeaponCapacity{2, 1}, &Damage{4, 1, 0}}, "{\"Name\":\"shotgun\",\"Icon\":{\"Icon\":115,\"Colour\":3},\"Range\":4,\"Type\":1,\"Weight\":20,\"WeaponCapacity\":{\"Capacity\":2,\"Loaded\":1},\"Damage\":{\"Dice\":4,\"Number\":1,\"Bonus\":0}}"},
	{Weapon{"pistol", icon.NewIcon(112, 1), 10, 0, 10, &WeaponCapacity{6, 6}, &Damage{4, 1, -1}}, "{\"Name\":\"pistol\",\"Icon\":{\"Icon\":112,\"Colour\":1},\"Range\":10,\"Type\":0,\"Weight\":10,\"WeaponCapacity\":{\"Capacity\":6,\"Loaded\":6},\"Damage\":{\"Dice\":4,\"Number\":1,\"Bonus\":-1}}"},
	{Weapon{"sawn-off shotgun", icon.NewIcon(115, 4), 3, 1, 15, &WeaponCapacity{2, 0}, &Damage{6, 1, 0}}, "{\"Name\":\"sawn-off shotgun\",\"Icon\":{\"Icon\":115,\"Colour\":4},\"Range\":3,\"Type\":1,\"Weight\":15,\"WeaponCapacity\":{\"Capacity\":2,\"Loaded\":0},\"Damage\":{\"Dice\":6,\"Number\":1,\"Bonus\":0}}"},
	{Weapon{"baseball bat", icon.NewIcon(98, 8), 0, 0, 10, nil, &Damage{6, 1, 0}}, "{\"Name\":\"baseball bat\",\"Icon\":{\"Icon\":98,\"Colour\":8},\"Range\":0,\"Type\":0,\"Weight\":10,\"Damage\":{\"Dice\":6,\"Number\":1,\"Bonus\":0}}"},
}

type weaponUnmarshallingPair struct {
	weaponJson string
	weapon     Weapon
}

var weaponUnmarshallingTests = []weaponUnmarshallingPair{
	{"{\"Name\":\"shotgun\",\"Icon\":{\"Icon\":115,\"Colour\":3},\"Range\":4,\"Type\":1,\"Weight\":20,\"WeaponCapacity\":{\"Capacity\":2,\"Loaded\":1},\"Damage\":{\"Dice\":4,\"Number\":1,\"Bonus\":0}}", Weapon{"shotgun", icon.NewIcon(115, 3), 4, 1, 20, &WeaponCapacity{2, 1}, &Damage{4, 1, 0}}},
	{"{\"Name\":\"pistol\",\"Icon\":{\"Icon\":112,\"Colour\":1},\"Range\":10,\"Type\":0,\"Weight\":10,\"WeaponCapacity\":{\"Capacity\":6,\"Loaded\":6},\"Damage\":{\"Dice\":4,\"Number\":1,\"Bonus\":-1}}", Weapon{"pistol", icon.NewIcon(112, 1), 10, 0, 10, &WeaponCapacity{6, 6}, &Damage{4, 1, -1}}},
	{"{\"Name\":\"sawn-off shotgun\",\"Icon\":{\"Icon\":115,\"Colour\":4},\"Range\":3,\"Type\":1,\"Weight\":15,\"WeaponCapacity\":{\"Capacity\":2,\"Loaded\":0},\"Damage\":{\"Dice\":6,\"Number\":1,\"Bonus\":0}}", Weapon{"sawn-off shotgun", icon.NewIcon(115, 4), 3, 1, 15, &WeaponCapacity{2, 0}, &Damage{6, 1, 0}}},
	{"{\"Name\":\"baseball bat\",\"Icon\":{\"Icon\":98,\"Colour\":8},\"Range\":0,\"Type\":0,\"Weight\":10,\"Damage\":{\"Dice\":6,\"Number\":1,\"Bonus\":0}}", Weapon{"baseball bat", icon.NewIcon(98, 8), 0, 0, 10, nil, &Damage{6, 1, 0}}},
}

func TestWeaponMarshalling(t *testing.T) {

	for _, pair := range weaponMarshallingTests {

		result, err := json.Marshal(&(pair.weapon))
		if err != nil {
			t.Error("Failed when marshalling", pair.weapon, err)
		}
		if string(result) != pair.result {
			t.Error(
				"For", pair.weapon,
				"expected", pair.result,
				"got", string(result),
			)
		}
	}
}

func TestWeaponUnmarshalling(t *testing.T) {

	for _, pair := range weaponUnmarshallingTests {
		weapon := Weapon{}

		if err := json.Unmarshal([]byte(pair.weaponJson), &weapon); err != nil {
			t.Error("Failed when unmarshalling", pair.weaponJson, err)
		}

		if weapon.name != pair.weapon.name {
			t.Error(
				"For", "Name",
				"expected", pair.weapon.name,
				"got", weapon.name,
			)
		}

		if weapon.ic != pair.weapon.ic {
			t.Error(
				"For", "Icon",
				"expected", pair.weapon.ic,
				"got", weapon.ic,
			)
		}

		if weapon.r != pair.weapon.r {
			t.Error(
				"For", "Range",
				"expected", pair.weapon.r,
				"got", weapon.r,
			)
		}

		if weapon.t != pair.weapon.t {
			t.Error(
				"For", "Type",
				"expected", pair.weapon.t,
				"got", weapon.t,
			)
		}

		if weapon.w != pair.weapon.w {
			t.Error(
				"For", "Weight",
				"expected", pair.weapon.w,
				"got", weapon.w,
			)
		}

		if weapon.wc != nil && pair.weapon.wc != nil {
			if *weapon.wc != *(pair.weapon.wc) {
				t.Error(
					"For", "Weapon Capacity",
					"expected", *(pair.weapon.wc),
					"got", *(weapon.wc),
				)
			}
		}

		if (weapon.wc != nil && pair.weapon.wc == nil) || (weapon.wc == nil && pair.weapon.wc != nil) {
			t.Error(
				"For", "Weapon Capacity",
				"expected", pair.weapon.wc,
				"got", weapon.wc,
			)
		}

		if *(weapon.damage) != *(pair.weapon.damage) {
			t.Error(
				"For", "Damage",
				"expected", *(pair.weapon.damage),
				"got", *(weapon.damage),
			)
		}

	}

}
