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
	for _, attr := range worldmap.Attributes {
		attributes[attr] = 8
	}
	name := message.RequestInput("Who are you?")
	for !creationComplete {
		printCreationScreen(attributes, worldmap.Attributes[currentSelection])
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
			if attributes[worldmap.Attributes[currentSelection]] > 0 {
				attributes[worldmap.Attributes[currentSelection]]--
			}
		case ui.Right:
			attributes[worldmap.Attributes[currentSelection]]++
		case ui.CreationDone:
			creationComplete = true
		}
	}
	ui.ClearScreen()

	return newPlayer(name, attributes)
}

func printCreationScreen(attributes map[string]int, selection string) {
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
}
