package item

import (
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/icon"
	"hash/fnv"
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

func (item *Item) GetName() string {
	return item.name
}
func (item *Item) Render(x, y int) {

	item.ic.Render(x, y)
}

func (item *Item) GetKey() rune {
	h := fnv.New32()
	h.Write([]byte(item.name))
	return rune(33 + h.Sum32()%93)
}
