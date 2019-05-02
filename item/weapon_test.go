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
	{Item{"shotgun", "bandit", icon.NewIcon(115, 3), 20, 5000, false, nil, false, NoAmmo, nil, &weaponComponent{4, Shotgun, &WeaponCapacity{2, 1}, Damage{4, 1, 0}}, nil}, "{\"Name\":\"shotgun\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":115,\"Colour\":3},\"Weight\":20,\"Value\":5000,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":null,\"Weapon\":{\"Range\":4,\"Type\":2,\"Capacity\":{\"Capacity\":2,\"Loaded\":1},\"Damage\":{\"Dice\":4,\"Number\":1,\"Bonus\":0}},\"Consumable\":null}"},
	{Item{"pistol", "bandit", icon.NewIcon(112, 1), 10, 6000, false, nil, false, NoAmmo, nil, &weaponComponent{10, Pistol, &WeaponCapacity{6, 6}, Damage{4, 1, -1}}, nil}, "{\"Name\":\"pistol\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":112,\"Colour\":1},\"Weight\":10,\"Value\":6000,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":null,\"Weapon\":{\"Range\":10,\"Type\":1,\"Capacity\":{\"Capacity\":6,\"Loaded\":6},\"Damage\":{\"Dice\":4,\"Number\":1,\"Bonus\":-1}},\"Consumable\":null}"},
	{Item{"sawn-off shotgun", "bandit", icon.NewIcon(115, 4), 15, 3000, false, nil, false, NoAmmo, nil, &weaponComponent{3, Shotgun, &WeaponCapacity{2, 0}, Damage{6, 1, 0}}, nil}, "{\"Name\":\"sawn-off shotgun\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":115,\"Colour\":4},\"Weight\":15,\"Value\":3000,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":null,\"Weapon\":{\"Range\":3,\"Type\":2,\"Capacity\":{\"Capacity\":2,\"Loaded\":0},\"Damage\":{\"Dice\":6,\"Number\":1,\"Bonus\":0}},\"Consumable\":null}"},
	{Item{"baseball bat", "bandit", icon.NewIcon(98, 8), 10, 200, false, nil, false, NoAmmo, nil, &weaponComponent{0, NoAmmo, nil, Damage{6, 1, 0}}, nil}, "{\"Name\":\"baseball bat\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":98,\"Colour\":8},\"Weight\":10,\"Value\":200,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":null,\"Weapon\":{\"Range\":0,\"Type\":0,\"Capacity\":null,\"Damage\":{\"Dice\":6,\"Number\":1,\"Bonus\":0}},\"Consumable\":null}"},
}

type weaponUnmarshallingPair struct {
	weaponJson string
	weapon     Item
}

var weaponUnmarshallingTests = []weaponUnmarshallingPair{
	{"{\"Name\":\"shotgun\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":115,\"Colour\":3},\"Weight\":20,\"Value\":5000,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":null,\"Weapon\":{\"Range\":4,\"Type\":2,\"Capacity\":{\"Capacity\":2,\"Loaded\":1},\"Damage\":{\"Dice\":4,\"Number\":1,\"Bonus\":0}},\"Consumable\":null}", Item{"shotgun", "bandit", icon.NewIcon(115, 3), 20, 5000, false, nil, false, NoAmmo, nil, &weaponComponent{4, Shotgun, &WeaponCapacity{2, 1}, Damage{4, 1, 0}}, nil}},
	{"{\"Name\":\"pistol\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":112,\"Colour\":1},\"Weight\":10,\"Value\":6000,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":null,\"Weapon\":{\"Range\":10,\"Type\":1,\"Capacity\":{\"Capacity\":6,\"Loaded\":6},\"Damage\":{\"Dice\":4,\"Number\":1,\"Bonus\":-1}},\"Consumable\":null}", Item{"pistol", "bandit", icon.NewIcon(112, 1), 10, 6000, false, nil, false, NoAmmo, nil, &weaponComponent{10, Pistol, &WeaponCapacity{6, 6}, Damage{4, 1, -1}}, nil}},
	{"{\"Name\":\"sawn-off shotgun\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":115,\"Colour\":4},\"Weight\":15,\"Value\":3000,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":null,\"Weapon\":{\"Range\":3,\"Type\":2,\"Capacity\":{\"Capacity\":2,\"Loaded\":0},\"Damage\":{\"Dice\":6,\"Number\":1,\"Bonus\":0}},\"Consumable\":null}", Item{"sawn-off shotgun", "bandit", icon.NewIcon(115, 4), 15, 3000, false, nil, false, NoAmmo, nil, &weaponComponent{3, Shotgun, &WeaponCapacity{2, 0}, Damage{6, 1, 0}}, nil}},
	{"{\"Name\":\"baseball bat\",\"Owner\":\"bandit\",\"Icon\":{\"Icon\":98,\"Colour\":8},\"Weight\":10,\"Value\":200,\"Cover\":false,\"Description\":null,\"Corpse\":false,\"AmmoType\":0,\"Armour\":null,\"Weapon\":{\"Range\":0,\"Type\":0,\"Capacity\":null,\"Damage\":{\"Dice\":6,\"Number\":1,\"Bonus\":0}},\"Consumable\":null}", Item{"baseball bat", "bandit", icon.NewIcon(98, 8), 10, 200, false, nil, false, NoAmmo, nil, &weaponComponent{0, NoAmmo, nil, Damage{6, 1, 0}}, nil}},
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

		if weapon.weapon.Range != pair.weapon.weapon.Range {
			t.Error(
				"For", "Range",
				"expected", pair.weapon.weapon.Range,
				"got", weapon.weapon.Range,
			)
		}

		if weapon.weapon.Type != pair.weapon.weapon.Type {
			t.Error(
				"For", "Type",
				"expected", pair.weapon.weapon.Type,
				"got", weapon.weapon.Type,
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

		if weapon.weapon.Capacity != nil && pair.weapon.weapon.Capacity != nil {
			if *weapon.weapon.Capacity != *(pair.weapon.weapon.Capacity) {
				t.Error(
					"For", "Weapon Capacity",
					"expected", *(pair.weapon.weapon.Capacity),
					"got", *(weapon.weapon.Capacity),
				)
			}
		}

		if (weapon.weapon.Capacity != nil && pair.weapon.weapon.Capacity == nil) || (weapon.weapon.Capacity == nil && pair.weapon.weapon.Capacity != nil) {
			t.Error(
				"For", "Weapon Capacity",
				"expected", pair.weapon.weapon.Capacity,
				"got", weapon.weapon.Capacity,
			)
		}

		if weapon.weapon.Damage != pair.weapon.weapon.Damage {
			t.Error(
				"For", "Damage",
				"expected", pair.weapon.weapon.Damage,
				"got", weapon.weapon.Damage,
			)
		}

	}

}
