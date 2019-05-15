package item

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"

	"github.com/onorton/cowboysindians/icon"
)

type WeaponAttributes struct {
	Icon        icon.Icon
	Damage      DamageAttributes
	Range       int
	Type        WeaponType
	Capacity    int
	Weight      float64
	Value       int
	Probability float64
}

type DamageAttributes struct {
	Dice   int
	Number int
	Bonus  int
}

type WeaponType int

const (
	NoAmmo WeaponType = iota
	Pistol
	Shotgun
	Rifle
	Bow
)

var weaponData map[string]WeaponAttributes
var weaponProbabilities map[string]float64

func fetchWeaponData() {
	data, err := ioutil.ReadFile("data/weapon.json")
	check(err)
	var wD map[string]WeaponAttributes
	err = json.Unmarshal(data, &wD)
	check(err)
	weaponData = wD

	weaponProbabilities = make(map[string]float64)
	for name, attributes := range weaponData {
		weaponProbabilities[name] = attributes.Probability
	}
}

type WeaponComponent struct {
	Range    int
	Type     WeaponType
	Capacity *WeaponCapacity
	Damage   Damage
}

type Damage struct {
	dice   int
	number int
	bonus  int
}

func (damage Damage) max() int {
	return damage.number*damage.dice + damage.bonus
}

type WeaponCapacity struct {
	capacity int
	loaded   int
}

func NewWeapon(name string) *Item {
	weapon := weaponData[name]

	var weaponCapacity *WeaponCapacity
	if weapon.Capacity != 0 {
		weaponCapacity = &WeaponCapacity{weapon.Capacity, 0}
	}
	wc := WeaponComponent{weapon.Range, weapon.Type, weaponCapacity, Damage{weapon.Damage.Dice, weapon.Damage.Number, weapon.Damage.Bonus}}
	return &Item{name, "", weapon.Icon, weapon.Weight, weapon.Value, map[string]component{"weapon": wc}}
}

func GenerateWeapon() *Item {
	return NewWeapon(Choose(weaponProbabilities))
}

func (weaponCapacity *WeaponCapacity) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	capacityValue, err := json.Marshal(weaponCapacity.capacity)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Capacity\":%s,", capacityValue))

	loadedValue, err := json.Marshal(weaponCapacity.loaded)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Loaded\":%s", loadedValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (damage Damage) MarshalJSON() ([]byte, error) {
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

func (weaponCapacity *WeaponCapacity) UnmarshalJSON(data []byte) error {

	type weaponCapacityJson struct {
		Capacity int
		Loaded   int
	}
	var v weaponCapacityJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	weaponCapacity.capacity = v.Capacity
	weaponCapacity.loaded = v.Loaded

	return nil
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

func (weapon WeaponComponent) MaxDamage() int {
	return weapon.Damage.max()
}

func (weapon WeaponComponent) GetDamage() int {

	result := 0
	for i := 0; i < weapon.Damage.number; i++ {
		result += rand.Intn(weapon.Damage.dice) + 1
	}

	result += weapon.Damage.bonus
	return result
}

func (weapon WeaponComponent) AmmoTypeMatches(ammo *Item) bool {
	return weapon.Type == ammo.Component("ammo").(AmmoComponent).AmmoType
}

func (weapon WeaponComponent) NeedsAmmo() bool {
	return weapon.Capacity != nil
}

func (weapon WeaponComponent) IsUnloaded() bool {
	return weapon.Capacity.loaded == 0

}

func (weapon WeaponComponent) IsFullyLoaded() bool {
	return weapon.Capacity.loaded == weapon.Capacity.capacity
}

func (weapon WeaponComponent) Load() {
	if weapon.Capacity != nil {
		weapon.Capacity.loaded++
	}
}

func (weapon WeaponComponent) Fire() {
	if weapon.Capacity != nil && weapon.Capacity.loaded > 0 {
		weapon.Capacity.loaded--
	}
}
