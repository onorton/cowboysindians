package main

import (
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/structs"
)

const width = 100
const height = 25

func print_tb(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func print_message(messages *structs.Queue) {
	for i := 0; i < width; i++ {
		termbox.SetCell(i, height+1, ' ', termbox.ColorDefault, termbox.ColorDefault)
	}
	m := messages.Dequeue().(string)
	if !messages.IsEmpty() {
		m += " --MORE--"

	}
	print_tb(width/2-len(m)/2, height+1, termbox.ColorWhite, termbox.ColorDefault, m)
	termbox.Flush()
}
func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	messages := new(structs.Queue)
	x := 0
	y := 0
	for {
		quit := false
		endTurn := false
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(x, y, '@', termbox.ColorWhite, termbox.ColorDefault)
		print_message(messages)
		for {
			e := termbox.PollEvent()
			if e.Type == termbox.EventKey {
				switch e.Key {
				case termbox.KeyArrowLeft:
					if x != 0 {
						x--
					}
				case termbox.KeyArrowRight:
					if x < width {
						x++
					}
				case termbox.KeyArrowUp:
					if y != 0 {
						y--
					}
				case termbox.KeyArrowDown:
					if y < height {
						y++
					}
				case termbox.KeySpace:
					{
						print_message(messages)
					}
				default:
					{

						switch e.Ch {
						case '1':
							if x != 0 && y < height {
								x--
								y++
							}
						case '2':
							if y < height {
								y++
							}
						case '3':
							if x < width && y < height {
								x++
								y++
							}

						case '4':
							if x != 0 {
								x--
							}
						case '5':
						case '6':
							if x < width {
								x++
							}
						case '7':
							if x != 0 && y != 0 {
								x--
								y--
							}
						case '8':
							if y != 0 {
								y--
							}
						case '9':
							if y != 0 && x < width {
								y--
								x++
							}
						default:
							quit = true
						}
					}
				}

				endTurn = e.Key != termbox.KeySpace
				if endTurn || quit {
					break
				}
			} else {
				break
			}
		}

		if quit {
			break
		}

	}

}
