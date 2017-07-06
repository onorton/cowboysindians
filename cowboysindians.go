package main

import (
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/creature"
	"github.com/onorton/cowboysindians/structs"
	"github.com/onorton/cowboysindians/worldmap"
)

const windowWidth = 100
const windowHeight = 25
const width = 400
const height = 100

func print_tb(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func print_message(messages *structs.Queue) {
	for i := 0; i < width; i++ {
		termbox.SetCell(i, windowWidth+1, ' ', termbox.ColorDefault, termbox.ColorDefault)
	}
	m := messages.Dequeue().(string)
	if !messages.IsEmpty() {
		m += " --MORE--"

	}
	print_tb(0, windowHeight, termbox.ColorWhite, termbox.ColorDefault, m)
	termbox.Flush()
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	messages := new(structs.Queue)

	worldMap := worldmap.NewMap(width, height, windowWidth, windowHeight)
	player := creature.NewPlayer()

	for {

		quit := false
		endTurn := false
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		worldMap.Render()
		x := player.X
		y := player.Y
		messages.Enqueue(fmt.Sprintf("%d %d", x, y))
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
					if x < width-1 {
						x++
					}
				case termbox.KeyArrowUp:
					if y != 0 {
						y--
					}
				case termbox.KeyArrowDown:
					if y < height-1 {
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
							if x != 0 && y < height-1 {
								x--
								y++
							}
						case '2':
							if y < height-1 {
								y++
							}
						case '3':
							if x < width-1 && y < height-1 {
								x++
								y++
							}

						case '4':
							if x != 0 {
								x--
							}
						case '5':
						case '6':
							if x < width-1 {
								x++
							}
						case '7':
							if x != 0 && x != 0 {
								x--
								y--
							}
						case '8':
							if y != 0 {
								y--
							}
						case '9':
							if y != 0 && x < width-1 {
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
		worldMap.MoveCreature(&player, x, y)

	}

}
