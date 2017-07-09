package main

import (
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/creature"
	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/worldmap"
	"io/ioutil"
	"os"
)

const windowWidth = 100
const windowHeight = 25
const width = 10
const height = 10
const saveFilename = "game.dat"

func check(e error) {
	if e != nil {
		panic(e)
	}
}
func save(m worldmap.Map, p *creature.Player) {
	data := m.Serialize()
	data += "\n\n" + p.Serialize()
	err := ioutil.WriteFile(saveFilename, []byte(data), 0644)
	check(err)
}

func load() worldmap.Map {

	dat, err := ioutil.ReadFile(saveFilename)
	check(err)
	return worldmap.DeserializeMap(string(dat))

}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	message.SetWindowSize(windowWidth, windowHeight)
	worldMap := worldmap.NewMap(width, height, windowWidth, windowHeight)
	player := creature.NewPlayer()
	if _, err := os.Stat(saveFilename); !os.IsNotExist(err) {
		message.PrintMessage("Do you wish to load the last save? [yn]")
		l := termbox.PollEvent()
		if l.Type == termbox.EventKey && l.Ch == 'y' {
			worldMap = load()
		}

	}
	for {

		quit := false
		endTurn := false
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		worldMap.Render()
		x := player.X
		y := player.Y
		message.Enqueue(fmt.Sprintf("%d %d", x, y))
		message.PrintMessages()

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
						message.PrintMessages()
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
						case 'c':
							endTurn = worldMap.ToggleDoor(x, y, false)
						case 'o':
							endTurn = worldMap.ToggleDoor(x, y, true)
						case 'q':

							message.PrintMessage("Do you wish to save? [yn]")
							quitEvent := termbox.PollEvent()
							if quitEvent.Type == termbox.EventKey && quitEvent.Ch == 'y' {
								save(worldMap)
							}
							quit = true
						default:
							quit = true
						}
					}
				}
				endTurn = endTurn || (e.Key != termbox.KeySpace && e.Ch != 'c' && e.Ch != 'o' && e.Ch != 'q')
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
