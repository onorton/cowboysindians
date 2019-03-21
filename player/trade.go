package player

import (
	"fmt"

	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/npc"
	"github.com/onorton/cowboysindians/ui"
)

func trade(p *Player, npc *npc.Npc) {
	tradeComplete := false
	for !tradeComplete {
		printTradeScreen(p, npc)
		npcItems := npc.GetItems()
		action := ui.GetInput()
		if action == ui.Buy {
			validSelection := false
			for !validSelection {
				message.PrintMessage("Buy: ")
				_, selection := ui.GetItemSelection()
				item := npcItems[selection]
				if item != nil {
					validSelection = true
				}
				if item[0].GetValue() > p.money {
					ui.WriteText(0, 0, "You don't have enough money for that!")
				} else {
					p.money -= item[0].GetValue()
					npc.AddMoney(item[0].GetValue())
					p.AddItem(item[0])
					message.Enqueue(fmt.Sprintf("You bought a %s.", item[0].GetName()))
					npc.RemoveItem(item[0])
				}
			}

		} else if action == ui.Sell {
			message.PrintMessage("Sell: ")
		} else if action == ui.Exit || action == ui.CancelAction {
			tradeComplete = true
			message.PrintMessage("")
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
		ui.WriteText(0, padding+i, fmt.Sprintf("%s %dx %s $%.2f", string(c), len(items), items[0].GetName(), float64(items[0].GetValue())/100))
		i++
	}

	i = 0
	for c, items := range npc.GetItems() {
		ui.WriteText(npcX, padding+i, fmt.Sprintf("%s %dx %s $%.2f", string(c), len(items), items[0].GetName(), float64(items[0].GetValue())/100))
		i++
	}

}
