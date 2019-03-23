package npc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/onorton/cowboysindians/message"
)

type dialogueType int

const (
	Basic dialogueType = iota
	Shopkeeper
)

func getDialogue(t dialogueType) dialogue {
	if t == Basic {
		return &basicDialogue{false}
	} else {
		return &shopkeeperDialogue{false}
	}
}

var greetings []string = []string{"Howdy, partner!", "Howdy!", "Howdy, stranger!"}
var storeGreetings []string = []string{"Welcome to my store.", "Can I interest you in any of my wares?", "Welcome!"}

type basicDialogue struct {
	seenPlayerBefore bool
}

func (d *basicDialogue) initialGreeting() {
	if !d.seenPlayerBefore {
		message.Enqueue(fmt.Sprintf("\"%s\"", greetings[rand.Intn(len(greetings))]))
		d.seenPlayerBefore = true
	}
}

func (d *basicDialogue) interact() bool {
	message.PrintMessage(fmt.Sprintf("\"%s\"", greetings[rand.Intn(len(greetings))]))
	return false
}

func (d *basicDialogue) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	buffer.WriteString(fmt.Sprintf("\"Type\":\"Basic\","))

	seenPlayerBeforeValue, err := json.Marshal(d.seenPlayerBefore)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"SeenPlayerBefore\":%s", seenPlayerBeforeValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

type shopkeeperDialogue struct {
	seenPlayerBefore bool
}

func (d *shopkeeperDialogue) initialGreeting() {
	if !d.seenPlayerBefore {
		message.Enqueue(fmt.Sprintf("\"%s %s\"", greetings[rand.Intn(len(greetings))], storeGreetings[rand.Intn(len(storeGreetings))]))
		d.seenPlayerBefore = true
	}
}

func (d *shopkeeperDialogue) interact() bool {
	message.PrintMessage("\"Sure. Feel free to look around.\"")
	return true
}

func (d *shopkeeperDialogue) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	buffer.WriteString(fmt.Sprintf("\"Type\":\"Shopkeeper\","))
	seenPlayerBeforeValue, err := json.Marshal(d.seenPlayerBefore)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"SeenPlayerBefore\":%s", seenPlayerBeforeValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func unmarshalDialogue(dialogue map[string]interface{}) dialogue {

	if dialogue["Type"] == "Basic" {
		return &basicDialogue{dialogue["SeenPlayerBefore"].(bool)}
	} else {
		return &shopkeeperDialogue{dialogue["SeenPlayerBefore"].(bool)}
	}
}

type dialogue interface {
	initialGreeting()
	interact() bool
}
