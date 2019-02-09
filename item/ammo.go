package item

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/onorton/cowboysindians/icon"
)

type AmmoAttributes struct {
	Icon   icon.Icon
	Type   WeaponType
	Weight float64
}

var ammoData map[string]AmmoAttributes

func fetchAmmoData() {
	data, err := ioutil.ReadFile("data/ammo.json")
	check(err)
	var aD map[string]AmmoAttributes
	err = json.Unmarshal(data, &aD)
	check(err)
	ammoData = aD
}

type Ammo struct {
	name string
	ic   icon.Icon
	t    WeaponType
	w    float64
}

func NewAmmo(name string) Item {
	ammo := ammoData[name]
	var itm Item = &Ammo{name, ammo.Icon, ammo.Type, ammo.Weight}
	return itm
}

func (ammo *Ammo) Serialize() string {
	if ammo == nil {
		return ""
	}

	iconJson, err := json.Marshal(ammo.ic)
	check(err)

	return fmt.Sprintf("Ammo{%s %d %f %s}", strings.Replace(ammo.name, " ", "_", -1), ammo.t, ammo.w, iconJson)
}

func DeserializeAmmo(ammoString string) Item {

	if len(ammoString) == 0 {
		return nil
	}
	ammoString = ammoString[1 : len(ammoString)-2]
	ammo := new(Ammo)
	ammoAttributes := strings.SplitN(ammoString, " ", 4)
	ammo.name = strings.Replace(ammoAttributes[0], "_", " ", -1)

	err := json.Unmarshal([]byte(ammoAttributes[3]), &(ammo.ic))
	check(err)

	t, _ := strconv.Atoi(ammoAttributes[1])
	ammo.t = WeaponType(t)
	ammo.w, _ = strconv.ParseFloat(ammoAttributes[2], 64)
	var itm Item = ammo
	return itm
}

func (ammo *Ammo) GetName() string {
	return ammo.name
}
func (ammo *Ammo) Render(x, y int) {

	ammo.ic.Render(x, y)
}

func (ammo *Ammo) GetKey() rune {
	h := fnv.New32()
	h.Write([]byte(ammo.name))
	key := rune(33 + h.Sum32()%93)
	if key == '*' {
		key++
	}
	return key
}

func (ammo *Ammo) GetWeight() float64 {
	return ammo.w
}
