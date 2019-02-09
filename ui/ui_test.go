package ui

import "testing"

type testpair struct {
	action           PlayerAction
	isMovementAction bool
}

var tests = []testpair{
	{MoveNorth, true},
	{MoveSouth, true},
	{MoveSouthEast, true},
	{MoveNorthWest, true},
	{OpenDoor, false},
	{Exit, false},
	{PrintMessages, false},
}

func TestIsMovementAction(t *testing.T) {
	for _, pair := range tests {
		v := pair.action.IsMovementAction()
		if v != pair.isMovementAction {
			t.Error(
				"For", pair.action,
				"expected", pair.isMovementAction,
				"got", v,
			)
		}
	}
}
