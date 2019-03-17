package npc

import (
	"fmt"
	"math/rand"

	"github.com/onorton/cowboysindians/message"
)

var greetings []string = []string{"Howdy, partner!", "Howdy!", "Howdy, stranger!"}

type Dialogue struct {
	seenPlayerBefore bool
}

func (d *Dialogue) Greet() {
	if !d.seenPlayerBefore {
		message.Enqueue(fmt.Sprintf("\"%s\"", greetings[rand.Intn(len(greetings))]))
		d.seenPlayerBefore = true
	}
}
