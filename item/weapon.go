package item

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"math/rand"

	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/ui"
)

type WeaponAttributes struct {
	Icon     icon.Icon
	Damage   DamageAttributes
	Range    int
	Type     WeaponType
	Capacity int
	Weight   float64
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
	return &Weapon{name, weapon.Icon, weapon.Range, weapon.Type, weapon.Weight, weaponCapacity, &Damage{weapon.Damage.Dice, weapon.Damage.Number, weapon.Damage.Bonus}}
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
		Icon           icon.Icon
		Range          *int
		Type           *WeaponType
		Weight         float64
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
	weapon.ic = v.Icon
	weapon.r = *(v.Range)
	weapon.t = *(v.Type)
	weapon.w = v.Weight
	weapon.wc = v.WeaponCapacity
	weapon.damage = v.Damage

	return nil
}

func (weapon *Weapon) GetName() string {
	return weapon.name
}
func (weapon *Weapon) Render() ui.Element {
	return weapon.ic.Render()
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

func (weapon *Weapon) GivesCover() bool {
	return false
}
