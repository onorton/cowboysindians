package npc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/worldmap"
)

type dialogueType int

const (
	Basic dialogueType = iota
	Shopkeeper
)

var greetings []string = []string{"Howdy, partner!", "Howdy!", "Howdy, stranger!"}
var storeGreetings []string = []string{"Welcome to my store.", "Can I interest you in any of my wares?", "Welcome!"}

type basicDialogue struct {
	seenPlayer bool
}

func (d *basicDialogue) initialGreeting() {
	if !d.seenPlayer {
		message.Enqueue(fmt.Sprintf("\"%s\"", greetings[rand.Intn(len(greetings))]))
		d.seenPlayer = true
	}
}

func (d *basicDialogue) interact() bool {
	message.PrintMessage(fmt.Sprintf("\"%s\"", greetings[rand.Intn(len(greetings))]))
	return false
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

	if !d.seenPlayer && d.b.Inside(pX, pY) {
		message.Enqueue(fmt.Sprintf("\"%s %s\"", greetings[rand.Intn(len(greetings))], storeGreetings[rand.Intn(len(storeGreetings))]))
		d.seenPlayer = true
	}
	if d.seenPlayer && !d.b.Inside(pX, pY) {
		message.Enqueue("\"Hope you stop by again soon.\"")
		d.seenPlayer = false
	}
}

func (d *shopkeeperDialogue) interact() bool {
	message.PrintMessage("\"Sure. Feel free to look around.\"")
	return true
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

func unmarshalDialogue(dialogue map[string]interface{}) dialogue {
	dialogueJson, err := json.Marshal(dialogue)
	check(err)

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
	interact() bool
	resetSeen()
}
