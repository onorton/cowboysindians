package item

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/onorton/cowboysindians/icon"
)

type corpseAttributes struct {
	Icon        icon.Icon
	Weight      float64
	Value       int
	Cover       bool
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

func NewCorpse(corpseType string, owner string, ownerName string, ownerIcon icon.Icon) Item {
	corpse := corpseData[corpseType]
	name := fmt.Sprintf("%s's %s", ownerName, corpseType)
	var itm Item = &NormalItem{baseItem{name, owner, icon.NewCorpseIcon(ownerIcon), corpse.Weight, corpse.Value}, corpse.Cover, nil, true, NoAmmo, nil, nil}
	return itm
}
