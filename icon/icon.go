package icon

import (
	termbox "github.com/nsf/termbox-go"
)

func (i Icon) Render(x, y int) {
	termbox.SetCell(x, y, i.icon, i.colour, termbox.ColorDefault)
}

func (i Icon) RenderDoor(x, y int, passable bool) {
	if passable {
		termbox.SetCell(x, y, ' ', i.colour, termbox.ColorDefault)
	} else {
		i.Render(x, y)
	}
}

func CreatePlayerIcon() Icon {
	return Icon{'@', termbox.ColorWhite}
}

func NewIcon(icon rune, colour termbox.Attribute) Icon {
	return Icon{icon, colour}
}

type Icon struct {
	icon   rune
	colour termbox.Attribute
}
