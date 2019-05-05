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

func NewCorpse(corpseType string, owner string, ownerName string, ownerIcon icon.Icon) *Item {
	corpse := corpseData[corpseType]
	name := fmt.Sprintf("%s's %s", ownerName, corpseType)
	components := map[string]component{"corpse": tag{}}
	if corpse.Cover {
		components["cover"] = tag{}
	}
	return &Item{name, owner, icon.NewCorpseIcon(ownerIcon), corpse.Weight, corpse.Value, components, nil, nil, nil, nil}
}
