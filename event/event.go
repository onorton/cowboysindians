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
	item        *item.Item
	location    worldmap.Coordinates
}

type PickpocketEvent struct {
	id          string
	perpetrator worldmap.Creature
	item        *item.Item
	location    worldmap.Coordinates
}

type AttackEvent struct {
	id          string
	perpetrator worldmap.Creature
	victim      worldmap.Creature
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

func (e PickpocketEvent) Id() string {
	return e.id
}

func (e PickpocketEvent) Perpetrator() string {
	return e.perpetrator.GetID()
}

func (e PickpocketEvent) PerpetratorName() string {
	return e.perpetrator.GetName().FullName()
}

func (e PickpocketEvent) Crime() string {
	return "Pickpocketing"
}

func (e PickpocketEvent) Value() int {
	return 2 * e.item.GetValue()
}

func (e PickpocketEvent) Witness(world *worldmap.Map, c worldmap.Creature) {
	if e.Perpetrator() != c.GetID() && world.IsVisible(c, e.location.X, e.location.Y) {
		Emit(WitnessedCrimeEvent{e})
	}
}

func (e PickpocketEvent) Location() worldmap.Coordinates {
	return e.location
}

func (e AttackEvent) Id() string {
	return e.id
}

func (e AttackEvent) Perpetrator() worldmap.Creature {
	return e.perpetrator
}

func (e AttackEvent) Victim() worldmap.Creature {
	return e.victim
}

func (e AttackEvent) Location() worldmap.Coordinates {
	return e.location
}

func NewMurder(perpetrator worldmap.Creature, victim worldmap.Creature, location worldmap.Coordinates) MurderEvent {
	return MurderEvent{xid.New().String(), perpetrator, victim, location}
}

func NewTheft(perpetrator worldmap.Creature, item *item.Item, location worldmap.Coordinates) TheftEvent {
	return TheftEvent{xid.New().String(), perpetrator, item, location}
}

func NewPickpocket(perpetrator worldmap.Creature, item *item.Item, location worldmap.Coordinates) PickpocketEvent {
	return PickpocketEvent{xid.New().String(), perpetrator, item, location}
}

func NewAttack(perpetrator worldmap.Creature, victim worldmap.Creature) AttackEvent {
	vX, vY := victim.GetCoordinates()
	return AttackEvent{xid.New().String(), perpetrator, victim, worldmap.Coordinates{vX, vY}}
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
