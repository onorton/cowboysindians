package worldmap

import (
	"testing"

	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/ui"
)

func init() {
	terrainDataPath = "../data/terrain.json"
}

func TestNewWorldCoordinatesAreValidOrInvalid(t *testing.T) {
	width, height := 128, 64
	world := NewWorld(width, height)

	type coordinateTestPair struct {
		x, y  int
		valid bool
	}

	testCases := []coordinateTestPair{
		{10, 20, true},
		{10, 64, false},
		{-5, -4, false},
		{0, 0, true},
		{45, 63, true},
		{127, 50, true},
		{128, 20, false},
		{128, 67, false},
	}

	for _, testCase := range testCases {
		if world.IsValid(testCase.x, testCase.y) != testCase.valid {
			t.Errorf("Expected validity of coordinates %d, %d to be %t but was %t", testCase.x, testCase.y, testCase.valid, world.IsValid(testCase.x, testCase.y))
		}
	}
}

func TestNewWorldStartsWithAllGroundTilesAndCorrectSize(t *testing.T) {
	width, height := 128, 64
	world := NewWorld(width, height)

	if world.Width() != width {
		t.Errorf("Expected width to be %d but was %d", width, world.Width())
	}

	if world.Height() != height {
		t.Errorf("Expected height to be %d but was %d", height, world.Height())
	}

	for x := 0; x < world.Width(); x++ {
		for y := 0; y < world.Height(); y++ {

			if !world.IsValid(x, y) {
				t.Errorf("Expected tile at %d, %d to be valid but it wasn't", x, y)
			}

			if !world.IsPassable(x, y) {
				t.Errorf("Expected tile at %d, %d to be passable but it wasn't", x, y)
			}

			if world.IsOccupied(x, y) {
				t.Errorf("Expected tile at %d, %d to not be passable but it was", x, y)
			}

			if world.Door(x, y) != nil {
				t.Errorf("Expected tile at %d, %d not to have a door but it did", x, y)
			}
		}
	}
}

func TestNewTileForWorldAddsDifferentTileToLocation(t *testing.T) {
	width, height := 128, 64
	world := NewWorld(width, height)

	type tileTestPair struct {
		tileType          string
		passable, hasDoor bool
	}

	testCases := []tileTestPair{
		{"wall", false, false},
		{"window", false, false},
		{"door", false, true},
		{"counter flap", false, true},
		{"counter", false, false},
		{"path", true, false},
		{"ground", true, false},
	}

	for _, testCase := range testCases {
		world.NewTile(testCase.tileType, 1, 1)
		if world.IsPassable(1, 1) != testCase.passable {
			t.Errorf("Expected passability of tile %s to be %t but was %t", testCase.tileType, testCase.passable, world.IsPassable(1, 1))
		}

		hasDoor := world.Door(1, 1) != nil
		if hasDoor != testCase.hasDoor {
			t.Errorf("Expected door value of tile %s to be %t but was %t", testCase.tileType, testCase.hasDoor, hasDoor)
		}

		if world.IsOccupied(1, 1) {
			t.Error("Expected tile not be occupied but it was")
		}
	}
}

func TestPlacePlacesCreatureInWorld(t *testing.T) {
	width, height := 128, 64
	world := NewWorld(width, height)

	cX, cY := 10, 20
	c := &testCreature{cX, cY}
	world.Place(c)

	if !world.IsOccupied(cX, cY) {
		t.Errorf("Expected tile at %d, %d to have a creature but it did not", cX, cY)
	}
}

type testCreature struct {
	x, y int
}

func (c *testCreature) Render() ui.Element                                  { return ui.Element{} }
func (c *testCreature) GetInitiative() int                                  { return 0 }
func (c *testCreature) MeleeAttack(cr Creature)                             {}
func (c *testCreature) TakeDamage(d item.Damage, e item.Effects, bonus int) {}
func (c *testCreature) IsDead() bool                                        { return false }
func (c *testCreature) IsCrouching() bool                                   { return false }
func (c *testCreature) AttackHits(int) bool                                 { return false }
func (c *testCreature) GetName() ui.Name                                    { return ui.PlainName{} }
func (c *testCreature) GetAlignment() Alignment                             { return Neutral }
func (c *testCreature) Update()                                             {}
func (c *testCreature) GetID() string                                       { return "" }
func (c *testCreature) SetMap(m *Map)                                       {}
func (c testCreature) GetCoordinates() (int, int)                           { return c.x, c.y }
func (c *testCreature) SetCoordinates(x, y int)                             { c.x = x; c.y = y }
func (c *testCreature) GetVisionDistance() int                              { return 0 }
func (c *testCreature) Standup()                                            {}
func (c *testCreature) Crouch()                                             {}
