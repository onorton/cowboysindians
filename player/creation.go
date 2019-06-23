package player

import (
	"fmt"

	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/ui"
	"github.com/onorton/cowboysindians/worldmap"
)

func CreatePlayer() *Player {
	creationComplete := false
	attributes := make(map[string]int)
	currentSelection := 0
	pointsAvailable := 8
	for _, attr := range worldmap.Attributes {
		attributes[attr] = 8
	}
	name := message.RequestInput("Who are you?")
	for !creationComplete {
		printCreationScreen(attributes, pointsAvailable, worldmap.Attributes[currentSelection])
		action := ui.CreationInput()

		switch action {
		case ui.Up:
			if currentSelection > 0 {
				currentSelection--
			}
		case ui.Down:
			if currentSelection < len(worldmap.Attributes)-1 {
				currentSelection++
			}
		case ui.Left:
			cost := pointsCost(attributes[worldmap.Attributes[currentSelection]], false)
			if attributes[worldmap.Attributes[currentSelection]] > 3 {
				attributes[worldmap.Attributes[currentSelection]]--
				pointsAvailable += cost
			}
		case ui.Right:
			cost := pointsCost(attributes[worldmap.Attributes[currentSelection]], true)
			if pointsAvailable >= cost && attributes[worldmap.Attributes[currentSelection]] < 18 {
				attributes[worldmap.Attributes[currentSelection]]++
				pointsAvailable -= cost
			}
		case ui.CreationDone:
			creationComplete = true
		}
	}
	ui.ClearScreen()

	return newPlayer(name, attributes)
}

func pointsCost(currentValue int, increase bool) int {
	if !increase {
		currentValue -= 1
	}

	if currentValue < 18 && currentValue >= 16 || currentValue >= 3 && currentValue < 5 {
		return 3
	} else if currentValue < 16 && currentValue >= 13 || currentValue >= 5 && currentValue < 8 {
		return 2
	} else {
		return 1
	}
}

func printCreationScreen(attributes map[string]int, pointsAvailable int, selection string) {
	ui.ClearScreen()

	padding := 2
	ui.WriteText(0, padding, "Attributes:")
	padding += 2

	i := 0
	for _, attr := range worldmap.Attributes {
		if attr == selection {
			ui.WriteHightlightedText(0, padding+i, fmt.Sprintf("%s: %d", attr, attributes[attr]))
		} else {
			ui.WriteText(0, padding+i, fmt.Sprintf("%s: %d", attr, attributes[attr]))
		}
		i++
	}

	ui.WriteText(0, padding+len(worldmap.Attributes)+2, fmt.Sprintf("Points Available: %d", pointsAvailable))

}
