package event

import "github.com/onorton/cowboysindians/worldmap"

var subscribers = make([]subscriber, 0)

type CrimeEvent struct {
	Perpetrator worldmap.Creature
	Location    worldmap.Coordinates
	Crime       string
}

func Emit(e CrimeEvent) {
	for _, subscriber := range subscribers {
		subscriber.ProcessEvent(e)
	}
}

func Subscribe(s subscriber) {
	subscribers = append(subscribers, s)
}

type subscriber interface {
	ProcessEvent(CrimeEvent)
}
