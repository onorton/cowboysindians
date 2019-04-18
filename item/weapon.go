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

type Weapon struct {
	baseItem
	r      int
	t      WeaponType
	wc     *WeaponCapacity
	damage *Damage
}

type Damage struct {
	dice   int
	number int
	bonus  int
}

type WeaponCapacity struct {
	capacity int
	loaded   int
}

func NewWeapon(name string) *Weapon {
	weapon := weaponData[name]

	var weaponCapacity *WeaponCapacity
	if weapon.Capacity != 0 {
		weaponCapacity = &WeaponCapacity{weapon.Capacity, 0}
	}
	return &Weapon{baseItem{name, "", weapon.Icon, weapon.Weight, weapon.Value}, weapon.Range, weapon.Type, weaponCapacity, &Damage{weapon.Damage.Dice, weapon.Damage.Number, weapon.Damage.Bonus}}
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

func (weapon *Weapon) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	nameValue, err := json.Marshal(weapon.name)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Name\":%s,", nameValue))

	ownerValue, err := json.Marshal(weapon.owner)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Owner\":%s,", ownerValue))

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
	buffer.WriteString(fmt.Sprintf("\"Value\":%d,", weapon.v))

	if weapon.wc != nil {
		weaponCapacityValue, err := json.Marshal(weapon.wc)
		if err != nil {
			return nil, err
		}

		buffer.WriteString(fmt.Sprintf("\"WeaponCapacity\":%s,", weaponCapacityValue))
	}
	damageValue, err := json.Marshal(weapon.damage)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Damage\":%s", damageValue))
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

func (weapon *Weapon) UnmarshalJSON(data []byte) error {

	type weaponJson struct {
		Name           string
		Owner          string
		Icon           icon.Icon
		Range          *int
		Type           *WeaponType
		Weight         float64
		Value          int
		WeaponCapacity *WeaponCapacity
		Damage         *Damage
	}
	var v weaponJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if v.Range == nil {
		return fmt.Errorf("The Range field is required")
	}

	if v.Type == nil {
		return fmt.Errorf("The Type field is required")
	}

	if v.Damage == nil {
		return fmt.Errorf("The Damage field is required")
	}

	weapon.name = v.Name
	weapon.owner = v.Owner
	weapon.ic = v.Icon
	weapon.r = *(v.Range)
	weapon.t = *(v.Type)
	weapon.w = v.Weight
	weapon.v = v.Value
	weapon.wc = v.WeaponCapacity
	weapon.damage = v.Damage

	return nil
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

func (weapon *Weapon) NeedsAmmo() bool {
	return weapon.wc != nil
}

func (weapon *Weapon) IsUnloaded() bool {
	return weapon.wc.loaded == 0
}

func (weapon *Weapon) IsFullyLoaded() bool {
	return weapon.wc.capacity == weapon.wc.loaded
}

func (weapon *Weapon) Load() {
	weapon.wc.loaded++
}

func (weapon *Weapon) Fire() {
	if weapon.wc != nil && weapon.wc.loaded > 0 {
		weapon.wc.loaded--
	}
}

func (weapon *Weapon) GivesCover() bool {
	return false
}
