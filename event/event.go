package event

import (
	"github.com/onorton/cowboysindians/worldmap"
	"github.com/rs/xid"
)

var subscribers = make([]subscriber, 0)

type Event interface{}

type CrimeEvent struct {
	Id          string
	Perpetrator worldmap.Creature
	Victim      worldmap.Creature
	Location    worldmap.Coordinates
	Crime       string
}

func NewCrime(perpetrator worldmap.Creature, victim worldmap.Creature, location worldmap.Coordinates, crime string) CrimeEvent {
	return CrimeEvent{xid.New().String(), perpetrator, victim, location, crime}
}

type WitnessedCrimeEvent struct {
	Crime CrimeEvent
}

func Emit(e Event) {
	for _, subscriber := range subscribers {
		subscriber.ProcessEvent(e)
	}
}

func Subscribe(s subscriber) {
	subscribers = append(subscribers, s)
}

type subscriber interface {
	ProcessEvent(Event)
}
