package item

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/onorton/cowboysindians/icon"
)

type itemAttributes struct {
	Icon        icon.Icon
	Weight      float64
	Value       int
	Cover       bool
	Probability float64
}

var normalItemData map[string]itemAttributes
var normalItemProbabilities map[string]float64

func fetchItemData() {
	data, err := ioutil.ReadFile("data/item.json")
	check(err)
	var iD map[string]itemAttributes
	err = json.Unmarshal(data, &iD)
	check(err)
	normalItemData = iD

	normalItemProbabilities = make(map[string]float64)
	for name, attributes := range normalItemData {
		normalItemProbabilities[name] = attributes.Probability
	}
}

type NormalItem struct {
	baseItem
	cover       bool
	description *string
	corpse      bool
	ammoType    WeaponType
	armour      *armourComponent
	weapon      *weaponComponent
	consumable  *consumableComponent
}

func NewNormalItem(name string) Item {
	item := normalItemData[name]
	var itm Item = &NormalItem{baseItem{name, "", item.Icon, item.Weight, item.Value}, item.Cover, nil, false, NoAmmo, nil, nil, nil}
	return itm
}

func Money(amount int) Item {
	return &NormalItem{baseItem{"money", "", icon.NewIcon('$', 4), 0, amount}, false, nil, false, NoAmmo, nil, nil, nil}
}

func GenerateNormalItem() Item {
	return NewNormalItem(Choose(normalItemProbabilities))
}

func (item *NormalItem) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	nameValue, err := json.Marshal(item.name)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Name\":%s,", nameValue))

	ownerValue, err := json.Marshal(item.owner)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Owner\":%s,", ownerValue))

	iconValue, err := json.Marshal(item.ic)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Icon\":%s,", iconValue))

	weightValue, err := json.Marshal(item.w)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Weight\":%s,", weightValue))
	buffer.WriteString(fmt.Sprintf("\"Value\":%d,", item.v))

	coverValue, err := json.Marshal(item.cover)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Cover\":%s,", coverValue))

	descriptionValue, err := json.Marshal(item.description)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Description\":%s,", descriptionValue))

	corpseValue, err := json.Marshal(item.corpse)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Corpse\":%s,", corpseValue))

	ammoValue, err := json.Marshal(item.ammoType)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"AmmoType\":%s,", ammoValue))

	armourValue, err := json.Marshal(item.armour)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Armour\":%s,", armourValue))

	weaponValue, err := json.Marshal(item.weapon)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Weapon\":%s,", weaponValue))

	consumableValue, err := json.Marshal(item.consumable)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Consumable\":%s", consumableValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (item *NormalItem) UnmarshalJSON(data []byte) error {

	type itemJson struct {
		Name        string
		Owner       string
		Icon        icon.Icon
		Weight      float64
		Value       int
		Cover       bool
		Description *string
		Corpse      bool
		AmmoType    WeaponType
		Armour      *armourComponent
		Weapon      *weaponComponent
		Consumable  *consumableComponent
	}
	var v itemJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	item.name = v.Name
	item.owner = v.Owner
	item.ic = v.Icon
	item.w = v.Weight
	item.v = v.Value
	item.cover = v.Cover
	item.description = v.Description
	item.corpse = v.Corpse
	item.ammoType = v.AmmoType
	item.armour = v.Armour
	item.weapon = v.Weapon
	item.consumable = v.Consumable

	return nil
}

func (item *NormalItem) Owner() string {
	return item.owner
}

func (item *NormalItem) Owned(id string) bool {
	if item.owner == "" || item.corpse {
		return true
	}
	return item.owner == id
}

func (item *NormalItem) TransferOwner(newOwner string) {
	if item.corpse {
		return
	}

	// Only assign owner if item not owned
	if item.owner == "" {
		item.owner = newOwner
	}
}

func (item *NormalItem) IsReadable() bool {
	return item.description != nil
}

func (item *NormalItem) IsCorpse() bool {
	return item.corpse
}

func (item *NormalItem) IsAmmo() bool {
	return item.ammoType != NoAmmo
}

func (item *NormalItem) IsArmour() bool {
	return item.armour != nil
}

func (item *NormalItem) ACBonus() int {
	if item.armour != nil {
		return item.armour.Bonus
	}
	return 0
}

func (item *NormalItem) IsWeapon() bool {
	return item.weapon != nil
}

func (item *NormalItem) IsConsumable() bool {
	return item.consumable != nil
}

func (item *NormalItem) Effects(attr string) []Effect {
	if item.consumable != nil {
		return item.consumable.Effects[attr]
	}
	return []Effect{}
}

func (item *NormalItem) GivesCover() bool {
	return item.cover
}
