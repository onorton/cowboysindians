package main

import (
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/creature"
	"github.com/onorton/cowboysindians/structs"
	"github.com/onorton/cowboysindians/worldmap"
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
	print_tb(0, height, termbox.ColorWhite, termbox.ColorDefault, m)
	termbox.Flush()
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	messages := new(structs.Queue)

	worldMap := worldmap.NewMap(width, height)
	player := creature.NewPlayer()

	for {
		worldMap.MoveCreature(&player)
		quit := false
		endTurn := false
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		worldMap.Render()
		//messages.Enqueue(fmt.Sprintf("%d %d", x, y))
		print_message(messages)
		for {
			e := termbox.PollEvent()
			if e.Type == termbox.EventKey {
				switch e.Key {
				case termbox.KeyArrowLeft:
					if player.X != 0 {
						player.X--
					}
				case termbox.KeyArrowRight:
					if player.X < width-1 {
						player.X++
					}
				case termbox.KeyArrowUp:
					if player.Y != 0 {
						player.Y--
					}
				case termbox.KeyArrowDown:
					if player.Y < height-1 {
						player.Y++
					}
				case termbox.KeySpace:
					{
						print_message(messages)
					}
				default:
					{

						switch e.Ch {
						case '1':
							if player.X != 0 && player.Y < height-1 {
								player.X--
								player.Y++
							}
						case '2':
							if player.Y < height-1 {
								player.Y++
							}
						case '3':
							if player.X < width-1 && player.Y < height-1 {
								player.X++
								player.Y++
							}

						case '4':
							if player.X != 0 {
								player.X--
							}
						case '5':
						case '6':
							if player.X < width-1 {
								player.X++
							}
						case '7':
							if player.X != 0 && player.X != 0 {
								player.X--
								player.Y--
							}
						case '8':
							if player.Y != 0 {
								player.Y--
							}
						case '9':
							if player.Y != 0 && player.X < width-1 {
								player.Y--
								player.X++
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
