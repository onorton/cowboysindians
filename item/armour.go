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

type ArmourAttributes struct {
	Colour termbox.Attribute
	Icon   rune
	Bonus  int
}

var armourData map[string]ArmourAttributes = fetchArmourData()

func fetchArmourData() map[string]ArmourAttributes {
	data, err := ioutil.ReadFile("data/armour.json")
	check(err)
	var eD map[string]ArmourAttributes
	err = json.Unmarshal(data, &eD)
	check(err)
	return eD
}

type Armour struct {
	name  string
	ic    icon.Icon
	bonus int
}

func NewArmour(name string) *Armour {
	armour := armourData[name]
	return &Armour{name, icon.NewIcon(armour.Icon, armour.Colour), armour.Bonus}
}

func (armour *Armour) Serialize() string {
	if armour == nil {
		return ""
	}
	return fmt.Sprintf("Armour{%s %s %d}", armour.name, armour.ic.Serialize(), armour.bonus)
}

func DeserializeArmour(armourString string) *Armour {

	if len(armourString) == 1 {
		return nil
	}
	armourString = armourString[7 : len(armourString)-2]
	armour := new(Armour)
	armourAttributes := strings.SplitN(armourString, " ", 3)
	armour.name = armourAttributes[0]
	armour.ic = icon.Deserialize(armourAttributes[1])

	return armour
}

func (armour *Armour) GetName() string {
	return armour.name
}
func (armour *Armour) Render(x, y int) {

	armour.ic.Render(x, y)
}

func (armour *Armour) GetACBonus() int {
	return armour.bonus
}

func (armour *Armour) GetKey() rune {
	h := fnv.New32()
	h.Write([]byte(armour.name))
	key := rune(33 + h.Sum32()%93)
	if key == '*' {
		key++
	}
	return key
}
