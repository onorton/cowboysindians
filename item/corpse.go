package item

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/onorton/cowboysindians/icon"
)

type corpseAttributes struct {
	Icon        icon.Icon
	Components  map[string]interface{}
	Weight      float64
	Value       int
	Probability float64
}

var corpseData map[string]corpseAttributes

func fetchCorpseData() {
	data, err := ioutil.ReadFile("data/corpse.json")
	check(err)
	var cD map[string]corpseAttributes
	err = json.Unmarshal(data, &cD)
	check(err)
	corpseData = cD
}

func NewCorpse(corpseType string, owner string, ownerName string, ownerIcon icon.Icon) *Item {
	corpse := corpseData[corpseType]
	name := fmt.Sprintf("%s's %s", ownerName, corpseType)
	return &Item{name, owner, icon.NewCorpseIcon(ownerIcon), corpse.Weight, corpse.Value, UnmarshalComponents(corpse.Components)}
}
