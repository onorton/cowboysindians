package item

import (
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/icon"
	"hash/fnv"
	"strings"
)

type Item struct {
	name string
	ic   icon.Icon
}

func NewItem(name string) *Item {
	item := new(Item)
	item.name = name
	item.ic = icon.NewIcon('*', termbox.ColorYellow)
	return item
}

func (item *Item) Serialize() string {
	if item == nil {
		return ""
	}
	return fmt.Sprintf("Item{%s %s}", item.name, item.ic.Serialize())
}

func Deserialize(itemString string) *Item {

	if len(itemString) == 1 {
		return nil
	}
	itemString = itemString[5 : len(itemString)-2]
	item := new(Item)
	itemAttributes := strings.SplitN(itemString, " ", 2)
	item.name = itemAttributes[0]
	item.ic = icon.Deserialize(itemAttributes[1])
	return item
}

func (item *Item) GetName() string {
	return item.name
}
func (item *Item) Render(x, y int) {

	item.ic.Render(x, y)
}

func (item *Item) GetKey() rune {
	h := fnv.New32()
	h.Write([]byte(item.name))
	key := rune(33 + h.Sum32()%93)
	if key == '*' {
		key++
	}
	return key
}
