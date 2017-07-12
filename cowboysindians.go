package main

import (
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/creature"
	"github.com/onorton/cowboysindians/enemy"
	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/worldmap"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
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
func save(m worldmap.Map, p *creature.Player, enemies []*enemy.Enemy) {
	data := m.Serialize()
	data += "\n\n" + p.Serialize()
	data += "\n\n"
	for _, e := range enemies {
		data += e.Serialize() + "\n"
	}
	err := ioutil.WriteFile(saveFilename, []byte(data), 0644)
	check(err)
}

func load() (worldmap.Map, *creature.Player, []*enemy.Enemy) {

	data, err := ioutil.ReadFile(saveFilename)
	check(err)
	items := strings.Split(string(data), "\n\n")
	player := (*creature.Deserialize(items[1])).(*creature.Player)
	enemyStrings := strings.Split(items[2], "\n")
	enemyStrings = enemyStrings[0 : len(enemyStrings)-1]
	enemies := make([]*enemy.Enemy, len(enemyStrings))

	for i, e := range enemyStrings {
		//fmt.Println(enemy, i)
		enemies[i] = (*enemy.Deserialize(e)).(*enemy.Enemy)
	}
	return worldmap.DeserializeMap(items[0]), player, enemies

}

func generateEnemies(m worldmap.Map, p *creature.Player, n int) []*enemy.Enemy {
	enemies := make([]*enemy.Enemy, n)
	for i := 0; i < n; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		pX, pY := p.GetCoordinates()
		if !m.IsPassable(x, y) || (x == pX && y == pY) || m.IsOccupied(x, y) {
			i--
			continue
		}
		enemies[i] = enemy.NewEnemy(x, y, 'b', termbox.ColorBlue)
	}
	return enemies
}

func printTime(t int) {
	time := fmt.Sprintf("T:%d", t)
	for i, c := range time {
		termbox.SetCell(i, windowHeight+1, c, termbox.ColorWhite, termbox.ColorDefault)
	}
	termbox.Flush()
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
	enemies := generateEnemies(worldMap, player, 2)
	t := 1
	if _, err := os.Stat(saveFilename); !os.IsNotExist(err) {
		message.PrintMessage("Do you wish to load the last save? [yn]")
		l := termbox.PollEvent()
		if l.Type == termbox.EventKey && l.Ch == 'y' {
			worldMap, player, enemies = load()
		}

	}
	enemies = generateEnemies(worldMap, player, 2)
	x, y := player.GetCoordinates()
	worldMap.MovePlayer(player, x, y)
	for _, e := range enemies {
		x, y = e.GetCoordinates()
		worldMap.MoveCreature(e, x, y)
	}
	for {
		quit := false
		endTurn := false
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		worldMap.Render()
		x, y = player.GetCoordinates()
		message.Enqueue(fmt.Sprintf("%d %d", x, y))
		message.PrintMessages()

		printTime(t)

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
								save(worldMap, player, enemies)
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
		worldMap.MovePlayer(player, x, y)
		for _, enemy := range enemies {
			eX, eY := enemy.Update(worldMap)
			worldMap.MoveCreature(enemy, eX, eY)
		}
		t++
	}

}
