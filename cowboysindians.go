package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/npc"
	"github.com/onorton/cowboysindians/player"
	"github.com/onorton/cowboysindians/ui"
	"github.com/onorton/cowboysindians/world"
	"github.com/onorton/cowboysindians/worldmap"

	"io/ioutil"
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
	Mounts      []*npc.Mount
	Enemies     []*npc.Enemy
	Npcs        []*npc.Npc
	Player      *player.Player
	Target      string
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

	npcsValue, err := json.Marshal(state.Npcs)
	check(err)
	buffer.WriteString(fmt.Sprintf("\"Npcs\":%s,\n", npcsValue))

	playerValue, err := json.Marshal(state.Player)
	check(err)
	buffer.WriteString(fmt.Sprintf("\"Player\":%s,\n", playerValue))

	targetValue, err := json.Marshal(state.Target)
	check(err)
	buffer.WriteString(fmt.Sprintf("\"Target\":%s\n", targetValue))

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
	state.Player.LoadMount(state.Mounts)

	for _, enemy := range state.Enemies {
		enemy.SetMap(state.Map)
		enemy.LoadMount(state.Mounts)
	}

	for _, mount := range state.Mounts {
		mount.SetMap(state.Map)
	}

	for _, npc := range state.Npcs {
		npc.SetMap(state.Map)
		npc.LoadMount(state.Mounts)
	}

	return state

}

// Combine enemies and player into same slice
func allCreatures(enemies []*npc.Enemy, mounts []*npc.Mount, npcs []*npc.Npc, p *player.Player) []worldmap.Creature {
	all := make([]worldmap.Creature, len(enemies)+len(mounts)+len(npcs)+1)
	i := 0
	for _, e := range enemies {
		all[i] = e
		i++
	}

	for _, m := range mounts {
		all[i] = m
		i++
	}

	for _, npc := range npcs {
		all[i] = npc
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

func printOpeningText(name string) {
	beginning := 4
	ui.WriteTextCentred(beginning, "You wake up bruised. You feel a dull pain in your head.")
	ui.WriteTextCentred(beginning+1, "It's beginning to come back to you now. They beat you. They tortured you.")
	ui.WriteTextCentred(beginning+2, "They took everything from you. Even your family. You don't remember where they went.")
	ui.WriteTextCentred(beginning+3, "But you remember their name...")
	ui.GetInput()
	ui.WriteTextCentred(beginning+5, name)
	ui.GetInput()
}

func main() {
	ui.Init(windowWidth)
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
		ui.ClearScreen()
	}

	if !loaded {
		m, mounts, enemies, npcs := world.GenerateWorld(width, height, windowWidth, windowHeight)
		state.Map = m
		state.Player = player.NewPlayer(state.Map)
		state.Mounts = mounts
		state.Enemies = enemies
		state.Npcs = npcs
		state.Time = 1
		state.PlayerIndex = 0
		target := npcs[rand.Intn(len(npcs))]
		state.Target = target.GetID()

		printOpeningText(target.FullName())
	}

	worldMap := state.Map
	player := state.Player
	mounts := state.Mounts
	enemies := state.Enemies
	npcs := state.Npcs

	all := allCreatures(enemies, mounts, npcs, player)
	for _, c := range all {
		x, y := c.GetCoordinates()
		worldMap.Move(c, x, y)
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
				message.PrintMessages()
				player.Update()

				// Game over, skip other enemies
				if player.IsDead() {
					break
				}

				for {
					worldMap.Render()
					stats := player.GetStats()
					stats = append([]string{fmt.Sprintf("T:%d", state.Time)}, stats...)
					printStatus(stats)
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
							action = ui.NoAction
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
						case ui.Talk:
							player.Talk()
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
				e := c.(*npc.Enemy)
				if e.IsDead() {
					continue
				}
				eX, eY := e.Update()
				worldMap.MoveCreature(e, eX, eY)
			} else if m, ok := c.(*npc.Mount); ok {
				if m.IsDead() {
					continue
				}

				mX, mY := m.Update()
				// If mounted, controlled by rider
				if !m.IsMounted() {
					worldMap.MoveCreature(m, mX, mY)
				}
			} else {
				npc := c.(*npc.Npc)
				if npc.IsDead() {
					continue
				}
				nX, nY := npc.Update()
				worldMap.MoveCreature(npc, nX, nY)
			}
		}

		// Remove dead enemies, npcs and mounts
		for i, c := range all {
			if e, ok := c.(*npc.Enemy); ok && e.IsDead() {
				e.EmptyInventory()
				worldMap.DeleteCreature(e)
				all = append(all[:i], all[i+1:]...)
			}
			if m, ok := c.(*npc.Mount); ok && m.IsDead() {
				worldMap.DeleteCreature(m)
				all = append(all[:i], all[i+1:]...)
			}
			if npc, ok := c.(*npc.Npc); ok && npc.IsDead() {
				npc.EmptyInventory()
				worldMap.DeleteCreature(npc)
				all = append(all[:i], all[i+1:]...)
				if npc.GetID() == state.Target {
					message.PrintMessage(fmt.Sprintf("%s is dead! You have been avenged.", npc.FullName()))
					ui.GetInput()
					quit = true
				}
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
