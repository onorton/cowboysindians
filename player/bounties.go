package player

import (
	"fmt"

	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/npc"
	"github.com/onorton/cowboysindians/ui"
)

func claimBounties(p *Player, npc *npc.Npc) {
	dialogueComplete := false
	collectedBounty := false
	for !dialogueComplete {
		printBountyScreen(npc.GetBounties())
		message.PrintMessages()
		action := ui.GetBountyInput()
		if action == ui.Claim {
			totalReward := 0
			for _, items := range p.inventory {
				for _, itm := range items {
					if c, ok := itm.(*item.NormalItem); ok && c.IsCorpse() {
						reward, criminal := npc.GetBounties().RemoveBounty(c.Owner())
						if reward > 0 {
							totalReward += reward
							message.Enqueue(fmt.Sprintf("You managed to track down %s. Your reward is $%.2f.", criminal, float64(reward)/100))
						}
					}
				}
			}
			if totalReward == 0 {
				message.PrintMessage(fmt.Sprintf("You have no bounties to claim here."))
			} else {
				p.money += totalReward
				collectedBounty = true
			}

		} else if action == ui.Exit {
			dialogueComplete = true
			if collectedBounty {
				message.PrintMessage("Thanks for helping out!")
			} else {
				message.PrintMessage("If you see any of those scoundrels, let me know.")
			}
		}
	}
}

func printBountyScreen(bounties *npc.Bounties) {
	ui.ClearScreen()
	padding := 2

	ui.WriteText(0, 0, "Bounties")

	for i, bounty := range bounties.Bounties() {
		ui.WriteText(0, padding+i, fmt.Sprintf("%d. %s", i+1, bounty))
	}
}
