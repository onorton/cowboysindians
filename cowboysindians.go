package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/onorton/cowboysindians/enemy"
	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/player"
	"github.com/onorton/cowboysindians/ui"
	"github.com/onorton/cowboysindians/worldmap"

	"io/ioutil"
	"math/rand"
	"os"
	"sort"
)

const windowWidth = 100
const windowHeight = 25
const width = 20
const height = 20
const saveFilename = "game.json"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type GameState struct {
	PlayerIndex int
	Time        int
	Map         *worldmap.Map
	Enemies     []*enemy.Enemy
	Player      *player.Player
}

func save(state GameState) {
	buffer := bytes.NewBufferString("{")

	buffer.WriteString(fmt.Sprintf("\"PlayerIndex\":%d,\n", state.PlayerIndex))
	buffer.WriteString(fmt.Sprintf("\"Time\":%d,\n", state.Time))

	mapValue, err := json.Marshal(state.Map)
	check(err)
	buffer.WriteString(fmt.Sprintf("\"Map\":%s,\n", mapValue))

	enemiesValue, err := json.Marshal(state.Enemies)
	check(err)
	buffer.WriteString(fmt.Sprintf("\"Enemies\":%s,\n", enemiesValue))

	playerValue, err := json.Marshal(state.Player)
	check(err)
	buffer.WriteString(fmt.Sprintf("\"Player\":%s\n", playerValue))

	buffer.WriteString("}")

	err = ioutil.WriteFile(saveFilename, buffer.Bytes(), 0644)
	check(err)

}

func load() GameState {

	data, err := ioutil.ReadFile(saveFilename)
	check(err)
	state := GameState{}
	err = json.Unmarshal(data, &state)
	check(err)

	state.Player.SetMap(state.Map)

	for _, enemy := range state.Enemies {
		enemy.SetMap(state.Map)
	}

	return state

}

func generateEnemies(m *worldmap.Map, p *player.Player, n int) []*enemy.Enemy {
	enemies := make([]*enemy.Enemy, n)
	for i := 0; i < n; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		pX, pY := p.GetCoordinates()
		if !m.IsPassable(x, y) || (x == pX && y == pY) || m.IsOccupied(x, y) {
			i--
			continue
		}
		enemies[i] = enemy.NewEnemy("bandit", x, y, m)
	}
	return enemies
}

// Combine enemies and player into same slice
func allCreatures(enemies []*enemy.Enemy, p *player.Player) []worldmap.Creature {
	all := make([]worldmap.Creature, len(enemies)+1)
	for i, e := range enemies {
		all[i] = e
	}
	all[len(enemies)] = p
	return all
}

func printStatus(status []string) {
	statusString := ""
	for _, stat := range status {
		statusString += stat + " "
	}
	ui.WriteText(0, windowHeight+1, statusString)

}
func main() {
	ui.Init()
	defer ui.Close()
	item.LoadAllData()
	message.SetWindowSize(windowWidth, windowHeight)
	state := GameState{}

	loaded := false
	if _, err := os.Stat(saveFilename); !os.IsNotExist(err) {
		message.PrintMessage("Do you wish to load the last save? [yn]")
		// Load from save file if player wants to
		if l := ui.GetInput(); l == ui.Confirm {
			state = load()
			loaded = true
		}
	}

	if !loaded {
		state.Map = worldmap.NewMap(width, height, windowWidth, windowHeight)
		state.Player = player.NewPlayer(state.Map)
		state.Enemies = generateEnemies(state.Map, state.Player, 2)
		state.Time = 1
		state.PlayerIndex = 0
	}

	worldMap := state.Map
	player := state.Player
	enemies := state.Enemies

	all := allCreatures(enemies, player)
	for _, c := range all {
		x, y := c.GetCoordinates()
		worldMap.MoveCreature(c, x, y)
	}

	inventory := false
	for {
		quit := false
		endTurn := false
		ui.ClearScreen()
		x, y := player.GetCoordinates()

		// Sort by initiative order
		sort.Slice(all, func(i, j int) bool {
			return all[i].GetInitiative() > all[j].GetInitiative()
		})

		for i, c := range all {
			// Used when initially loading, to make sure faster enemies do not move twice
			if i < state.PlayerIndex {
				continue
			} else {
				state.PlayerIndex = 0
			}

			if c.GetAlignment() == worldmap.Player {
				// Only render when it is the player's turn
				worldMap.Render()
				message.PrintMessages()
				player.Update()
				stats := player.GetStats()
				stats = append([]string{fmt.Sprintf("T:%d", state.Time)}, stats...)
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
								save(state)
							}
							quit = true
						case ui.Wait:
							endTurn = true
						case ui.CloseDoor:
							endTurn = player.ToggleDoor(x, y, false)
						case ui.OpenDoor:
							endTurn = player.ToggleDoor(x, y, true)
						case ui.ToggleCrouch:
							player.ToggleCrouch()
							endTurn = true
						case ui.RangedAttack:
							endTurn = player.RangedAttack()
						case ui.PickUpItem:
							endTurn = player.PickupItem()
						case ui.DropItem:
							endTurn = player.DropItem()
						case ui.ToggleInventory:
							ui.ClearScreen()
							worldMap.Render()
							inventory = !inventory
						case ui.WieldItem:
							endTurn = player.WieldItem()
						case ui.WieldArmour:
							endTurn = player.WearArmour()
						case ui.LoadWeapon:
							endTurn = player.LoadWeapon()
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
				eX, eY := e.Update()
				worldMap.MoveCreature(e, eX, eY)
			}
		}

		// Remove dead enemies
		for i, c := range all {
			if e, ok := c.(*enemy.Enemy); c.IsDead() && ok {
				e.EmptyInventory()
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
		state.Time++
	}
}
