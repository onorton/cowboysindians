package main

import (
	"fmt"

	termbox "github.com/nsf/termbox-go"
	"github.com/onorton/cowboysindians/creature"
	"github.com/onorton/cowboysindians/enemy"
	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/ui"
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
	ui.Init()
	defer ui.Close()
	message.SetWindowSize(windowWidth, windowHeight)
	worldMap := worldmap.NewMap(width, height, windowWidth, windowHeight)
	player := creature.NewPlayer()
	enemies := generateEnemies(worldMap, player, 2)
	t := 1
	playerIndex := 0
	if _, err := os.Stat(saveFilename); !os.IsNotExist(err) {
		message.PrintMessage("Do you wish to load the last save? [yn]")
		// Load from save file if player wants to
		if l := ui.GetInput(); l == ui.Confirm {
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
				player.Update()
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
					action := ui.GetInput()

					playerMoved := action.IsMovementAction()

					if player.OverEncumbered() && playerMoved {
					}

					if playerMoved {

						if player.OverEncumbered() {
							message.PrintMessage("You are too encumbered to move.")
							continue
						} else {
							worldMap.MovePlayer(player, action)
						}
					} else {
						switch action {
						case ui.PrintMessages:
							{
								message.PrintMessages()
							}
						case ui.Exit:
							message.PrintMessage("Do you wish to save? [yn]")

							if quitAction := ui.GetInput(); quitAction == ui.Confirm {
								save(worldMap, player, enemies, t, i)
							}
							quit = true

						case ui.CloseDoor:
							endTurn = worldMap.ToggleDoor(x, y, false)
						case ui.OpenDoor:
							endTurn = worldMap.ToggleDoor(x, y, true)
						case ui.RangedAttack:
							target := worldMap.FindTarget(player)
							if target != nil {
								player.RangedAttack(target)
								endTurn = true
							}
						case ui.PickUpItem:
							endTurn = worldMap.PickupItem()
						case ui.DropItem:
							endTurn = worldMap.DropItem()
						case ui.ToggleInventory:
							termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
							worldMap.Render()
							inventory = !inventory
						case ui.WieldItem:
							endTurn = player.WieldItem()
						case ui.WieldArmour:
							endTurn = player.WearArmour()
						case ui.Consume:
							endTurn = player.ConsumeItem()
						default:
							quit = true
						}
					}
					// End turn if player selects action that takes a turn
					endTurn = (endTurn || playerMoved)

					if endTurn || quit {
						break
					}
				}
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
				e.EmptyInventory(worldMap)
				worldMap.DeleteCreature(c)
				all = append(all[:i], all[i+1:]...)
			}
		}
		// End game if player is dead
		if player.IsDead() {
			message.PrintMessage("You died.")
			ui.GetInput()
			break
		}
		if quit {
			break
		}
		t++
	}
}
