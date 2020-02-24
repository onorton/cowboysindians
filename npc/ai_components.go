package npc

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/onorton/cowboysindians/event"
	"github.com/onorton/cowboysindians/structs"
	"github.com/onorton/cowboysindians/worldmap"
)

type bountiesComponent struct {
	t        worldmap.Town
	bounties *Bounties
}

func (c bountiesComponent) ProcessEvent(e event.Event) {
	switch ev := e.(type) {
	case event.WitnessedCrimeEvent:
		{
			crime := ev.Crime
			location := crime.Location()
			if location.X >= c.t.TownArea.X1() && location.X <= c.t.TownArea.X2() && location.Y >= c.t.TownArea.Y1() && location.Y <= c.t.TownArea.Y2() {
				c.bounties.addBounty(crime)
			}
		}
	}

}

func (c bountiesComponent) targets(ai hasAi, world *worldmap.Map) []worldmap.Creature {
	d := ai.GetVisionDistance()
	aiX, aiY := ai.GetCoordinates()

	targets := make([]worldmap.Creature, 0)

	for i := -d; i < d+1; i++ {
		for j := -d; j < d+1; j++ {
			// Translate location into world coordinates
			wX, wY := aiX+j, aiY+i
			if world.IsValid(wX, wY) && world.IsVisible(ai, wX, wY) && world.GetCreature(wX, wY) != nil && c.bounties.hasBounty(world.GetCreature(wX, wY).GetID()) {
				targets = append(targets, world.GetCreature(wX, wY))
			}
		}
	}

	return targets
}

func (c bountiesComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	townValue, err := json.Marshal(c.t)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Town\":%s,", townValue))

	bountiesValue, err := json.Marshal(c.bounties)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Bounties\":%s", bountiesValue))

	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (c *bountiesComponent) UnmarshalJSON(data []byte) error {
	type bountiesJSON struct {
		Town     worldmap.Town
		Bounties *Bounties
	}

	var v bountiesJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	c.t = v.Town
	c.bounties = v.Bounties

	return nil
}

type threatsComponent struct {
	possibleThreats *structs.Set
	creatureID      string
}

func (c threatsComponent) ProcessEvent(e event.Event) {
	switch ev := e.(type) {
	case event.AttackEvent:
		{
			if ev.Victim().GetID() == c.creatureID {
				c.possibleThreats.Add(ev.Perpetrator().GetID())
			}
		}
	}

}

func (c threatsComponent) threats(ai hasAi, world *worldmap.Map) []worldmap.Creature {

	d := ai.GetVisionDistance()
	aiX, aiY := ai.GetCoordinates()

	visibleThreats := make([]worldmap.Creature, 0)

	for i := -d; i < d+1; i++ {
		for j := -d; j < d+1; j++ {
			// Translate location into world coordinates
			wX, wY := aiX+j, aiY+i
			if world.IsValid(wX, wY) && world.IsVisible(ai, wX, wY) && world.GetCreature(wX, wY) != nil && c.possibleThreats.Exists(world.GetCreature(wX, wY).GetID()) {
				visibleThreats = append(visibleThreats, world.GetCreature(wX, wY))
			}
		}
	}

	// Only consider visible threats
	c.possibleThreats = structs.Initialise()
	for _, t := range visibleThreats {
		c.possibleThreats.Add(t.GetID())
	}

	return visibleThreats
}

func (c threatsComponent) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	possibleThreatsValue, err := json.Marshal(c.possibleThreats.Items())
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"PossibleThreats\":%s,", possibleThreatsValue))

	creatureIDValue, err := json.Marshal(c.creatureID)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"CreatureId\":%s", creatureIDValue))

	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (c *threatsComponent) UnmarshalJSON(data []byte) error {
	type threatsJSON struct {
		PossibleThreats []string
		CreatureId      string
	}

	var v threatsJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	c.possibleThreats = structs.Initialise()
	for _, t := range v.PossibleThreats {
		c.possibleThreats.Add(t)
	}

	c.creatureID = v.CreatureId
	return nil
}
