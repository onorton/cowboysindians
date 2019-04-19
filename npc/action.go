package npc

import (
	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/worldmap"
)

type Action interface {
	execute()
}

type PickupAction struct {
	h     holdsItems
	world *worldmap.Map
	x, y  int
}

func (a PickupAction) execute() {
	items := a.world.GetItems(a.x, a.y)
	for _, item := range items {
		a.h.PickupItem(item)
	}
}

type DropAction struct {
	h    holdsItems
	item item.Item
}

func (a DropAction) execute() {
	a.h.dropItem(a.item)
}

type MountAction struct {
	r     Rider
	world *worldmap.Map
	x, y  int
}

func (a MountAction) execute() {
	m := a.world.GetCreature(a.x, a.y).(*Mount)
	a.world.DeleteCreature(m)
	m.AddRider(a.r)
	a.r.AddMount(m)
	// Move rider to position of mount
	MoveAction{a.r.(hasAi), a.world, a.x, a.y}.execute()
}

type HealAction struct {
	c   hasAi
	con *item.Consumable
}

func (a HealAction) execute() {
	a.c.heal(a.con.GetEffect("hp"))
	a.c.(holdsItems).RemoveItem(a.con)
}

type OpenAction struct {
	world *worldmap.Map
	x, y  int
}

func (a OpenAction) execute() {
	a.world.ToggleDoor(a.x, a.y, true)
}

type RangedAttackAction struct {
	c     hasAi
	world *worldmap.Map
	t     worldmap.Creature
}

func (a RangedAttackAction) execute() {
	itemUser := a.c.(usesItems)
	itemUser.Weapon().Fire()
	coverPenalty := 0
	if a.world.TargetBehindCover(a.c, a.t) {
		coverPenalty = 5
	}
	itemUser.rangedAttack(a.t, -coverPenalty)
}

type LoadAction struct {
	u usesItems
}

func (a LoadAction) execute() {
	for !a.u.weaponFullyLoaded() && a.u.hasAmmo() {
		a.u.getAmmo()
		a.u.Weapon().Load()
	}
}

type MoveAction struct {
	h     hasAi
	world *worldmap.Map
	x, y  int
}

func (a MoveAction) execute() {
	c := a.h.(worldmap.Creature)
	a.world.MoveCreature(c, a.x, a.y)
}

type MountedMoveAction struct {
	r     Rider
	world *worldmap.Map
	x, y  int
}

func (a MountedMoveAction) execute() {
	c := a.r.(worldmap.Creature)
	a.r.Mount().Move()
	a.world.MoveCreature(c, a.x, a.y)
}

type CrouchAction struct {
	c worldmap.CanCrouch
}

func (a CrouchAction) execute() {
	a.c.Crouch()
}

type StandupAction struct {
	c worldmap.CanCrouch
}

func (a StandupAction) execute() {
	a.c.Standup()
}

type NoAction struct{}

func (a NoAction) execute() {}