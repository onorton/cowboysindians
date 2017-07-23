package item

import (
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/icon"
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

func (item *Item) Render(x, y int) {

	item.ic.Render(x, y)
}
