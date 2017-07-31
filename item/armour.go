package item

import (
	"encoding/json"
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/icon"
	"hash/fnv"
	"io/ioutil"
	"strconv"
	"strings"
)

type ArmourAttributes struct {
	Colour termbox.Attribute
	Icon   rune
	Bonus  int
	Weight float64
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
	w     float64
}

func NewArmour(name string) *Armour {
	armour := armourData[name]
	return &Armour{name, icon.NewIcon(armour.Icon, armour.Colour), armour.Bonus, armour.Weight}
}

func (armour *Armour) Serialize() string {
	if armour == nil {
		return ""
	}
	return fmt.Sprintf("Armour{%s %s %d %f}", strings.Replace(armour.name, " ", "_", -1), armour.ic.Serialize(), armour.bonus, armour.w)
}

func DeserializeArmour(armourString string) *Armour {

	if len(armourString) == 1 {
		return nil
	}
	armourString = armourString[1 : len(armourString)-2]
	armour := new(Armour)
	armourAttributes := strings.SplitN(armourString, " ", 5)
	armour.name = strings.Replace(armourAttributes[0], "_", " ", -1)
	armour.ic = icon.Deserialize(armourAttributes[1] + " " + armourAttributes[2])
	armour.bonus, _ = strconv.Atoi(armourAttributes[3])
	armour.w, _ = strconv.ParseFloat(armourAttributes[4], 64)

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

func (armour *Armour) GetWeight() float64 {
	return armour.w
}
