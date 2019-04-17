package event

import (
	"math/rand"

	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/worldmap"
	"github.com/rs/xid"
)

var subscribers = make([]subscriber, 0)

type Event interface{}

type CrimeEvent interface {
	Id() string
	Perpetrator() string
	PerpetratorName() string
	Location() worldmap.Coordinates
	Crime() string
	Value() int
	Witness(*worldmap.Map, worldmap.Creature)
}

type WitnessedCrimeEvent struct {
	Crime CrimeEvent
}

type MurderEvent struct {
	id          string
	perpetrator worldmap.Creature
	victim      worldmap.Creature
	location    worldmap.Coordinates
}

type TheftEvent struct {
	id          string
	perpetrator worldmap.Creature
	item        item.Item
	location    worldmap.Coordinates
}

func (e MurderEvent) Id() string {
	return e.id
}

func (e MurderEvent) Perpetrator() string {
	return e.perpetrator.GetID()
}

func (e MurderEvent) PerpetratorName() string {
	return e.perpetrator.GetName().FullName()
}

func (e MurderEvent) Location() worldmap.Coordinates {
	return e.location
}

func (e MurderEvent) Crime() string {
	return "Murder"
}

func (e MurderEvent) Value() int {
	return (1 + rand.Intn(10)) * 10000
}

func (e MurderEvent) Witness(world *worldmap.Map, c worldmap.Creature) {
	if e.Perpetrator() != c.GetID() && e.victim.GetID() != c.GetID() && world.IsVisible(c, e.location.X, e.location.Y) {
		Emit(WitnessedCrimeEvent{e})
	}
}

func (e TheftEvent) Id() string {
	return e.id
}

func (e TheftEvent) Perpetrator() string {
	return e.perpetrator.GetID()
}

func (e TheftEvent) PerpetratorName() string {
	return e.perpetrator.GetName().FullName()
}

func (e TheftEvent) Crime() string {
	return "Theft"
}

func (e TheftEvent) Value() int {
	return 2 * e.item.GetValue()
}

func (e TheftEvent) Witness(world *worldmap.Map, c worldmap.Creature) {
	if e.Perpetrator() != c.GetID() && world.IsVisible(c, e.location.X, e.location.Y) {
		Emit(WitnessedCrimeEvent{e})
	}
}

func (e TheftEvent) Location() worldmap.Coordinates {
	return e.location
}

func NewMurder(perpetrator worldmap.Creature, victim worldmap.Creature, location worldmap.Coordinates) MurderEvent {
	return MurderEvent{xid.New().String(), perpetrator, victim, location}
}

func NewTheft(perpetrator worldmap.Creature, item item.Item, location worldmap.Coordinates) TheftEvent {
	return TheftEvent{xid.New().String(), perpetrator, item, location}
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
