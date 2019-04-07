package player

import (
	"fmt"

	"github.com/onorton/cowboysindians/npc"
	"github.com/onorton/cowboysindians/ui"
)

func claimBounties(p *Player, npc *npc.Npc) {
	dialogueComplete := false
	for !dialogueComplete {
		printBountyScreen(npc.GetBounties())
		action := ui.GetInput()
		if action == ui.Exit || action == ui.CancelAction {
			dialogueComplete = true
		}
	}
}

func printBountyScreen(bounties npc.Bounties) {
	ui.ClearScreen()
	padding := 2

	ui.WriteText(0, 0, "Bounties")

	for i, bounty := range bounties {
		ui.WriteText(0, padding+i, fmt.Sprintf("%d. %s", i+1, bounty))
	}
}
