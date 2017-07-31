package item

import (
	"encoding/json"
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/icon"
	"hash/fnv"
	"io/ioutil"
	"strings"
)

type AmmoAttributes struct {
	Colour termbox.Attribute
	Icon   rune
	Type   WeaponType
}

var ammoData map[string]AmmoAttributes = fetchAmmoData()

func fetchAmmoData() map[string]AmmoAttributes {
	data, err := ioutil.ReadFile("data/ammo.json")
	check(err)
	var eD map[string]AmmoAttributes
	err = json.Unmarshal(data, &eD)
	check(err)
	return eD
}

type Ammo struct {
	name string
	ic   icon.Icon
	t    WeaponType
}

func NewAmmo(name string) Item {
	ammo := ammoData[name]
	var itm Item = &Ammo{name, icon.NewIcon(ammo.Icon, ammo.Colour), ammo.Type}
	return itm
}

func (ammo *Ammo) Serialize() string {
	if ammo == nil {
		return ""
	}
	return fmt.Sprintf("Ammo{%s %s}", strings.Replace(ammo.name, " ", "_", -1), ammo.ic.Serialize())
}

func DeserializeAmmo(ammoString string) Item {

	if len(ammoString) == 0 {
		return nil
	}
	ammoString = ammoString[1 : len(ammoString)-2]
	ammo := new(Ammo)
	ammoAttributes := strings.SplitN(ammoString, " ", 2)
	ammo.name = strings.Replace(ammoAttributes[0], "_", " ", -1)
	ammo.ic = icon.Deserialize(ammoAttributes[1])
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
