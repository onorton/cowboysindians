package message

import (
	"fmt"
	"unicode"

	"github.com/onorton/cowboysindians/structs"
	"github.com/onorton/cowboysindians/ui"
)

// Singleton message queue
var Mq = MessageQueue{new(structs.Queue), 0, 0}

type MessageQueue struct {
	queue        *structs.Queue
	windowWidth  int
	windowHeight int
}

func clearMessageBar() {

	width := Mq.windowWidth
	cells := make([]ui.Cell, width, width)
	for i := 0; i < width; i++ {
		cells[i] = ui.NewCell(i, Mq.windowHeight)
	}
	ui.ClearCells(cells)

}

// Prints first message in queue and prompts if there are any more
func PrintMessages() {
	clearMessageBar()
	m := Mq.queue.Dequeue().(string)
	if !Mq.queue.IsEmpty() {
		m += " --MORE--"

	}
	ui.WriteText(0, Mq.windowHeight, m)
}

// Prints single message immediately to message bar
func PrintMessage(m string) {

	clearMessageBar()
	ui.WriteText(0, Mq.windowHeight, capitalise(m))
}

func RequestInput(m string) string {
	clearMessageBar()
	input := ""
	PrintMessage(m)
	for {
		action, c := ui.TextInput()
		switch action {
		case ui.Erase:
			if len(input) > 0 {
				input = input[:len(input)-1]
			}
		case ui.Character:
			input = input + string(c)
		}

		if action == ui.Done {
			break
		}
		PrintMessage(fmt.Sprintf("%s %s", m, input))
	}
	clearMessageBar()
	return input
}

func SetWindowSize(windowWidth, windowHeight int) {
	Mq.windowWidth = windowWidth
	Mq.windowHeight = windowHeight
}

func Enqueue(m string) {
	Mq.queue.Enqueue(capitalise(m))
}

func capitalise(m string) string {
	runes := []rune(m)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
