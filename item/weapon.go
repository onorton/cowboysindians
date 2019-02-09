package item

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"math/rand"

	"github.com/onorton/cowboysindians/icon"
)

type WeaponAttributes struct {
	Icon   icon.Icon
	Damage DamageAttributes
	Range  int
	Type   WeaponType
	Weight float64
}

type DamageAttributes struct {
	Dice   int
	Number int
	Bonus  int
}

type WeaponType int

const (
	Pistol WeaponType = iota
	Shotgun
)

var weaponData map[string]WeaponAttributes

func fetchWeaponData() {
	data, err := ioutil.ReadFile("data/weapon.json")
	check(err)
	var wD map[string]WeaponAttributes
	err = json.Unmarshal(data, &wD)
	check(err)
	weaponData = wD
}

type Weapon struct {
	name   string
	ic     icon.Icon
	r      int
	t      WeaponType
	w      float64
	damage *Damage
}

type Damage struct {
	dice   int
	number int
	bonus  int
}

func NewWeapon(name string) *Weapon {
	weapon := weaponData[name]
	return &Weapon{name, weapon.Icon, weapon.Range, weapon.Type, weapon.Weight, &Damage{weapon.Damage.Dice, weapon.Damage.Number, weapon.Damage.Bonus}}
}

func (damage *Damage) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	diceValue, err := json.Marshal(damage.dice)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Dice\":%s,", diceValue))

	numberValue, err := json.Marshal(damage.number)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Number\":%s,", numberValue))

	bonusValue, err := json.Marshal(damage.bonus)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Bonus\":%s", bonusValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (weapon *Weapon) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	nameValue, err := json.Marshal(weapon.name)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Name\":%s,", nameValue))

	iconValue, err := json.Marshal(weapon.ic)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Icon\":%s,", iconValue))

	rangeValue, err := json.Marshal(weapon.r)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Range\":%s,", rangeValue))

	typeValue, err := json.Marshal(weapon.t)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Type\":%s,", typeValue))

	weightValue, err := json.Marshal(weapon.w)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Weight\":%s,", weightValue))

	damageValue, err := json.Marshal(weapon.damage)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Damage\":%s", damageValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (damage *Damage) UnmarshalJSON(data []byte) error {

	type damageJson struct {
		Dice   int
		Number int
		Bonus  int
	}
	var v damageJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	damage.dice = v.Dice
	damage.number = v.Number
	damage.bonus = v.Bonus

	return nil
}

func (weapon *Weapon) UnmarshalJSON(data []byte) error {

	type weaponJson struct {
		Name   string
		Icon   icon.Icon
		Range  int
		Type   WeaponType
		Weight float64
		Damage *Damage
	}
	var v weaponJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	weapon.name = v.Name
	weapon.ic = v.Icon
	weapon.r = v.Range
	weapon.t = v.Type
	weapon.w = v.Weight
	weapon.damage = v.Damage

	return nil
}

func (weapon *Weapon) GetName() string {
	return weapon.name
}
func (weapon *Weapon) Render(x, y int) {

	weapon.ic.Render(x, y)
}

func (weapon *Weapon) GetRange() int {
	return weapon.r
}

// Maximum possible damage
func (weapon *Weapon) GetMaxDamage() int {
	return weapon.damage.number*weapon.damage.dice + weapon.damage.bonus
}
func (weapon *Weapon) GetDamage() int {
	result := 0
	for i := 0; i < weapon.damage.number; i++ {
		result += rand.Intn(weapon.damage.dice) + 1
	}

	result += weapon.damage.bonus
	return result
}

func (weapon *Weapon) AmmoTypeMatches(ammo *Ammo) bool {
	return weapon.t == ammo.t
}

func (weapon *Weapon) GetKey() rune {
	h := fnv.New32()
	h.Write([]byte(weapon.name))
	key := rune(33 + h.Sum32()%93)
	if key == '*' {
		key++
	}
	return key
}

func (weapon *Weapon) GetWeight() float64 {
	return weapon.w
}
