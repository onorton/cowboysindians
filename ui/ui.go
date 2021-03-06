package ui

import (
	termbox "github.com/nsf/termbox-go"
)

// Inputs

// PlayerAction is a type that represents player actions
type PlayerAction int

// IsMovementAction returns true if it is one of the movement actions, false otherwise.
func (a PlayerAction) IsMovementAction() bool {
	return a < PrintMessages
}

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
	ToggleCrouch
	RangedAttack
	PickUpItem
	DropItem
	ToggleInventory
	WieldItem
	WieldArmour
	LoadWeapon
	Consume
	Mount
	Talk
	Buy
	Sell
	Claim
	Read
	Use
	Pickpocket
	Place
	Primary
	Secondary
	Confirm
	CancelAction
	NoAction
)

// ItemSelection is a type that represents what item players have selected
type ItemSelection int

var centre int

const (
	All ItemSelection = iota
	AllRelevant
	Cancel
	SpecificItem
)

type TextInputAction int

const (
	Erase TextInputAction = iota
	Character
	OtherInput
	Done
)

type CreationAction int

const (
	Invalid CreationAction = iota
	Up
	Down
	Left
	Right
	Select
)

// Init initialises the termbox instance
func Init(width int) {
	centre = width / 2
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetOutputMode(termbox.Output256)
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
	case termbox.KeyCtrlC:
		action = Talk
	case termbox.KeyEsc:
		action = Exit
	case termbox.KeyEnter:
		action = CancelAction
	default:
		{

			switch e.Ch {
			case '1':
				action = MoveSouthWest
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
			case 'C':
				action = ToggleCrouch
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
			case 'l':
				action = LoadWeapon
			case 'e':
				action = Consume
			case 'm':
				action = Mount
			case 'b':
				action = Buy
			case 's':
				action = Sell
			case 'r':
				action = Read
			case 'a':
				action = Use
			case 'p':
				action = Pickpocket
			case 'P':
				action = Place
			case 'y':
				action = Confirm
			case 'n':
				action = CancelAction
			default:
				action = NoAction
			}
		}
	}
	return action
}

func GetBountyInput() (action PlayerAction) {
	e := termbox.PollEvent()

	switch e.Key {
	case termbox.KeyEsc:
		action = Exit
	case termbox.KeyEnter:
		action = Exit
	default:
		{
			if e.Ch == 'c' {
				action = Claim
			} else {
				action = NoAction
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

func EquippedSelection() PlayerAction {
	e := termbox.PollEvent()

	if e.Key == termbox.KeyEnter {
		return CancelAction
	} else if e.Ch == 'p' {
		return Primary
	} else if e.Ch == 's' {
		return Secondary
	}
	return NoAction
}

func TextInput() (TextInputAction, rune) {
	e := termbox.PollEvent()
	switch e.Key {
	case termbox.KeyEnter:
		return Done, 0
	case termbox.KeySpace:
		return Character, ' '
	case termbox.KeyBackspace | termbox.KeyBackspace2:
		return Erase, 0
	default:
		if e.Ch != 0 {
			return Character, e.Ch
		}
		return OtherInput, 0
	}
}

func CreationInput() CreationAction {
	e := termbox.PollEvent()
	switch e.Key {
	case termbox.KeyArrowUp:
		return Up
	case termbox.KeyArrowDown:
		return Down
	case termbox.KeyArrowLeft:
		return Left
	case termbox.KeyArrowRight:
		return Right
	case termbox.KeyEnter:
		return Select
	default:
		return Invalid
	}
}

// Rendering
type Cell struct {
	x int
	y int
}

type Element struct {
	char   rune
	colour termbox.Attribute
	bg     termbox.Attribute
}

func NewCell(x int, y int) Cell {
	return Cell{x, y}
}

func NewElement(char rune, colour termbox.Attribute) Element {
	return Element{char, colour, termbox.ColorDefault}
}

func NewElementWithBg(char rune, colour, bg termbox.Attribute) Element {
	return Element{char, colour, bg}
}

func EmptyElement() Element {
	return Element{' ', termbox.ColorDefault, termbox.ColorDefault}
}

func ClearCells(cells []Cell) {
	for _, cell := range cells {
		termbox.SetCell(cell.x, cell.y, ' ', termbox.ColorDefault, termbox.ColorDefault)
	}
	termbox.Flush()
}

func ClearScreen() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

func DrawElement(x, y int, elem Element) {
	termbox.SetCell(x, y, elem.char, elem.colour, elem.bg)
	termbox.Flush()
}

func WriteText(x, y int, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, termbox.ColorWhite, termbox.ColorDefault)
		x++
	}
	termbox.Flush()
}

func WriteHighlightedText(x, y int, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, termbox.ColorBlack, termbox.ColorWhite)
		x++
	}
	termbox.Flush()
}

func WriteTextCentred(y int, msg string) {
	x := centre - len(msg)/2
	for _, c := range msg {
		termbox.SetCell(x, y, c, termbox.ColorWhite, termbox.ColorDefault)
		x++
	}
	termbox.Flush()
}

func RenderGrid(x, y int, elems [][]Element) {
	currY := y
	for _, row := range elems {
		for i, elem := range row {
			termbox.SetCell(x+i, currY, elem.char, elem.colour, elem.bg)
		}
		currY++
	}
	termbox.Flush()

}

type Name interface {
	WithDefinite() string
	WithIndefinite() string
	FullName() string
	String() string
	PlayerKnows()
}

type PlainName struct {
	Name string
}

func (n PlainName) WithDefinite() string {
	return "the " + n.Name

}

func (n PlainName) WithIndefinite() string {
	return "a " + n.Name

}

func (n PlainName) String() string {
	return n.Name
}

func (n PlainName) FullName() string {
	return n.String()
}

func (n PlainName) PlayerKnows() {}
