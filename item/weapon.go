package item

import (
	"encoding/json"
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/icon"
	"hash/fnv"
	"io/ioutil"
	"math/rand"
	"strings"
)

type WeaponAttributes struct {
	Colour termbox.Attribute
	Icon   rune
	Damage DamageAttributes
}

type DamageAttributes struct {
	Dice   int
	Number int
	Bonus  int
}

var weaponData map[string]WeaponAttributes = fetchWeaponData()

func fetchWeaponData() map[string]WeaponAttributes {
	data, err := ioutil.ReadFile("data/weapon.json")
	check(err)
	var eD map[string]WeaponAttributes
	err = json.Unmarshal(data, &eD)
	check(err)
	return eD
}

type Weapon struct {
	name   string
	ic     icon.Icon
	damage Damage
}

type Damage struct {
	dice   int
	number int
	bonus  int
}

func NewWeapon(name string) *Weapon {
	weapon := weaponData[name]
	return &Weapon{name, icon.NewIcon(weapon.Icon, weapon.Colour), Damage{weapon.Damage.Dice, weapon.Damage.Number, weapon.Damage.Bonus}}
}

func (weapon *Weapon) Serialize() string {
	if weapon == nil {
		return ""
	}
	return fmt.Sprintf("Weapon{%s %s %d}", weapon.name, weapon.ic.Serialize(), weapon.damage)
}

func DeserializeWeapon(weaponString string) *Weapon {

	if len(weaponString) == 1 {
		return nil
	}
	weaponString = weaponString[7 : len(weaponString)-2]
	weapon := new(Weapon)
	weaponAttributes := strings.SplitN(weaponString, " ", 3)
	weapon.name = weaponAttributes[0]
	weapon.ic = icon.Deserialize(weaponAttributes[1])

	return weapon
}

func (weapon *Weapon) GetName() string {
	return weapon.name
}
func (weapon *Weapon) Render(x, y int) {

	weapon.ic.Render(x, y)
}

func (weapon *Weapon) GetDamage() int {
	result := 0
	for i := 0; i < weapon.damage.number; i++ {
		result += rand.Intn(weapon.damage.dice) + 1
	}

	result += weapon.damage.bonus
	return result
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