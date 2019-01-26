package ui

import (
	termbox "github.com/nsf/termbox-go"
)

// PlayerAction is a type that represents player actions
type PlayerAction int

// IsMovementAction returns true if it is one of the movement actions, false otherwise.
func (a PlayerAction) IsMovementAction() bool {
	return a < PrintMessages
}

// ItemSelection is a type that represents what item players have selected
type ItemSelection int

const (
	MoveNorth PlayerAction = iota
	MoveSouth
	MoveWest
	MoveEast
	MoveNorthWest
	MoveNorthEast
	MoveSouthWest
	MoveSouthEast
	PrintMessages
	Exit
	Wait
	CloseDoor
	OpenDoor
	RangedAttack
	PickUpItem
	DropItem
	ToggleInventory
	WieldItem
	WieldArmour
	Consume
	Confirm
)

const (
	All ItemSelection = iota
	AllRelevant
	Cancel
	SpecificItem
)

// Init initialises the termbox instance
func Init() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
}

// Close closes the termbox instance
func Close() {
	termbox.Close()
}

// GetInput waits for the user to enter a key.
// Returns the action corresponding to the key entered.
func GetInput() (action PlayerAction) {
	e := termbox.PollEvent()

	switch e.Key {
	case termbox.KeyArrowLeft:
		action = MoveWest
	case termbox.KeyArrowRight:
		action = MoveEast
	case termbox.KeyArrowUp:
		action = MoveNorth
	case termbox.KeyArrowDown:
		action = MoveSouth
	case termbox.KeySpace:
		action = PrintMessages
	case termbox.KeyEsc:
		action = Exit
	default:
		{
			switch e.Ch {
			case '1':
				action = MoveNorthWest
			case '2':
				action = MoveSouth
			case '3':
				action = MoveSouthEast
			case '4':
				action = MoveWest
			case '5':
				action = Wait
			case '6':
				action = MoveEast
			case '7':
				action = MoveNorthWest
			case '8':
				action = MoveNorth
			case '9':
				action = MoveNorthEast
			case 'c':
				action = CloseDoor
			case 'o':
				action = OpenDoor
			case 't':
				action = RangedAttack
			case ',':
				action = PickUpItem
			case 'd':
				action = DropItem
			case 'i':
				action = ToggleInventory
			case 'w':
				action = WieldItem
			case 'W':
				action = WieldArmour
			case 'e':
				action = Consume
			case 'y':
				action = Confirm
			default:
				action = Exit
			}
		}
	}
	return action
}

// GetItemSelection returns a rune corresponding to the item that is selected.
func GetItemSelection() (ItemSelection, rune) {
	e := termbox.PollEvent()

	if e.Key == termbox.KeyEnter {
		return Cancel, 0
	} else if e.Ch == '*' {
		return All, 0
	} else if e.Ch == '?' {
		return AllRelevant, 0
	} else {
		return SpecificItem, e.Ch
	}
}
