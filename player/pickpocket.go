package player

import (
	"fmt"
	"math/rand"

	"github.com/onorton/cowboysindians/event"
	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/npc"
	"github.com/onorton/cowboysindians/ui"
	"github.com/onorton/cowboysindians/worldmap"
)

func pickpocket(p *Player, npc *npc.Npc) {
	pickpocketComplete := false
	chanceCaught := 0.25
	if p.hasSkill(worldmap.Pickpocketing) {
		chanceCaught -= 0.2
	}

	for !pickpocketComplete {
		printPickpocketScreen(p, npc)
		message.PrintMessages()
		npcItems := npc.GetItems(true)
		action := ui.GetInput()
		if action == ui.Pickpocket {
			validSelection := false
			for !validSelection {
				message.PrintMessage("Take: ")
				command, selection := ui.GetItemSelection()
				item := npcItems[selection]

				if command == ui.Cancel {
					break
				}

				if item != nil {
					validSelection = true
					if item[0].GetName() == "money" {
						npc.RemoveMoney(item[0].GetValue())
						p.money += item[0].GetValue()
					} else {
						p.AddItem(item[0])
						npc.RemoveItem(item[0])
					}
					message.Enqueue(fmt.Sprintf("You took a %s.", item[0].GetName()))
					if rand.Float64() < chanceCaught {
						event.Emit(event.NewPickpocket(p, item[0], p.location))
						message.Enqueue("You've been caught!")
						return
					}
				}

			}

		} else if action == ui.Place {
			validSelection := false
			for !validSelection {
				message.PrintMessage("Place: ")
				command, selection := ui.GetItemSelection()

				if command == ui.Cancel {
					break
				}

				item := p.GetItem(selection)
				if item != nil {
					validSelection = true
					if item.GetName() == "money" {
						npc.AddMoney(item.GetValue())
						p.money -= item.GetValue()
					} else {
						npc.PickupItem(item)
					}
					message.Enqueue(fmt.Sprintf("You placed a %s on %s's person.", item.GetName(), npc.GetName().WithDefinite()))
					if rand.Float64() < chanceCaught {
						event.Emit(event.NewPickpocket(p, item, p.location))
						message.Enqueue("You've been caught!")
						return
					}
				}
			}
		} else if action == ui.Exit || action == ui.CancelAction {
			pickpocketComplete = true
		}
	}
}

func printPickpocketScreen(p *Player, npc *npc.Npc) {
	ui.ClearScreen()
	padding := 2
	npcX := 50

	ui.WriteText(0, 0, "You:")
	ui.WriteText(npcX, 0, fmt.Sprintf("%s:", npc.GetName()))

	playerInventory := p.inventory
	money := item.Money(p.money)
	playerInventory[money.GetKey()] = []*item.Item{money}
	i := 0
	for c, items := range p.inventory {
		ui.WriteText(0, padding+i, fmt.Sprintf("%s %dx %s $%.2f", string(c), len(items), items[0].GetName(), float64(items[0].GetValue())/100))
		i++
	}

	i = 0
	for c, items := range npc.GetItems(true) {
		ui.WriteText(npcX, padding+i, fmt.Sprintf("%s %dx %s $%.2f", string(c), len(items), items[0].GetName(), float64(items[0].GetValue())/100))
		i++
	}

}
