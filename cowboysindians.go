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
	player := (creature.Deserialize(items[1])).(*creature.Player)
	enemyStrings := strings.Split(items[2], "\n")
	enemyStrings = enemyStrings[0 : len(enemyStrings)-1]
	enemies := make([]*enemy.Enemy, len(enemyStrings))

	for i, e := range enemyStrings {
		enemies[i] = enemy.Deserialize(e).(*enemy.Enemy)
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
		enemies[i] = enemy.NewEnemy("bandit", x, y)
	}
	return enemies
}

// Combine enemies and player into same slice
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
	check(err)
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
		// Load from save file if player wants to
		if l.Type == termbox.EventKey && l.Ch == 'y' {
			worldMap, player, enemies, t, playerIndex = load()
		}

	}
	all := allCreatures(enemies, player)
	for _, c := range all {
		x, y := c.GetCoordinates()
		worldMap.MoveCreature(c, x, y)
	}
	inventory := false
	for {
		quit := false
		endTurn := false
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		x, y := player.GetCoordinates()

		// Sort by initiative order
		sort.Slice(all, func(i, j int) bool {
			return all[i].GetInitiative() > all[j].GetInitiative()

		})

		for i, c := range all {

			// Used when initially loading, to make sure faster enemies do not move twice
			if i < playerIndex {
				continue
			} else {
				playerIndex = 0
			}

			if _, ok := c.(*creature.Player); ok {
				// Only render when it is the player's turn
				worldMap.Render()
				message.PrintMessages()
				stats := player.GetStats()
				stats = append([]string{fmt.Sprintf("T:%d", t)}, stats...)
				printStatus(stats)
				// Game over, skip other enemies
				if player.IsDead() {
					break
				}
				for {
					if inventory {
						player.PrintInventory()
					}
					e := termbox.PollEvent()
					if e.Type == termbox.EventKey {
						playerMoved := e.Key == termbox.KeyArrowUp || e.Key == termbox.KeyArrowDown || e.Key == termbox.KeyArrowLeft || e.Key == termbox.KeyArrowRight || (e.Ch >= '1' && e.Ch <= '9')

						if player.OverEncumbered() && playerMoved {
							message.PrintMessage("You are too over encumbered to move.")
							continue
						}
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
								case ',':
									endTurn = worldMap.PickupItem()
								case 'd':
									endTurn = worldMap.DropItem()
								case 'i':
									if inventory {
										worldMap.Render()
									}
									inventory = !inventory
								case 'w':
									endTurn = player.WieldItem()
								case 'W':
									endTurn = player.WearArmour()
								default:
									quit = true
								}
							}
						}
						// End turn if player selects action that takes a turn

						endTurn = (endTurn || playerMoved)

						if endTurn || quit {
							break
						}
					} else {
						break
					}
				}

				worldMap.MovePlayer(player, x, y)

			} else {
				e := c.(*enemy.Enemy)
				if e.IsDead() {

					continue
				}
				eX, eY := e.Update(worldMap)
				worldMap.MoveCreature(e, eX, eY)
			}

		}

		// Remove dead enemies
		for i, c := range all {
			if e, ok := c.(*enemy.Enemy); c.IsDead() && ok {
				eX, eY := e.GetCoordinates()
				inventory := e.GetInventory()
				itemTypes := make(map[string]int)
				for _, item := range inventory {
					worldMap.PlaceItem(eX, eY, item)
					itemTypes[item.GetName()]++
				}
				if worldMap.IsVisible(player, eY, eY) {
					for name, count := range itemTypes {
						message.Enqueue(fmt.Sprintf("The enemy dropped %d %ss.", count, name))
					}
				}
				worldMap.DeleteCreature(c)
				all = append(all[:i], all[i+1:]...)
			}
		}
		// End game if player is dead
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
