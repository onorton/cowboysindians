package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

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
const width = 1024
const height = 1024
const saveFilename = "game.json"
const worldSaveFilename = "game_world.json"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type GameState struct {
	PlayerIndex int
	Time        int
	Viewer      *worldmap.Viewer
	Mounts      []*npc.Mount
	Npcs        []*npc.Npc
	Player      *player.Player
	Target      string
}

func save(state GameState, m *worldmap.Map) {
	m.SaveChunks()

	buffer := bytes.NewBufferString("{")

	buffer.WriteString(fmt.Sprintf("\"PlayerIndex\":%d,\n", state.PlayerIndex))
	buffer.WriteString(fmt.Sprintf("\"Time\":%d,\n", state.Time))

	viewerValue, err := json.Marshal(state.Viewer)
	check(err)
	buffer.WriteString(fmt.Sprintf("\"Viewer\":%s,\n", viewerValue))

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

	state.Player.LoadMount(state.Mounts)

	for _, npc := range state.Npcs {
		npc.LoadMount(state.Mounts)
	}

	return state

}

// Combine enemies and player into same slice
func allCreatures(mounts []*npc.Mount, npcs []*npc.Npc, p *player.Player) []worldmap.Creature {
	all := make([]worldmap.Creature, len(mounts)+len(npcs)+1)
	i := 0

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
	rand.Seed(time.Now().UTC().UnixNano())

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
		p, mounts, npcs := world.GenerateWorld(worldSaveFilename, width, height)
		state.Player = p
		x, y := state.Player.GetCoordinates()
		state.Viewer = worldmap.NewViewer(x, y, windowWidth, windowHeight)
		state.Mounts = mounts
		state.Npcs = npcs
		state.Time = 1
		state.PlayerIndex = 0
		targets := make([]*npc.Npc, 0)
		for _, npc := range state.Npcs {
			if npc.Human() {
				targets = append(targets, npc)
			}
		}
		target := targets[rand.Intn(len(targets))]
		state.Target = target.GetID()

		printOpeningText(target.GetName().FullName())
	}

	player := state.Player
	mounts := state.Mounts
	npcs := state.Npcs

	all := allCreatures(mounts, npcs, player)
	worldMap := worldmap.NewMap(worldSaveFilename, width, height, state.Viewer, state.Player, all)
	worldMap.LoadActiveChunks()

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
								save(state, worldMap)
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
						case ui.Read:
							player.Read()
						case ui.Use:
							endTurn = player.Use()
						case ui.Pickpocket:
							endTurn = player.Pickpocket()
						}
						action = ui.NoAction
					}

					if endTurn || quit {
						break
					}
				}
			} else {
				if c.IsDead() {
					continue
				}
				cX, cY := c.GetCoordinates()
				if !worldMap.InActiveChunks(cX, cY) {
					continue
				}
				c.Update()
			}
		}

		// Remove dead enemies, npcs and mounts
		for i, c := range all {
			if m, ok := c.(*npc.Mount); ok && m.IsDead() {
				m.DropCorpse()
				worldMap.DeleteCreature(m)
				all = append(all[:i], all[i+1:]...)
			}
			if npc, ok := c.(*npc.Npc); ok && npc.IsDead() {
				npc.EmptyInventory()
				worldMap.DeleteCreature(npc)
				all = append(all[:i], all[i+1:]...)
				if npc.GetID() == state.Target {
					message.PrintMessage(fmt.Sprintf("%s is dead! You have been avenged.", npc.GetName().FullName()))
					ui.GetInput()
					quit = true
				}
			}

		}
		// End game if player is dead
		if player.IsDead() {
			message.PrintMessage("You died.")

			// Delete game files
			os.Remove(saveFilename)
			os.Remove(worldSaveFilename)

			ui.GetInput()
			break
		}

		if quit {
			break
		}
		state.Time++
	}
}
