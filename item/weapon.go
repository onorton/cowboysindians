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

type weaponComponent struct {
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

func NewWeapon(name string) *NormalItem {
	weapon := weaponData[name]

	var weaponCapacity *WeaponCapacity
	if weapon.Capacity != 0 {
		weaponCapacity = &WeaponCapacity{weapon.Capacity, 0}
	}
	wc := weaponComponent{weapon.Range, weapon.Type, weaponCapacity, Damage{weapon.Damage.Dice, weapon.Damage.Number, weapon.Damage.Bonus}}
	return &NormalItem{baseItem{name, "", weapon.Icon, weapon.Weight, weapon.Value}, false, nil, false, NoAmmo, nil, &wc}
}

func GenerateWeapon() Item {
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

func (item *NormalItem) Range() int {
	if item.weapon != nil {
		return item.weapon.Range
	}
	return 0
}

// Maximum possible damage
func (item *NormalItem) GetMaxDamage() int {
	if item.weapon != nil {
		return item.weapon.Damage.max()
	}
	return 0
}
func (item *NormalItem) GetDamage() int {
	if item.weapon == nil {
		return 0
	}

	result := 0
	for i := 0; i < item.weapon.Damage.number; i++ {
		result += rand.Intn(item.weapon.Damage.dice) + 1
	}

	result += item.weapon.Damage.bonus
	return result
}

func (item *NormalItem) AmmoTypeMatches(ammo *NormalItem) bool {
	if item.weapon != nil {
		return item.weapon.Type == ammo.ammoType
	}
	return false
}

func (item *NormalItem) NeedsAmmo() bool {
	if item.weapon != nil {
		return item.weapon.Capacity != nil
	}
	return false
}

func (item *NormalItem) IsUnloaded() bool {
	if item.weapon != nil {
		return item.weapon.Capacity.loaded == 0
	}
	return false
}

func (item *NormalItem) IsFullyLoaded() bool {
	if item.weapon != nil {
		return item.weapon.Capacity.loaded == item.weapon.Capacity.capacity
	}
	return false
}

func (item *NormalItem) Load() {
	if item.weapon != nil && item.weapon.Capacity != nil {
		item.weapon.Capacity.loaded++
	}
}

func (item *NormalItem) Fire() {
	if item.weapon != nil && item.weapon.Capacity != nil {
		if item.weapon.Capacity.loaded > 0 {
			item.weapon.Capacity.loaded--
		}
	}
}
