package item

import (
	"encoding/json"
	"math/rand"

	"github.com/onorton/cowboysindians/ui"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type ItemDefinition struct {
	Category string
	Name     string
	Amount   int
}

func SelectItem(probabilites map[string]float64) string {
	max := 0.0

	for _, probability := range probabilites {
		inverse := 1.0 / probability
		if inverse > max {
			max = inverse
		}
	}
	items := make([]string, 0)

	for name, probability := range probabilites {
		count := int(probability * max)
		for i := 0; i < count; i++ {
			items = append(items, name)
		}
	}
	n := rand.Intn(len(items))
	return items[n]
}

type ItemList []Item

func (itemList *ItemList) UnmarshalJSON(data []byte) error {
	var rawItems []interface{}

	if err := json.Unmarshal(data, &rawItems); err != nil {
		return err
	}

	items := []Item{}

	//convert raw items back into byte data and unmarshal individually,
	for _, rawItem := range rawItems {
		itemJson, err := json.Marshal(rawItem)
		check(err)

		//Tying unmarshalling for each item type

		//armour
		armour := &Armour{}
		err = json.Unmarshal(itemJson, armour)
		if err == nil {
			items = append(items, armour)
			continue
		}

		//consumable
		consumable := &Consumable{}
		err = json.Unmarshal(itemJson, consumable)
		if err == nil {
			items = append(items, consumable)
			continue
		}

		//weapon
		weapon := &Weapon{}
		err = json.Unmarshal(itemJson, weapon)
		if err == nil {
			items = append(items, weapon)
			continue
		}

		//ammo
		ammo := &Ammo{}
		err = json.Unmarshal(itemJson, ammo)
		if err == nil {
			items = append(items, ammo)
			continue
		}

		//Must be a plain ordinary item
		item := &NormalItem{}
		err = json.Unmarshal(itemJson, item)
		if err == nil {
			items = append(items, item)
			continue
		} else {
			return err
		}
	}

	*itemList = ItemList(items)

	return nil
}

func GenerateItem() Item {
	// Pick random type (all same probability)

	n := rand.Intn(5)

	var itm Item = nil
	switch n {
	case 0:
		itm = GenerateAmmo()
	case 1:
		itm = GenerateArmour()
	case 2:
		itm = GenerateConsumable()
	case 3:
		itm = GenerateNormalItem()
	case 4:
		itm = GenerateWeapon()
	}

	return itm
}

func LoadAllData() {
	fetchAmmoData()
	fetchArmourData()
	fetchConsumableData()
	fetchItemData()
	fetchWeaponData()
}

type Item interface {
	GetKey() rune
	GetName() string
	Render() ui.Element
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
	GetWeight() float64
	GivesCover() bool
}
