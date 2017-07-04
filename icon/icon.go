package icon

import (
	termbox "github.com/nsf/termbox-go"
)

func (i Icon) Render(x, y int) {
	termbox.SetCell(x, y, i.icon, i.colour, termbox.ColorDefault)
}

func NewIcon(icon rune, colour termbox.Attribute) Icon {
	return Icon{icon, colour}
}

type Icon struct {
	icon   rune
	colour termbox.Attribute
}
