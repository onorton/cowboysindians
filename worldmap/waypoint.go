package worldmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
)

// Random waypoint selects a random waypoint up to 5 tiles apart
type RandomWaypoint struct {
	world           *Map
	currentWaypoint Coordinates
}

func NewRandomWaypoint(world *Map, location Coordinates) *RandomWaypoint {
	return &RandomWaypoint{world, location}
}

func (r *RandomWaypoint) SetMap(world *Map) {
	r.world = world
}

func (w *RandomWaypoint) NextWaypoint(location Coordinates) Coordinates {
	if w.currentWaypoint == location {
		for {
			newX := location.X + rand.Intn(11) - 5
			newY := location.Y + rand.Intn(11) - 5
			if w.world.IsValid(newX, newY) {
				w.currentWaypoint = Coordinates{newX, newY}
				break
			}
		}
	}
	return w.currentWaypoint
}

func (r *RandomWaypoint) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	currentWaypointValue, err := json.Marshal(r.currentWaypoint)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"CurrentWaypoint\":%s", currentWaypointValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (r *RandomWaypoint) UnmarshalJSON(data []byte) error {
	type randomJson struct {
		CurrentWaypoint Coordinates
	}

	var v randomJson
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	r.currentWaypoint = v.CurrentWaypoint
	return nil
}

// Patrol means that creatures move between defined waypoints in order
type Patrol struct {
	waypoints []Coordinates
	index     int
}

func NewPatrol(waypoints []Coordinates) *Patrol {
	return &Patrol{waypoints, 0}
}

func (p *Patrol) NextWaypoint(location Coordinates) Coordinates {
	if p.waypoints[p.index] == location {
		p.index = (p.index + 1) % len(p.waypoints)
	}
	return p.waypoints[p.index]
}

func (p *Patrol) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	waypointsValue, err := json.Marshal(p.waypoints)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Waypoints\":%s,", waypointsValue))

	buffer.WriteString(fmt.Sprintf("\"Index\":%d", p.index))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (p *Patrol) UnmarshalJSON(data []byte) error {

	type patrolJson struct {
		Waypoints []Coordinates
		Index     int
	}

	var v patrolJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	p.waypoints = v.Waypoints
	p.index = v.Index

	return nil
}

func UnmarshalWaypointSystem(waypoint map[string]interface{}) WaypointSystem {
	waypointJson, err := json.Marshal(waypoint)
	check(err)

	if _, ok := waypoint["Index"]; ok {
		var patrol Patrol
		err = json.Unmarshal(waypointJson, &patrol)
		check(err)
		return &patrol
	}
	var random RandomWaypoint
	err = json.Unmarshal(waypointJson, &random)
	check(err)
	return &random
}

type WaypointSystem interface {
	NextWaypoint(Coordinates) Coordinates
}
