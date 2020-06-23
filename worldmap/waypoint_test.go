package worldmap

import (
	"math"
	"testing"
)

type TestMap struct {
	passable [][]bool
}

func (tm TestMap) IsValid(x, y int) bool {
	return x >= 0 && x < len(tm.passable[0]) && y >= 0 && y < len(tm.passable)
}

func (tm TestMap) IsPassable(x, y int) bool {
	return tm.passable[y][x]
}

func TestRandomWaypointNextWaypointReturnsCurrentWaypointWhenNotReached(t *testing.T) {
	grid := make([][]bool, 10)
	for i := range grid {
		grid[i] = make([]bool, 10)
	}

	grid[3][2] = true
	grid[4][3] = true
	currentWaypoint := Coordinates{3, 4}
	rw := RandomWaypoint{world: TestMap{grid}, currentWaypoint: currentWaypoint}
	nextWaypoint := rw.NextWaypoint(Coordinates{2, 2})
	if nextWaypoint != currentWaypoint {
		t.Errorf("Expected NextWaypoint to return the current waypoint %v but returned %v", currentWaypoint, nextWaypoint)
	}
}

func TestRandomWaypointNextWaypointReturnsWaypointWithinFiveTilesVerticalAndHorizontal(t *testing.T) {
	grid := make([][]bool, 10)
	for i := range grid {
		grid[i] = make([]bool, 10)
	}
	currentWaypoint := Coordinates{1, 0}

	// Too far away
	grid[5][7] = true
	// Close enough
	grid[3][2] = true
	expectedWaypoint := Coordinates{2, 3}

	rw := RandomWaypoint{world: TestMap{grid}, currentWaypoint: currentWaypoint}
	nextWaypoint := rw.NextWaypoint(currentWaypoint)

	if math.Abs(float64(currentWaypoint.X-nextWaypoint.X)) > 5 || math.Abs(float64(currentWaypoint.Y-nextWaypoint.Y)) > 5 {
		t.Errorf("NextWaypoint %v is too far away from %v", nextWaypoint, currentWaypoint)
	}

	if nextWaypoint != expectedWaypoint {
		t.Errorf("Expected NextWaypoint to return the expected waypoint %v but returned %v", expectedWaypoint, nextWaypoint)
	}

	if rw.currentWaypoint != nextWaypoint {
		t.Error("Expected currentWaypoint to be set to nextWaypoint but it was not")
	}
}

func TestRandomWaypointNextWaypointReturnsValidAndPassableWaypoint(t *testing.T) {
	grid := make([][]bool, 10)
	for i := range grid {
		grid[i] = make([]bool, 10)
	}
	currentWaypoint := Coordinates{1, 0}

	grid[0][1] = true
	grid[3][2] = true

	world := TestMap{grid}
	rw := RandomWaypoint{world: world, currentWaypoint: currentWaypoint}
	nextWaypoint := rw.NextWaypoint(currentWaypoint)

	if !world.IsPassable(nextWaypoint.X, nextWaypoint.Y) {
		t.Error("Expected NextWaypoint to be passable but it was not")
	}
}
