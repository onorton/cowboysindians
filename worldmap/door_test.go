package worldmap

import (
	"testing"
)

func TestDoorToggleLockedInitiallyLocked(t *testing.T) {
	door := doorComponent{locked: true}
	door.ToggleLocked()
	if door.locked {
		t.Error("Door should have been unlocked but wasn't")
	}
}

func TestDoorToggleUnlockedInitiallyUnlocked(t *testing.T) {
	door := doorComponent{locked: false}
	door.ToggleLocked()
	if !door.locked {
		t.Error("Door should have been locked but wasn't")
	}
}

func TestDoorLock(t *testing.T) {
	for _, locked := range [2]bool{true, false} {
		door := doorComponent{locked: locked}
		door.Lock()
		if !door.locked {
			t.Errorf("Door locked was initially %t and is now unlocked but shouldn't be", locked)
		}
	}
}
