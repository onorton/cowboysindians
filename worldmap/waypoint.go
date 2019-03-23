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

func (r *RandomWaypoint) NextWaypoint(location Coordinates) Coordinates {
	if r.currentWaypoint == location {
		for {
			newX := location.X + rand.Intn(11) - 5
			newY := location.Y + rand.Intn(11) - 5
			if r.world.IsValid(newX, newY) && r.world.IsPassable(newX, newY) {
				r.currentWaypoint = Coordinates{newX, newY}
				break
			}
		}
	}
	return r.currentWaypoint
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

// WithinBuilding means that creatures move randomly within a given building
type WithinBuilding struct {
	world           *Map
	building        Building
	currentWaypoint Coordinates
}

func NewWithinBuilding(world *Map, building Building, location Coordinates) *WithinBuilding {
	return &WithinBuilding{world, building, location}
}

func (wb *WithinBuilding) SetMap(world *Map) {
	wb.world = world
}

func (wb *WithinBuilding) NextWaypoint(location Coordinates) Coordinates {
	if wb.currentWaypoint == location {
		for {
			newX := wb.building.X1 + rand.Intn(wb.building.X2-wb.building.X1-1) + 1
			newY := wb.building.Y1 + rand.Intn(wb.building.Y2-wb.building.Y1-1) + 1
			if wb.world.IsPassable(newX, newY) {
				wb.currentWaypoint = Coordinates{newX, newY}
				break
			}
		}
	}
	return wb.currentWaypoint
}

func (wb *WithinBuilding) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	buildingValue, err := json.Marshal(wb.building)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Building\":%s,", buildingValue))

	currentWaypointValue, err := json.Marshal(wb.currentWaypoint)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"CurrentWaypoint\":%s", currentWaypointValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (wb *WithinBuilding) UnmarshalJSON(data []byte) error {

	type wbJson struct {
		Building        Building
		CurrentWaypoint Coordinates
	}

	var v wbJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	wb.building = v.Building
	wb.currentWaypoint = v.CurrentWaypoint

	return nil
}

func UnmarshalWaypointSystem(waypoint map[string]interface{}) WaypointSystem {
	waypointJson, err := json.Marshal(waypoint)
	check(err)

	if _, ok := waypoint["Building"]; ok {
		var wb WithinBuilding
		err = json.Unmarshal(waypointJson, &wb)
		check(err)
		return &wb
	}

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
