package npc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"

	"github.com/onorton/cowboysindians/event"
	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/worldmap"
)

type dialogueType int

const (
	Basic dialogueType = iota
	Shopkeeper
	Sheriff
)

type interaction int

const (
	Normal interaction = iota
	Trade
	Bounty
)

var dialogueData map[string][]string = fetchDialogueData()

func fetchDialogueData() map[string][]string {
	data, err := ioutil.ReadFile("data/dialogue.json")
	check(err)
	var eD map[string][]string
	err = json.Unmarshal(data, &eD)
	check(err)
	return eD
}

type basicDialogue struct {
	seenPlayer bool
}

func (d *basicDialogue) initialGreeting() {
	if !d.seenPlayer {
		message.Enqueue(fmt.Sprintf("\"%s\"", dialogueData["Greetings"][rand.Intn(len(dialogueData["Greetings"]))]))
		d.seenPlayer = true
	}
}

func (d *basicDialogue) interact() interaction {
	message.PrintMessage(fmt.Sprintf("\"%s\"", dialogueData["Greetings"][rand.Intn(len(dialogueData["Greetings"]))]))
	return Normal
}

func (d *basicDialogue) resetSeen() {
	d.seenPlayer = false
}

func (d *basicDialogue) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	seenPlayerValue, err := json.Marshal(d.seenPlayer)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"SeenPlayer\":%s", seenPlayerValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (bd *basicDialogue) UnmarshalJSON(data []byte) error {

	type sdJson struct {
		SeenPlayer bool
	}

	var v sdJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	bd.seenPlayer = v.SeenPlayer

	return nil
}

type shopkeeperDialogue struct {
	seenPlayer bool
	world      *worldmap.Map
	b          worldmap.Building
}

func (d *shopkeeperDialogue) initialGreeting() {
	pX, pY := d.world.GetPlayer().GetCoordinates()

	storeGreetings := dialogueData[d.b.T.String()]
	if !d.seenPlayer && d.b.Inside(pX, pY) {
		message.Enqueue(fmt.Sprintf("\"%s %s\"", dialogueData["Greetings"][rand.Intn(len(dialogueData["Greetings"]))], storeGreetings[rand.Intn(len(storeGreetings))]))
		d.seenPlayer = true
	}
	if d.seenPlayer && !d.b.Inside(pX, pY) {
		message.Enqueue("\"Hope you stop by again soon.\"")
		d.seenPlayer = false
	}
}

func (d *shopkeeperDialogue) interact() interaction {
	message.PrintMessage("\"Sure. Feel free to look around.\"")
	return Trade
}

func (d *shopkeeperDialogue) resetSeen() {
	pX, pY := d.world.GetPlayer().GetCoordinates()

	// If player has not left the shop but is currently not visible, do not reset
	if !d.b.Inside(pX, pY) {
		d.seenPlayer = false
	}
}

func (d *shopkeeperDialogue) setMap(world *worldmap.Map) {
	d.world = world
}

func (d *shopkeeperDialogue) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	seenPlayerValue, err := json.Marshal(d.seenPlayer)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"SeenPlayer\":%s,", seenPlayerValue))

	buildingValue, err := json.Marshal(d.b)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Building\":%s", buildingValue))

	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (sd *shopkeeperDialogue) UnmarshalJSON(data []byte) error {

	type sdJson struct {
		SeenPlayer bool
		Building   worldmap.Building
	}

	var v sdJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	sd.seenPlayer = v.SeenPlayer
	sd.b = v.Building

	return nil
}

type sheriffDialogue struct {
	seenPlayer bool
	world      *worldmap.Map
	t          worldmap.Town
	b          worldmap.Building
	bounties   Bounties
}

func (d *sheriffDialogue) initialGreeting() {
	pX, pY := d.world.GetPlayer().GetCoordinates()

	if !d.seenPlayer && d.b.Inside(pX, pY) {
		message.Enqueue(fmt.Sprintf("\"%s %s\"", dialogueData["Greetings"][rand.Intn(len(dialogueData["Greetings"]))], dialogueData["Sheriff"][rand.Intn(len(dialogueData["Sheriff"]))]))
		d.seenPlayer = true
	}
	if d.seenPlayer && !d.b.Inside(pX, pY) {
		message.Enqueue("\"Don't be getting into no trouble, now.\"")
		d.seenPlayer = false
	}
}

func (d *sheriffDialogue) interact() interaction {
	message.PrintMessage("\"Yeah. We still got a few varmints to round up.\"")
	return Bounty
}

func (d *sheriffDialogue) resetSeen() {
	pX, pY := d.world.GetPlayer().GetCoordinates()

	// If player has not left the shop but is currently not visible, do not reset
	if !d.b.Inside(pX, pY) {
		d.seenPlayer = false
	}
}

func (d *sheriffDialogue) setMap(world *worldmap.Map) {
	d.world = world
}

func (d *sheriffDialogue) ProcessEvent(e event.Event) {
	switch ev := e.(type) {
	case event.WitnessedCrimeEvent:
		{
			crime := ev.Crime
			location := crime.Location()
			if location.X >= d.t.TX1 && location.X <= d.t.TX2 && location.Y >= d.t.TY1 && location.Y <= d.t.TY2 {
				d.bounties.addBounty(crime)
			}
		}
	}

}

func (d *sheriffDialogue) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	seenPlayerValue, err := json.Marshal(d.seenPlayer)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"SeenPlayer\":%s,", seenPlayerValue))

	townValue, err := json.Marshal(d.t)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Town\":%s,", townValue))

	buildingValue, err := json.Marshal(d.b)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Building\":%s,", buildingValue))

	bountiesValue, err := json.Marshal(d.bounties)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Bounties\":%s", bountiesValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (sd *sheriffDialogue) UnmarshalJSON(data []byte) error {

	type sdJson struct {
		SeenPlayer bool
		Building   worldmap.Building
		Bounties   Bounties
		Town       worldmap.Town
	}

	var v sdJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	sd.seenPlayer = v.SeenPlayer
	sd.t = v.Town
	sd.b = v.Building
	sd.bounties = v.Bounties

	event.Subscribe(sd)

	return nil
}

func unmarshalDialogue(dialogue map[string]interface{}) dialogue {
	dialogueJson, err := json.Marshal(dialogue)
	check(err)

	if _, ok := dialogue["Bounties"]; ok {
		var sd sheriffDialogue
		err = json.Unmarshal(dialogueJson, &sd)
		check(err)
		return &sd
	}

	if _, ok := dialogue["Building"]; ok {
		var sd shopkeeperDialogue
		err = json.Unmarshal(dialogueJson, &sd)
		check(err)
		return &sd
	}
	var bd basicDialogue
	err = json.Unmarshal(dialogueJson, &bd)
	check(err)
	return &bd
}

type dialogue interface {
	initialGreeting()
	interact() interaction
	resetSeen()
}
