package npc

import (
	"bytes"
	"encoding/json"
	"fmt"
)

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
