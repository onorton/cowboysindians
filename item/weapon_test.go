package item

import (
	"encoding/json"
	"testing"

	"github.com/onorton/cowboysindians/icon"
)

type weaponMarshallingPair struct {
	weapon Item
	result string
}

var weaponMarshallingTests = []weaponMarshallingPair{
	{Item{"shotgun", "bandit", icon.NewIcon(115, 3), 20, 5000, map[string]component{"weapon": WeaponComponent{4, Shotgun, &WeaponCapacity{2, 1}, Damage{4, 1, 0}}}}, "{\"Name\":\"shotgun\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":115,\"Colour\":3},\"Weight\":20,\"Value\":5000,\"Components\":{\"weapon\":{\"Range\":4,\"Type\":2,\"Capacity\":{\"Capacity\":2,\"Loaded\":1},\"Damage\":{\"Dice\":4,\"Number\":1,\"Bonus\":0}}}}"},
	{Item{"pistol", "bandit", icon.NewIcon(112, 1), 10, 6000, map[string]component{"weapon": WeaponComponent{10, Pistol, &WeaponCapacity{6, 6}, Damage{4, 1, -1}}}}, "{\"Name\":\"pistol\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":112,\"Colour\":1},\"Weight\":10,\"Value\":6000,\"Components\":{\"weapon\":{\"Range\":10,\"Type\":1,\"Capacity\":{\"Capacity\":6,\"Loaded\":6},\"Damage\":{\"Dice\":4,\"Number\":1,\"Bonus\":-1}}}}"},
	{Item{"sawn-off shotgun", "bandit", icon.NewIcon(115, 4), 15, 3000, map[string]component{"weapon": WeaponComponent{3, Shotgun, &WeaponCapacity{2, 0}, Damage{6, 1, 0}}}}, "{\"Name\":\"sawn-off shotgun\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":115,\"Colour\":4},\"Weight\":15,\"Value\":3000,\"Components\":{\"weapon\":{\"Range\":3,\"Type\":2,\"Capacity\":{\"Capacity\":2,\"Loaded\":0},\"Damage\":{\"Dice\":6,\"Number\":1,\"Bonus\":0}}}}"},
	{Item{"baseball bat", "bandit", icon.NewIcon(98, 8), 10, 200, map[string]component{"weapon": WeaponComponent{0, NoAmmo, nil, Damage{6, 1, 0}}}}, "{\"Name\":\"baseball bat\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":98,\"Colour\":8},\"Weight\":10,\"Value\":200,\"Components\":{\"weapon\":{\"Range\":0,\"Type\":0,\"Capacity\":null,\"Damage\":{\"Dice\":6,\"Number\":1,\"Bonus\":0}}}}"},
}

type weaponUnmarshallingPair struct {
	weaponJson string
	weapon     Item
}

