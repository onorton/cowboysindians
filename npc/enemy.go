package npc

import (
	"encoding/json"
	"io/ioutil"

	"github.com/onorton/cowboysindians/icon"
	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/worldmap"
	"github.com/rs/xid"
)

type EnemyAttributes struct {
	Icon         icon.Icon
	Initiative   int
	Hp           int
	Ac           int
	Str          int
	Dex          int
	Encumbrance  int
	Money        int
	Unarmed      item.WeaponComponent
	DialogueType *dialogueType
	AiType       string
	Inventory    [][]item.ItemChoice
	Mount        map[string]float64
	Probability  float64
	Human        bool
}

var enemyData map[string]EnemyAttributes = fetchEnemyData()

func fetchEnemyData() map[string]EnemyAttributes {
	data, err := ioutil.ReadFile("data/enemy.json")
	check(err)
	var eD map[string]EnemyAttributes
	err = json.Unmarshal(data, &eD)
	check(err)
	return eD
}

func RandomEnemyType() string {
	probabilities := map[string]float64{}
	for enemyType, enemyInfo := range enemyData {
		probabilities[enemyType] = enemyInfo.Probability
	}

	return chooseType(probabilities)
}

func NewEnemy(enemyType string, x, y int, world *worldmap.Map) *Npc {
	enemy := enemyData[enemyType]
	id := xid.New().String()
	dialogue := newDialogue(enemy.DialogueType, world, nil, nil)
	ai := newAi(enemy.AiType, world, worldmap.Coordinates{x, y}, nil, nil, dialogue, nil)
	attributes := map[string]*worldmap.Attribute{
		"hp":          worldmap.NewAttribute(enemy.Hp, enemy.Hp),
		"ac":          worldmap.NewAttribute(enemy.Ac, enemy.Ac),
		"str":         worldmap.NewAttribute(enemy.Str, enemy.Str),
		"dex":         worldmap.NewAttribute(enemy.Dex, enemy.Dex),
		"encumbrance": worldmap.NewAttribute(enemy.Encumbrance, enemy.Encumbrance)}
	name := generateName(enemyType, enemy.Human)
	e := &Npc{name, id, worldmap.Coordinates{x, y}, enemy.Icon, enemy.Initiative, attributes, worldmap.Enemy, false, enemy.Money, enemy.Unarmed, nil, nil, make([]*item.Item, 0), "", generateMount(enemy.Mount, x, y), world, ai, dialogue, enemy.Human}
	for _, itm := range generateInventory(enemy.Inventory) {
		e.PickupItem(itm)
	}
	return e
}
