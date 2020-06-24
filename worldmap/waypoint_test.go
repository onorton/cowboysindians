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

	if !world.IsValid(nextWaypoint.X, nextWaypoint.Y) {
		t.Error("Expected NextWaypoint to be valid but it was not")
	}

	if !world.IsPassable(nextWaypoint.X, nextWaypoint.Y) {
		t.Error("Expected NextWaypoint to be passable but it was not")
	}
}

func TestPatrolNextWaypointReturnsCurrentWaypointWhenNotReached(t *testing.T) {
	currentWaypoint := Coordinates{1, 1}
	waypoints := []Coordinates{currentWaypoint, Coordinates{2, 5}, Coordinates{3, 3}}
	p := Patrol{waypoints: waypoints, index: 0}
	nextWaypoint := p.NextWaypoint(Coordinates{2, 2})
	if nextWaypoint != currentWaypoint {
		t.Errorf("Expected NextWaypoint to return the current waypoint %v but returned %v", currentWaypoint, nextWaypoint)
	}
}

func TestPatrolNextWaypointReturnsNextWaypointInWaypointsArray(t *testing.T) {
	currentWaypoint := Coordinates{1, 1}
	expectedNextWaypoint := Coordinates{2, 5}
	waypoints := []Coordinates{currentWaypoint, expectedNextWaypoint, Coordinates{3, 3}}
	p := Patrol{waypoints: waypoints, index: 0}
	nextWaypoint := p.NextWaypoint(currentWaypoint)
	if nextWaypoint != expectedNextWaypoint {
		t.Errorf("Expected NextWaypoint to return the next waypoint in array %v but returned %v", expectedNextWaypoint, nextWaypoint)
	}
}

func TestPatrolNextWaypointLoopsBackAround(t *testing.T) {
	currentWaypoint := Coordinates{3, 3}
	expectedNextWaypoint := Coordinates{1, 1}
	waypoints := []Coordinates{expectedNextWaypoint, Coordinates{2, 5}, currentWaypoint}
	p := Patrol{waypoints: waypoints, index: 2}
	nextWaypoint := p.NextWaypoint(currentWaypoint)
	if nextWaypoint != expectedNextWaypoint {
		t.Errorf("Expected NextWaypoint to return the next waypoint in array %v but returned %v", expectedNextWaypoint, nextWaypoint)
	}
}

func TestWithinAreaNextWaypointReturnsCurrentWaypointWhenNotReached(t *testing.T) {
	grid := make([][]bool, 10)
	for i := range grid {
		grid[i] = make([]bool, 10)
	}

	grid[3][2] = true
	grid[4][3] = true

	area := Area{Coordinates{0, 0}, Coordinates{5, 5}}

	currentWaypoint := Coordinates{3, 4}
	wa := WithinArea{world: TestMap{grid}, area: area, currentWaypoint: currentWaypoint}
	nextWaypoint := wa.NextWaypoint(Coordinates{2, 2})
	if nextWaypoint != currentWaypoint {
		t.Errorf("Expected NextWaypoint to return the current waypoint %v but returned %v", currentWaypoint, nextWaypoint)
	}
}

func TestWithinAreaNextWaypointReturnsWaypointWithinAreaGiven(t *testing.T) {
	grid := make([][]bool, 10)
	for i := range grid {
		grid[i] = make([]bool, 10)
	}

	area := Area{Coordinates{0, 0}, Coordinates{5, 5}}

	// Outside area
	grid[5][7] = true
	// Within area
	grid[4][3] = true
	grid[2][4] = true

	world := TestMap{grid}
	currentWaypoint := Coordinates{1, 0}
	wa := WithinArea{world: world, area: area, currentWaypoint: currentWaypoint}
	nextWaypoint := wa.NextWaypoint(currentWaypoint)

	if !(nextWaypoint.X >= area.X1() && nextWaypoint.X <= area.X2() && nextWaypoint.Y >= area.Y1() && nextWaypoint.Y <= area.Y2()) {
		t.Errorf("Expected next waypoint %v to be in area %v but it was not", nextWaypoint, area)
	}

	if wa.currentWaypoint != nextWaypoint {
		t.Error("Expected currentWaypoint to be set to nextWaypoint but it was not")
	}
}

func TestWithinAreaNextWaypointReturnsValidAndPassableWaypoint(t *testing.T) {
	grid := make([][]bool, 10)
	for i := range grid {
		grid[i] = make([]bool, 10)
	}

	grid[0][1] = true
	grid[3][2] = true

	area := Area{Coordinates{0, 0}, Coordinates{5, 5}}
	world := TestMap{grid}
	currentWaypoint := Coordinates{1, 0}
	wa := WithinArea{world: world, area: area, currentWaypoint: currentWaypoint}

	nextWaypoint := wa.NextWaypoint(currentWaypoint)

	if !world.IsValid(nextWaypoint.X, nextWaypoint.Y) {
		t.Error("Expected NextWaypoint to be valid but it was not")
	}

	if !world.IsPassable(nextWaypoint.X, nextWaypoint.Y) {
		t.Error("Expected NextWaypoint to be passable but it was not")
	}
}
