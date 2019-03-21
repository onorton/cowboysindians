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
		action := ui.GetInput()
		if action == ui.Buy {
			message.PrintMessage("Buy: ")
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
