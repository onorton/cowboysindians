package message

import (
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/structs"
)

// Singleton message queue
var Mq = MessageQueue{new(structs.Queue), 0, 0}

type MessageQueue struct {
	queue        *structs.Queue
	windowWidth  int
	windowHeight int
}

func print_tb(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func clearMessageBar() {

	for i := 0; i < Mq.windowWidth; i++ {
		termbox.SetCell(i, Mq.windowHeight, ' ', termbox.ColorDefault, termbox.ColorDefault)
	}

}

// Prints first message in queue and prompts if there are any more
func PrintMessages() {
	clearMessageBar()
	m := Mq.queue.Dequeue().(string)
	if !Mq.queue.IsEmpty() {
		m += " --MORE--"

	}
	print_tb(0, Mq.windowHeight, termbox.ColorWhite, termbox.ColorDefault, m)
	termbox.Flush()
}

// Prints single message immediately to message bar
func PrintMessage(m string) {

	clearMessageBar()
	print_tb(0, Mq.windowHeight, termbox.ColorWhite, termbox.ColorDefault, m)
	termbox.Flush()
}

func SetWindowSize(windowWidth, windowHeight int) {
	Mq.windowWidth = windowWidth
	Mq.windowHeight = windowHeight
}

func Enqueue(m string) {
	Mq.queue.Enqueue(m)
}
