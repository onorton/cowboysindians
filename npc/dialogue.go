package npc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"

	"github.com/onorton/cowboysindians/message"
	"github.com/onorton/cowboysindians/worldmap"
)

type dialogueType int

const (
	Basic dialogueType = iota
	Shopkeeper
	Sheriff
	EnemyDialogue
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

func getDialogue(dialogueType *dialogueType, world *worldmap.Map, t worldmap.Town, b worldmap.Building) dialogue {
	if dialogueType == nil {
		return nil
	}

	switch *dialogueType {
	case Basic:
		return &basicDialogue{false}
	case Shopkeeper:
		return &shopkeeperDialogue{false, world, b, t}
	case Sheriff:
		return &sheriffDialogue{false, world, b, t}
	case EnemyDialogue:
		return &enemyDialogue{false}
	}
	return &basicDialogue{false}
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

	typeValue, err := json.Marshal(Basic)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Type\":%s,", typeValue))

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
	t          worldmap.Town
}

func (d *shopkeeperDialogue) initialGreeting() {
	pX, pY := d.world.GetPlayer().GetCoordinates()

	if !d.seenPlayer && d.b.Inside(pX, pY) {
		storeGreetings := dialogueData[d.b.T.String()]
		dialogue := dialogueData["Greetings"][rand.Intn(len(dialogueData["Greetings"]))] + " " + storeGreetings[rand.Intn(len(storeGreetings))]
		dialogue = addTownToDialogue(dialogue, d.t.Name)
		message.Enqueue(fmt.Sprintf("\"%s\"", dialogue))
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

	typeValue, err := json.Marshal(Shopkeeper)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Type\":%s,", typeValue))

	seenPlayerValue, err := json.Marshal(d.seenPlayer)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"SeenPlayer\":%s,", seenPlayerValue))

	buildingValue, err := json.Marshal(d.b)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Building\":%s,", buildingValue))

	townValue, err := json.Marshal(d.t)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Town\":%s", townValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (sd *shopkeeperDialogue) UnmarshalJSON(data []byte) error {

	type sdJson struct {
		SeenPlayer bool
		Building   worldmap.Building
		Town       worldmap.Town
	}

	var v sdJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	sd.seenPlayer = v.SeenPlayer
	sd.b = v.Building
	sd.t = v.Town

	return nil
}

type sheriffDialogue struct {
	seenPlayer bool
	world      *worldmap.Map
	b          worldmap.Building
	t          worldmap.Town
}

func (d *sheriffDialogue) initialGreeting() {
	pX, pY := d.world.GetPlayer().GetCoordinates()
	if !d.seenPlayer && d.b.Inside(pX, pY) {
		dialogue := choose(dialogueData["Greetings"]) + " " + choose(dialogueData["Sheriff"])
		dialogue = addTownToDialogue(dialogue, d.t.Name)
		message.Enqueue(fmt.Sprintf("\"%s\"", dialogue))
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

func (d *sheriffDialogue) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	typeValue, err := json.Marshal(Sheriff)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Type\":%s,", typeValue))

	seenPlayerValue, err := json.Marshal(d.seenPlayer)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"SeenPlayer\":%s,", seenPlayerValue))

	buildingValue, err := json.Marshal(d.b)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Building\":%s,", buildingValue))

	townValue, err := json.Marshal(d.t)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Town\":%s", townValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (sd *sheriffDialogue) UnmarshalJSON(data []byte) error {

	type sdJson struct {
		SeenPlayer bool
		Building   worldmap.Building
		Town       worldmap.Town
	}

	var v sdJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	sd.seenPlayer = v.SeenPlayer
	sd.b = v.Building
	sd.t = v.Town
	return nil
}

type enemyDialogue struct {
	seenPlayer bool
}

func (d *enemyDialogue) initialGreeting() {
	if !d.seenPlayer {
		message.Enqueue(fmt.Sprintf("\"%s\"", choose(dialogueData["Enemy Greetings"])))
		d.seenPlayer = true
	}
}

func (d *enemyDialogue) interact() interaction {
	message.PrintMessage(fmt.Sprintf("\"%s\"", choose(dialogueData["Threats"])))
	return Normal
}

func (d *enemyDialogue) resetSeen() {
	d.seenPlayer = false
}

func (d *enemyDialogue) potentiallyThreaten() {
	// chance of threatening player
	if rand.Intn(10) == 0 {
		message.Enqueue(fmt.Sprintf("\"%s\"", choose(dialogueData["Threats"])))
	}
}

func (d *enemyDialogue) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	typeValue, err := json.Marshal(EnemyDialogue)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"Type\":%s,", typeValue))

	seenPlayerValue, err := json.Marshal(d.seenPlayer)
	if err != nil {
		return nil, err
	}
	buffer.WriteString(fmt.Sprintf("\"SeenPlayer\":%s", seenPlayerValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (d *enemyDialogue) UnmarshalJSON(data []byte) error {

	type edJson struct {
		SeenPlayer bool
	}

	var v edJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	d.seenPlayer = v.SeenPlayer

	return nil
}

func choose(dialogueChoices []string) string {
	return dialogueChoices[rand.Intn(len(dialogueChoices))]
}

func addTownToDialogue(dialogue string, townName string) string {
	return strings.Replace(dialogue, "[town]", townName, -1)
}

func unmarshalDialogue(dialogue map[string]interface{}) dialogue {
	dialogueJson, err := json.Marshal(dialogue)
	check(err)

	switch dialogueType(int(dialogue["Type"].(float64))) {
	case Basic:
		var bd basicDialogue
		err = json.Unmarshal(dialogueJson, &bd)
		check(err)
		return &bd
	case Shopkeeper:
		var sd shopkeeperDialogue
		err = json.Unmarshal(dialogueJson, &sd)
		check(err)
		return &sd
	case Sheriff:
		var sd sheriffDialogue
		err = json.Unmarshal(dialogueJson, &sd)
		check(err)
		return &sd
	case EnemyDialogue:
		var ed enemyDialogue
		err = json.Unmarshal(dialogueJson, &ed)
		check(err)
		return &ed
	}
	return nil
}

type dialogue interface {
	initialGreeting()
	interact() interaction
	resetSeen()
}
