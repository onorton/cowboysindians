package worldmap

import (
	"testing"
)

func TestNewBuildingCorrectlySetupsArea(t *testing.T) {
	x1, y1, x2, y2 := 5, 4, 10, 20
	b := NewBuilding(x1, y1, x2, y2, Residential)
	
	if b.Area.X1() != x1 {
		t.Errorf("Expected x1 of building to be %d but was %d", x1, b.Area.X1())
	}

	if b.Area.Y1() != y1 {
		t.Errorf("Expected y1 of building to be %d but was %d", y1, b.Area.Y1())
	}

	if b.Area.X2() != x2 {
		t.Errorf("Expected x2 of building to be %d but was %d", x2, b.Area.X2())
	}

	if b.Area.Y2() != y2 {
		t.Errorf("Expected y2 of building to be %d but was %d", y2, b.Area.Y1())
	}
}

type buildingInsidePair struct {
	x, y int
	inside bool
}

func TestBuildingInside(t *testing.T) {
	b := NewBuilding(5, 4, 10, 20, Residential)

	buildingInsideTests := []buildingInsidePair {
		{5, 4, true},
		{10, 20, true},
		{7, 4, true},
		{10, 15, true},
		{7, 3, false},
		{7, 21, false},
		{-5, -10, false},
		{4, 15, false},
		{11, 15, false},
	}

	for _, test := range buildingInsideTests {
		if b.Inside(test.x, test.y) != test.inside {
			t.Errorf("Expected inside for the building %v to be %t for point %d, %d", b, test.inside, test.x, test.y)
		} 
	}

}
