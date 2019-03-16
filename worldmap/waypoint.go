package worldmap

import (
	"math/rand"
)

// Random waypoint selects a random waypoint up to 5 tiles apart
type RandomWaypoint struct {
	world           *Map
	currentWaypoint Coordinates
}

func NewRandomWaypoint(world *Map, location Coordinates) RandomWaypoint {
	return RandomWaypoint{world, location}
}

func (w RandomWaypoint) NextWaypoint(location Coordinates) Coordinates {
	if w.currentWaypoint == location {
		for {
			newX := location.X + rand.Intn(11) - 5
			newY := location.Y + rand.Intn(11) - 5
			if w.world.IsValid(newX, newY) && w.world.IsPassable(newX, newY) {
				w.currentWaypoint = Coordinates{newX, newY}
				break
			}
		}
	}
	return w.currentWaypoint
}

type WaypointSystem interface {
	NextWaypoint(Coordinates) Coordinates
}
