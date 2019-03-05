package icon

import (
	"bytes"
	"encoding/json"
	"fmt"

	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/ui"
)

type Icon struct {
	icon   rune
	colour termbox.Attribute
}

func (i Icon) Render() ui.Element {
	return ui.NewElement(i.icon, i.colour)
}

func MergeIcons(f, b Icon) ui.Element {
	return ui.NewElementWithBg(f.icon, f.colour, b.colour)
}

func CreatePlayerIcon() Icon {
	return Icon{'@', termbox.ColorWhite}
}

func NewIcon(icon rune, colour termbox.Attribute) Icon {
	return Icon{icon, colour}
}

func (i Icon) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	iconValue, err := json.Marshal(i.icon)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Icon\":%s,", iconValue))

	colourValue, err := json.Marshal(i.colour)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Colour\":%s", colourValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (i *Icon) UnmarshalJSON(data []byte) error {

	type iconJson struct {
		Icon   rune
		Colour termbox.Attribute
	}

	var v iconJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	i.icon = v.Icon
	i.colour = v.Colour

	return nil
}
