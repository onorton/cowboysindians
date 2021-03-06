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
			if s.index < len(worldmap.Attributes)-1 {
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
			if s.index < len(skillsInfo)-1 {
				s.index++
			} else {
				s.selection = completion
			}
		case ui.Up:
			if s.index > -1 {
				s.index--
			}
		}
	} else {
		switch action {
		case ui.Up:
			s.selection = attribute
			s.index = -1
		}
	}
}

type skillInformation struct {
	skillName   string
	skill       worldmap.Skill
	description string
}

var skillsInfo []skillInformation = []skillInformation{
	skillInformation{"Unarmed", worldmap.Unarmed, "Proficiency with your fists."},
	skillInformation{"Melee", worldmap.Melee, "Proficiency with melee weapons."},
	skillInformation{"Archery", worldmap.Archery, "Proficiency with bows."},
	skillInformation{"Shotguns", worldmap.Shotguns, "Proficiency with shotguns."},
	skillInformation{"Rifles", worldmap.Rifles, "Proficiency with rifles."},
	skillInformation{"Pistols", worldmap.Pistols, "Proficiency with pistols."},
	skillInformation{"Double Shot", worldmap.DoubleShot, "Fire twice with the same weapon per turn."},
	skillInformation{"Dual Wielding", worldmap.DualWielding, "Can use two ranged weapons per turn."},
	skillInformation{"Haggling", worldmap.Haggling, "Get better deals with merchants."},
	skillInformation{"Lockpicking", worldmap.Lockpicking, "Increased chance of lockpicks working."},
	skillInformation{"Pickpocketing", worldmap.Pickpocketing, "Reduced chance of being detected while pickpocketing."}}

func CreatePlayer(location worldmap.Coordinates) *Player {
	creationComplete := false
	attributes := make(map[string]int)
	currentSelection := &selection{attribute, -1}
	skillsAvailable := 3
	selectedSkills := structs.Initialise()
	pointsAvailable := 8
	for _, attr := range worldmap.Attributes {
		attributes[attr] = 8
	}
	name := message.RequestInput("Who are you?")
	for !creationComplete {
		printCreationScreen(attributes, *selectedSkills, pointsAvailable, skillsAvailable, *currentSelection)
		action := ui.CreationInput()
		currentSelection.next(action)

		if currentSelection.selection == attribute && currentSelection.index >= 0 {
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
		} else if currentSelection.selection == skill && currentSelection.index >= 0 {
			if action == ui.Select && currentSelection.index >= 0 {
				if selectedSkills.Exists(skillsInfo[currentSelection.index].skill) {
					selectedSkills.Delete(skillsInfo[currentSelection.index].skill)
					skillsAvailable++
				} else if skillsAvailable > 0 {
					selectedSkills.Add(skillsInfo[currentSelection.index].skill)
					skillsAvailable--
				}
			}
		} else if currentSelection.selection == completion {
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

	return newPlayer(location, name, attributes, skills)
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

func printCreationScreen(attributes map[string]int, selectedSkills structs.Set, pointsAvailable int, skillsAvailable int, s selection) {
	ui.ClearScreen()

	// Print skill description if skill selected
	if s.selection == skill && s.index >= 0 {
		message.PrintMessage(skillsInfo[s.index].description)
	}

	skillsOffset := 50
	padding := 2

	if s.selection == attribute && s.index == -1 {
		ui.WriteHighlightedText(0, padding, "Attributes:")
	} else {
		ui.WriteText(0, padding, "Attributes:")
	}
	padding += 2
	ui.WriteText(0, padding, fmt.Sprintf("Points Available: %d", pointsAvailable))
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

	padding = 2
	if s.selection == skill && s.index == -1 {
		ui.WriteHighlightedText(skillsOffset, padding, "Skills:")
	} else {
		ui.WriteText(skillsOffset, padding, "Skills:")
	}
	padding += 2
	ui.WriteText(skillsOffset, padding, fmt.Sprintf("Skills Available: %d", skillsAvailable))
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
		ui.WriteHighlightedText(0, padding+len(skillsInfo)+2, "Complete")
	} else {
		ui.WriteText(0, padding+len(skillsInfo)+2, "Complete")
	}

}
