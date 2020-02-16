package player

import (
	"fmt"

	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/npc"
	"github.com/onorton/cowboysindians/ui"
	"github.com/onorton/cowboysindians/worldmap"
)

func trade(p *Player, npc *npc.Npc) {
	tradeComplete := false
	for !tradeComplete {
		printTradeScreen(p, npc)
		message.PrintMessages()
		npcItems := npc.GetItems(false)
		action := ui.GetInput()
		if action == ui.Buy {
			validSelection := false
			for !validSelection {
				message.PrintMessage("Buy: ")
				command, selection := ui.GetItemSelection()

				if command == ui.Cancel {
					break
				}

				item := npcItems[selection]
				if item != nil {
					validSelection = true
					value := item[0].GetValue()
					if p.hasSkill(worldmap.Haggling) {
						value -= value / 5
					}

					if value > p.money {
						message.Enqueue("You don't have enough money for that!")
					} else {
						p.money -= value
						npc.AddMoney(value)
						p.AddItem(item[0])
						message.Enqueue(fmt.Sprintf("You bought a %s.", item[0].GetName()))
						npc.RemoveItem(item[0])
					}
				}
			}
		} else if action == ui.Sell {
			validSelection := false
			for !validSelection {
				message.PrintMessage("Sell: ")
				command, selection := ui.GetItemSelection()
				item := p.GetItem(selection)

				if command == ui.Cancel {
					break
				}

				if item != nil {
					value := item.GetValue()
					if p.hasSkill(worldmap.Haggling) {
						value += value / 5
					}

					validSelection = true
					if !npc.CanAfford(value) {
						p.AddItem(item)
						message.Enqueue(fmt.Sprintf("%s cannot afford that!", npc.GetName()))
					} else {
						p.money += value
						npc.RemoveMoney(value)
						message.Enqueue(fmt.Sprintf("You sold a %s.", item.GetName()))
						npc.PickupItem(item)
					}
				}

			}
		} else if action == ui.Exit || action == ui.CancelAction {
			tradeComplete = true
			message.PrintMessage("\"Pleasure doing business with you.\"")
		}
	}
}

func printTradeScreen(p *Player, npc *npc.Npc) {
	ui.ClearScreen()
	padding := 2
	npcX := 50

	ui.WriteText(0, 0, "You:")
	ui.WriteText(npcX, 0, fmt.Sprintf("%s:", npc.GetName()))

	i := 0
	for c, items := range p.inventory {
		value := items[0].GetValue()
		if p.hasSkill(worldmap.Haggling) {
			value += value / 5
		}
		ui.WriteText(0, padding+i, fmt.Sprintf("%s %dx %s $%.2f", string(c), len(items), items[0].GetName(), float64(value)/100))
		i++
	}

	i = 0
	for c, items := range npc.GetItems(false) {
		value := items[0].GetValue()
		if p.hasSkill(worldmap.Haggling) {
			value -= value / 5
		}
		ui.WriteText(npcX, padding+i, fmt.Sprintf("%s %dx %s $%.2f", string(c), len(items), items[0].GetName(), float64(value)/100))
		i++
	}

}
