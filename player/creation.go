package player

import (
	"fmt"

	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/structs"
	"github.com/onorton/cowboysindians/ui"
	"github.com/onorton/cowboysindians/worldmap"
)

type selectionType int

const (
	attribute selectionType = iota
	skill
	completion
)

type selection struct {
	selection selectionType
	index     int
}

func (s *selection) next(action ui.CreationAction) {
	if s.selection == attribute {
		switch action {
		case ui.Right:
			if s.index == -1 {
				s.selection = skill
			}
		case ui.Down:
			if s.index < len(worldmap.Attributes) {
				s.index++
			} else {
				s.selection = completion
			}
		case ui.Up:
			if s.index > -1 {
				s.index--
			}
		}
	} else if s.selection == skill {
		switch action {
		case ui.Left:
			if s.index == -1 {
				s.selection = attribute
			}
		case ui.Down:
			if s.index < len(skillsInfo) {
				s.index++
			} else {
				s.selection = completion
			}
		case ui.Up:
			if s.index > -1 {
				s.index--
			}
		}
	}
}

type skillInformation struct {
	skillName string
	skill     worldmap.Skill
}

var skillsInfo []skillInformation = []skillInformation{
	skillInformation{"Unarmed", worldmap.Unarmed},
	skillInformation{"Melee", worldmap.Melee},
	skillInformation{"Archery", worldmap.Archery},
	skillInformation{"Shotguns", worldmap.Shotguns},
	skillInformation{"Rifles", worldmap.Rifles},
	skillInformation{"Pistols", worldmap.Pistols},
	skillInformation{"Double Shot", worldmap.DoubleShot},
	skillInformation{"Dual Wielding", worldmap.DualWielding},
	skillInformation{"Haggling", worldmap.Haggling},
	skillInformation{"Lockpicking", worldmap.Lockpicking},
	skillInformation{"Pickpocketing", worldmap.Pickpocketing}}

func CreatePlayer() *Player {
	creationComplete := false
	attributes := make(map[string]int)
	currentSelection := &selection{attribute, -1}
	selectedSkills := structs.Initialise()
	pointsAvailable := 8
	for _, attr := range worldmap.Attributes {
		attributes[attr] = 8
	}
	name := message.RequestInput("Who are you?")
	for !creationComplete {
		printCreationScreen(attributes, *selectedSkills, pointsAvailable, *currentSelection)
		action := ui.CreationInput()
		currentSelection.next(action)

		if currentSelection.selection == attribute {
			switch action {
			case ui.Left:
				cost := pointsCost(attributes[worldmap.Attributes[currentSelection.index]], false)
				if attributes[worldmap.Attributes[currentSelection.index]] > 3 {
					attributes[worldmap.Attributes[currentSelection.index]]--
					pointsAvailable += cost
				}
			case ui.Right:
				cost := pointsCost(attributes[worldmap.Attributes[currentSelection.index]], true)
				if pointsAvailable >= cost && attributes[worldmap.Attributes[currentSelection.index]] < 18 {
					attributes[worldmap.Attributes[currentSelection.index]]++
					pointsAvailable -= cost
				}
			}
		} else if currentSelection.selection == skill {
			if action == ui.Select && currentSelection.index >= 0 {
				if selectedSkills.Exists(skillsInfo[currentSelection.index].skill) {
					selectedSkills.Delete(skillsInfo[currentSelection.index].skill)
				} else {
					selectedSkills.Add(skillsInfo[currentSelection.index].skill)
				}
			}
		} else {
			if action == ui.Select {
				creationComplete = true
			}
		}

	}
	ui.ClearScreen()

	skills := make([]worldmap.Skill, 0)
	for _, s := range selectedSkills.Items() {
		skills = append(skills, s.(worldmap.Skill))
	}

	return newPlayer(name, attributes, skills)
}

func pointsCost(currentValue int, increase bool) int {
	if !increase {
		currentValue--
	}

	if currentValue < 18 && currentValue >= 16 || currentValue >= 3 && currentValue < 5 {
		return 3
	} else if currentValue < 16 && currentValue >= 13 || currentValue >= 5 && currentValue < 8 {
		return 2
	} else {
		return 1
	}
}

func printCreationScreen(attributes map[string]int, selectedSkills structs.Set, pointsAvailable int, s selection) {
	ui.ClearScreen()
	skillsOffset := 50

	padding := 2

	if s.selection == attribute && s.index == -1 {
		ui.WriteHighlightedText(0, padding, "Attributes:")
	} else {
		ui.WriteText(0, padding, "Attributes:")
	}
	padding += 2

	i := 0
	for index, attr := range worldmap.Attributes {
		if s.selection == attribute && index == s.index {
			ui.WriteHighlightedText(0, padding+i, fmt.Sprintf("%s: %d", attr, attributes[attr]))
		} else {
			ui.WriteText(0, padding+i, fmt.Sprintf("%s: %d", attr, attributes[attr]))
		}
		i++
	}

	ui.WriteText(0, padding+len(worldmap.Attributes)+2, fmt.Sprintf("Points Available: %d", pointsAvailable))

	padding = 2
	if s.selection == skill && s.index == -1 {
		ui.WriteHighlightedText(skillsOffset, padding, "Skills:")
	} else {
		ui.WriteText(skillsOffset, padding, "Skills:")
	}
	padding += 2

	i = 0
	for index, sk := range skillsInfo {
		text := sk.skillName
		if selectedSkills.Exists(sk.skill) {
			ui.WriteText(skillsOffset-2, padding+i, "*")
		}

		if s.selection == skill && index == s.index {
			ui.WriteHighlightedText(skillsOffset, padding+i, fmt.Sprintf("%s", text))
		} else {
			ui.WriteText(skillsOffset, padding+i, fmt.Sprintf("%s", text))
		}
		i++
	}

	if s.selection == completion {
		ui.WriteHighlightedText(skillsOffset, padding+len(skillsInfo)+2, "Complete")
	} else {
		ui.WriteText(0, padding+len(skillsInfo)+2, "Complete")
	}

}
