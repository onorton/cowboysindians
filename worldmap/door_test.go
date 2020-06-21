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

type testKey struct {
	keyType int32
}

func (k testKey) KeyType() int32 {
	return k.keyType
}

func TestKeyFitsTrueKeyValueOfNegativeOne(t *testing.T) {
	key := testKey{-1}
	door := doorComponent{key: 10}
	if !door.KeyFits(key) {
		t.Error("Key should have fit the lock but does not")
	}
}

func TestKeyFitsTrueTypesMatch(t *testing.T) {
	key := testKey{10}
	door := doorComponent{key: 10}
	if !door.KeyFits(key) {
		t.Error("Key should have fit the lock but does not")
	}
}

func TestKeyFitsFalseTypesDoNotMatch(t *testing.T) {
	key := testKey{5}
	door := doorComponent{key: 10}
	if door.KeyFits(key) {
		t.Error("Key should not have fit the lock but does")
	}
}
