package icon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	termbox "github.com/nsf/termbox-go"
)

type Icon struct {
	icon   rune
	colour termbox.Attribute
}

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

func Deserialize(icon string) Icon {
	b := 0
	e := len(icon)
	for i, c := range icon {
		if c == '{' {
			b = i
		}
		if c == '}' {
			e = i
		}
	}
	result := icon[b+1 : e]
	fields := strings.Split(result, " ")
	iconRune, _ := strconv.Atoi(fields[0])
	colourNumber, _ := strconv.Atoi(fields[1])
	return Icon{rune(iconRune), termbox.Attribute(colourNumber)}

}

func (i Icon) Serialize() string {
	return fmt.Sprintf("Icon{%d %d}", i.icon, i.colour)
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
