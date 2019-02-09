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

type ArmourAttributes struct {
	Icon   icon.Icon
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
	return &Armour{name, armour.Icon, armour.Bonus, armour.Weight}
}

func (armour *Armour) Serialize() string {
	if armour == nil {
		return ""
	}

	iconJson, err := json.Marshal(armour.ic)
	check(err)

	return fmt.Sprintf("Armour{%s %s %d %f}", strings.Replace(armour.name, " ", "_", -1), iconJson, armour.bonus, armour.w)
}

func DeserializeArmour(armourString string) *Armour {

	if len(armourString) == 1 {
		return nil
	}
	armourString = armourString[1 : len(armourString)-2]
	armour := new(Armour)
	armourAttributes := strings.SplitN(armourString, " ", 5)
	armour.name = strings.Replace(armourAttributes[0], "_", " ", -1)

	err := json.Unmarshal([]byte(armourAttributes[1]), &(armour.ic))
	check(err)

	armour.bonus, _ = strconv.Atoi(armourAttributes[2])
	armour.w, _ = strconv.ParseFloat(armourAttributes[2], 64)

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
