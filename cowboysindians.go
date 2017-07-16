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
	"sort"
	"strconv"
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
func save(m worldmap.Map, p *creature.Player, enemies []*enemy.Enemy, t, playerIndex int) {
	data := fmt.Sprintf("%d %d\n", t, playerIndex)
	data += m.Serialize() + "\n\n"
	data += p.Serialize() + "\n\n"
	for _, e := range enemies {
		data += e.Serialize() + "\n"
	}
	err := ioutil.WriteFile(saveFilename, []byte(data), 0644)
	check(err)
}

func load() (worldmap.Map, *creature.Player, []*enemy.Enemy, int, int) {

	data, err := ioutil.ReadFile(saveFilename)
	check(err)
	items := strings.Split(string(data), "\n\n")
	player := (*creature.Deserialize(items[1])).(*creature.Player)
	enemyStrings := strings.Split(items[2], "\n")
	enemyStrings = enemyStrings[0 : len(enemyStrings)-1]
	enemies := make([]*enemy.Enemy, len(enemyStrings))

	for i, e := range enemyStrings {
		enemies[i] = (*enemy.Deserialize(e)).(*enemy.Enemy)
	}
	timeMap := strings.SplitN(items[0], "\n", 2)
	status := strings.Split(timeMap[0], " ")

	t, _ := strconv.Atoi(status[0])
	playerIndex, _ := strconv.Atoi(status[1])

	return worldmap.DeserializeMap(timeMap[1]), player, enemies, t, playerIndex

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

func allCreatures(enemies []*enemy.Enemy, p *creature.Player) []creature.Creature {
	all := make([]creature.Creature, len(enemies)+1)
	for i, e := range enemies {
		all[i] = e
	}
	all[len(enemies)] = p
	return all
}

func printStatus(status []string) {
	length := 0
	for _, stat := range status {
		for i, c := range stat {
			termbox.SetCell(length+i, windowHeight+1, c, termbox.ColorWhite, termbox.ColorDefault)
		}
		length += len(stat) + 1
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
	playerIndex := 0
	if _, err := os.Stat(saveFilename); !os.IsNotExist(err) {
		message.PrintMessage("Do you wish to load the last save? [yn]")
		l := termbox.PollEvent()
		if l.Type == termbox.EventKey && l.Ch == 'y' {
			worldMap, player, enemies, t, playerIndex = load()
		}

	}
	all := allCreatures(enemies, player)
	for _, c := range all {
		x, y := c.GetCoordinates()
		worldMap.MoveCreature(c, x, y)
	}
	for {
		quit := false
		endTurn := false
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		x, y := player.GetCoordinates()

		sort.Slice(all, func(i, j int) bool {
			return all[i].GetInitiative() > all[j].GetInitiative()

		})

		for i, c := range all {
			if i < playerIndex {
				continue
			} else {
				playerIndex = 0
			}

			if p, ok := c.(*creature.Player); ok {
				worldMap.Render()
				message.PrintMessages()
				status := make([]string, 2)
				status[0] = fmt.Sprintf("T:%d", t)
				status[1] = fmt.Sprintf("HP:%d", p.GetHP())
				printStatus(status)
				if p.IsDead() {
					continue
				}
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
						case termbox.KeyEsc:
							message.PrintMessage("Do you wish to save? [yn]")
							quitEvent := termbox.PollEvent()
							if quitEvent.Type == termbox.EventKey && quitEvent.Ch == 'y' {
								save(worldMap, player, enemies, t, i)
							}
							quit = true

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
								case 't':
									target := worldMap.FindTarget(player)
									if target != nil {
										player.RangedAttack(target)
										endTurn = true
									}
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

				worldMap.MovePlayer(p, x, y)

			} else {
				if c.IsDead() {
					continue
				}
				e := c.(*enemy.Enemy)
				eX, eY := e.Update(worldMap)
				worldMap.MoveCreature(e, eX, eY)
			}

		}

		for i, c := range all {
			if c.IsDead() {
				worldMap.DeleteCreature(c)
				all = append(all[:i], all[i+1:]...)
			}
		}
		if player.IsDead() {
			message.PrintMessage("You died.")
			termbox.PollEvent()
			break
		}
		if quit {

			break
		}
		t++
	}

}
