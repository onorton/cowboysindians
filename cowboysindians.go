package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/onorton/cowboysindians/enemy"
	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/mount"
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
const width = 100
const height = 100
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
	Mounts      []*mount.Mount
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

	mountsValue, err := json.Marshal(state.Mounts)
	check(err)
	buffer.WriteString(fmt.Sprintf("\"Mounts\":%s,\n", mountsValue))

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

func generateMounts(m *worldmap.Map, n int) []*mount.Mount {
	mounts := make([]*mount.Mount, n)
	for i := 0; i < n; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		if !m.IsPassable(x, y) || m.IsOccupied(x, y) {
			i--
			continue
		}
		mounts[i] = mount.NewMount("horse", x, y, m)
	}
	return mounts
}

func generateEnemies(m *worldmap.Map, n int) []*enemy.Enemy {
	enemies := make([]*enemy.Enemy, n)
	for i := 0; i < n; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		if !m.IsPassable(x, y) || m.IsOccupied(x, y) {
			i--
			continue
		}
		enemies[i] = enemy.NewEnemy("bandit", x, y, m)
	}
	return enemies
}

// Combine enemies and player into same slice
func allCreatures(enemies []*enemy.Enemy, mounts []*mount.Mount, p *player.Player) []worldmap.Creature {
	all := make([]worldmap.Creature, len(enemies)+len(mounts)+1)
	i := 0
	for _, e := range enemies {
		all[i] = e
		i++
	}

	for _, m := range mounts {
		all[i] = m
		i++
	}

	all[len(all)-1] = p
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
		state.Mounts = generateMounts(state.Map, 5)
		state.Enemies = generateEnemies(state.Map, 2)
		state.Time = 1
		state.PlayerIndex = 0
	}

	worldMap := state.Map
	player := state.Player
	mounts := state.Mounts
	enemies := state.Enemies

	all := allCreatures(enemies, mounts, player)
	for _, c := range all {
		x, y := c.GetCoordinates()
		worldMap.MoveCreature(c, x, y)
	}

	// Initial action is nothing
	action := ui.NoAction
	inventory := false
	for {
		quit := false
		endTurn := false

		ui.ClearScreen()

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
					worldMap.Render()
					if inventory {
						player.PrintInventory()
					}
					if action == ui.NoAction {
						action = ui.GetInput()
					}

					playerMoved := action.IsMovementAction()

					if playerMoved {

						if player.OverEncumbered() {
							message.PrintMessage("You are too encumbered to move.")
							continue
						} else {
							endTurn, action = player.Move(action)
						}

					} else {
						switch action {
						case ui.PrintMessages:
							message.PrintMessages()
						case ui.Exit:
							message.PrintMessage("Do you wish to save? [yn]")

							if quitAction := ui.GetInput(); quitAction == ui.Confirm {
								save(state)
							}
							quit = true
						case ui.Wait:
							endTurn = true
						case ui.CloseDoor:
							endTurn = player.ToggleDoor(false)
						case ui.OpenDoor:
							endTurn = player.ToggleDoor(true)
						case ui.ToggleCrouch:
							endTurn = player.ToggleCrouch()
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
						case ui.Mount:
							endTurn = player.ToggleMount()
						default:
							quit = true
						}
						action = ui.NoAction
					}

					if endTurn || quit {
						break
					}
				}
			} else if c.GetAlignment() == worldmap.Enemy {
				e := c.(*enemy.Enemy)
				if e.IsDead() {
					continue
				}
				eX, eY := e.Update()
				worldMap.MoveCreature(e, eX, eY)
			} else {
				m := c.(*mount.Mount)
				if m.IsDead() {
					continue
				}

				mX, mY := m.Update()
				// If mounted, controlled by rider
				if !m.IsMounted() {
					worldMap.MoveCreature(m, mX, mY)
				}
			}
		}

		// Remove dead enemies and mounts
		for i, c := range all {
			if e, ok := c.(*enemy.Enemy); ok && e.IsDead() {
				e.EmptyInventory()
				worldMap.DeleteCreature(c)
				all = append(all[:i], all[i+1:]...)
			}
			if m, ok := c.(*mount.Mount); ok && m.IsDead() {
				worldMap.DeleteCreature(m)
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