var weaponUnmarshallingTests = []weaponUnmarshallingPair{
	{"{\"Name\":\"shotgun\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":115,\"Colour\":3},\"Weight\":20,\"Value\":5000,\"Components\":{\"weapon\":{\"Range\":4,\"Type\":2,\"Capacity\":{\"Capacity\":2,\"Loaded\":1},\"Damage\":{\"Dice\":4,\"Number\":1,\"Bonus\":0}}}}", Item{"shotgun", "bandit", icon.NewIcon(115, 3), 20, 5000, map[string]component{"weapon": WeaponComponent{4, Shotgun, &WeaponCapacity{2, 1}, Damage{4, 1, 0}}}}},
	{"{\"Name\":\"pistol\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":112,\"Colour\":1},\"Weight\":10,\"Value\":6000,\"Components\":{\"weapon\":{\"Range\":10,\"Type\":1,\"Capacity\":{\"Capacity\":6,\"Loaded\":6},\"Damage\":{\"Dice\":4,\"Number\":1,\"Bonus\":-1}}}}", Item{"pistol", "bandit", icon.NewIcon(112, 1), 10, 6000, map[string]component{"weapon": WeaponComponent{10, Pistol, &WeaponCapacity{6, 6}, Damage{4, 1, -1}}}}},
	{"{\"Name\":\"sawn-off shotgun\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":115,\"Colour\":4},\"Weight\":15,\"Value\":3000,\"Components\":{\"weapon\":{\"Range\":3,\"Type\":2,\"Capacity\":{\"Capacity\":2,\"Loaded\":0},\"Damage\":{\"Dice\":6,\"Number\":1,\"Bonus\":0}}}}", Item{"sawn-off shotgun", "bandit", icon.NewIcon(115, 4), 15, 3000, map[string]component{"weapon": WeaponComponent{3, Shotgun, &WeaponCapacity{2, 0}, Damage{6, 1, 0}}}}},
	{"{\"Name\":\"baseball bat\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":98,\"Colour\":8},\"Weight\":10,\"Value\":200,\"Components\":{\"weapon\":{\"Range\":0,\"Type\":0,\"Capacity\":null,\"Damage\":{\"Dice\":6,\"Number\":1,\"Bonus\":0}}}}", Item{"baseball bat", "bandit", icon.NewIcon(98, 8), 10, 200, map[string]component{"weapon": WeaponComponent{0, NoAmmo, nil, Damage{6, 1, 0}}}}},
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
		weapon := Item{}

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

		if weapon.owner != pair.weapon.owner {
			t.Error(
				"For", "Owner",
				"expected", pair.weapon.owner,
				"got", weapon.owner,
			)
		}

		if weapon.ic != pair.weapon.ic {
			t.Error(
				"For", "Icon",
				"expected", pair.weapon.ic,
				"got", weapon.ic,
			)
		}

		if weapon.components["weapon"].(WeaponComponent).Range != pair.weapon.components["weapon"].(WeaponComponent).Range {
			t.Error(
				"For", "Range",
				"expected", pair.weapon.components["weapon"].(WeaponComponent).Range,
				"got", weapon.components["weapon"].(WeaponComponent).Range,
			)
		}

		if weapon.components["weapon"].(WeaponComponent).Type != pair.weapon.components["weapon"].(WeaponComponent).Type {
			t.Error(
				"For", "Type",
				"expected", pair.weapon.components["weapon"].(WeaponComponent).Type,
				"got", weapon.components["weapon"].(WeaponComponent).Type,
			)
		}

		if weapon.w != pair.weapon.w {
			t.Error(
				"For", "Weight",
				"expected", pair.weapon.w,
				"got", weapon.w,
			)
		}

		if weapon.v != pair.weapon.v {
			t.Error(
				"For", "Value",
				"expected", pair.weapon.v,
				"got", weapon.v,
			)
		}

		if weapon.components["weapon"].(WeaponComponent).Capacity != nil && pair.weapon.components["weapon"].(WeaponComponent).Capacity != nil {
			if *weapon.components["weapon"].(WeaponComponent).Capacity != *(pair.weapon.components["weapon"].(WeaponComponent).Capacity) {
				t.Error(
					"For", "Weapon Capacity",
					"expected", *(pair.weapon.components["weapon"].(WeaponComponent).Capacity),
					"got", *(weapon.components["weapon"].(WeaponComponent).Capacity),
				)
			}
		}

		if (weapon.components["weapon"].(WeaponComponent).Capacity != nil && pair.weapon.components["weapon"].(WeaponComponent).Capacity == nil) || (weapon.components["weapon"].(WeaponComponent).Capacity == nil && pair.weapon.components["weapon"].(WeaponComponent).Capacity != nil) {
			t.Error(
				"For", "Weapon Capacity",
				"expected", pair.weapon.components["weapon"].(WeaponComponent).Capacity,
				"got", weapon.components["weapon"].(WeaponComponent).Capacity,
			)
		}

		if weapon.components["weapon"].(WeaponComponent).Damage != pair.weapon.components["weapon"].(WeaponComponent).Damage {
			t.Error(
				"For", "Damage",
				"expected", pair.weapon.components["weapon"].(WeaponComponent).Damage,
				"got", weapon.components["weapon"].(WeaponComponent).Damage,
			)
		}

	}

}
