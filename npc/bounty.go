package npc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type bounty struct {
	criminal     string
	criminalName string
	crimes       []string
	value        int
}

func (b bounty) String() string {

	return fmt.Sprintf("%s - %s - $%.2f", b.criminalName, strings.Join(b.crimes, ", "), float64(b.value)/100)
}

func (b bounty) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	criminalValue, err := json.Marshal(b.criminal)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Criminal\":%s,", criminalValue))

	criminalNameValue, err := json.Marshal(b.criminalName)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"CriminalName\":%s,", criminalNameValue))

	crimesValue, err := json.Marshal(b.crimes)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Crimes\":%s,", crimesValue))

	value, err := json.Marshal(b.value)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Value\":%s", value))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (b *bounty) UnmarshalJSON(data []byte) error {

	type bountyJson struct {
		Criminal     string
		CriminalName string
		Crimes       []string
		Value        int
	}

	var v bountyJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	b.criminal = v.Criminal
	b.criminalName = v.CriminalName
	b.crimes = v.Crimes
	b.value = v.Value

	return nil
}

type Bounties []*bounty

func (bounties *Bounties) AddBounty(criminal *Npc, crime string, value int) {

	for _, b := range *bounties {
		if b.criminal == criminal.GetID() {
			b.crimes = append(b.crimes, crime)
			b.value += value
			return
		}
	}
	*bounties = append(*bounties, &bounty{criminal.GetID(), criminal.FullName(), []string{crime}, value})
}
