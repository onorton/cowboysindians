package npc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/onorton/cowboysindians/message"
)

var greetings []string = []string{"Howdy, partner!", "Howdy!", "Howdy, stranger!"}

type Dialogue struct {
	seenPlayerBefore bool
}

func (d *Dialogue) InitialGreeting() {
	if !d.seenPlayerBefore {
		message.Enqueue(fmt.Sprintf("\"%s\"", greetings[rand.Intn(len(greetings))]))
		d.seenPlayerBefore = true
	}
}

func (d *Dialogue) Greet() {
	message.PrintMessage(fmt.Sprintf("\"%s\"", greetings[rand.Intn(len(greetings))]))
}

func (d *Dialogue) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	seenPlayerBeforeValue, err := json.Marshal(d.seenPlayerBefore)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"SeenPlayerBefore\":%s", seenPlayerBeforeValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (d *Dialogue) UnmarshalJSON(data []byte) error {
	type dialogueJson struct {
		SeenPlayerBefore bool
	}

	var v dialogueJson
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	d.seenPlayerBefore = v.SeenPlayerBefore
	return nil
}
