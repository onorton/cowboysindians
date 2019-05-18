package worldmap

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/onorton/cowboysindians/item"
)

type doorComponent struct {
	locked        bool
	key           int32
	blocksVClosed bool
	open          bool
}

func (door *doorComponent) Open() bool {
	return door.open
}

func (door *doorComponent) Key() int32 {
	return door.key
}

func (door *doorComponent) Lock() {
	door.locked = true
}

func (door *doorComponent) ToggleLocked() {
	door.locked = !door.locked
}

func (door *doorComponent) Locked() bool {
	return door.locked
}

func (door *doorComponent) KeyFits(itm *item.Item) bool {
	keyComponent := itm.Component("key").(item.KeyComponent)
	if keyComponent.Key == -1 {
		return true
	}
	return keyComponent.Key == door.Key()
}

func (door *doorComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	keys := []string{"Locked", "Key", "BlocksVisionClosed", "Open"}

	gridValues := map[string]interface{}{
		"Locked":             door.locked,
		"Key":                door.key,
		"BlocksVisionClosed": door.blocksVClosed,
		"Open":               door.open,
	}

	length := len(gridValues)
	count := 0

	for _, key := range keys {
		jsonValue, err := json.Marshal(gridValues[key])
		if err != nil {
			return nil, err
		}
		buffer.WriteString(fmt.Sprintf("\"%s\":%s", key, jsonValue))
		count++
		if count < length {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (door *doorComponent) UnmarshalJSON(data []byte) error {

	type doorJson struct {
		Locked             bool
		Key                int32
		BlocksVisionClosed bool
		Open               bool
	}
	v := doorJson{}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	door.locked = v.Locked
	door.key = v.Key
	door.blocksVClosed = v.BlocksVisionClosed
	door.open = v.Open
	return nil
}
