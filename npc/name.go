package npc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/onorton/cowboysindians/ui"
)

type nameData struct {
	FirstNames []string
	LastNames  []string
	Towns      map[string][]string
}

var Names nameData = fetchNameData()

func fetchNameData() nameData {
	data, err := ioutil.ReadFile("data/names.json")
	check(err)
	var nD nameData
	err = json.Unmarshal(data, &nD)
	check(err)
	return nD
}

type npcName struct {
	name    string
	npcType string
	known   bool
}

func (n npcName) WithDefinite() string {
	if n.known {
		return n.name
	} else {
		return "the " + n.npcType
	}
}

func (n npcName) WithIndefinite() string {
	if n.known {
		return n.name
	} else {
		return "a " + n.npcType
	}
}

func (n npcName) String() string {
	if n.known {
		return n.name
	} else {
		return n.npcType
	}
}

func (n npcName) FullName() string {
	return n.name
}

func (n npcName) PlayerKnows() {
	n.known = true
}

func (n npcName) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	nameValue, err := json.Marshal(n.name)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Name\":%s,", nameValue))

	typeValue, err := json.Marshal(n.npcType)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Type\":%s,", typeValue))

	knownValue, err := json.Marshal(n.known)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Known\":%s", knownValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (n *npcName) UnmarshalJSON(data []byte) error {

	type npcNameJson struct {
		Name  string
		Type  string
		Known bool
	}

	var v npcNameJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	n.name = v.Name
	n.npcType = v.Type
	n.known = v.Known

	return nil
}

func unmarshalName(name map[string]interface{}) ui.Name {
	nameJson, err := json.Marshal(name)
	check(err)
	if _, ok := name["Known"]; ok {
		var npcName npcName
		err = json.Unmarshal(nameJson, &npcName)
		check(err)
		return &npcName
	}
	var plainName ui.PlainName
	err = json.Unmarshal(nameJson, &plainName)
	check(err)
	return &plainName
}
