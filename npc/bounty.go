package npc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/onorton/cowboysindians/event"
)

type bounty struct {
	criminal     string
	criminalName string
	crimes       map[string]struct{}
	value        int
}

func (b bounty) String() string {

	crimes := make([]string, 0)
	for c := range b.crimes {
		crimes = append(crimes, c)
	}

	return fmt.Sprintf("%s - %s - $%.2f", b.criminalName, strings.Join(crimes, ", "), float64(b.value)/100)
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
		Crimes       map[string]struct{}
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

type Bounties struct {
	bounties   []bounty
	seenCrimes []string
}

func (bounties *Bounties) addBounty(e event.CrimeEvent) {

	// Check if event has already been seen
	for _, eventId := range bounties.seenCrimes {
		if e.Id() == eventId {
			return
		}
	}
	bounties.seenCrimes = append(bounties.seenCrimes, e.Id())

	for _, b := range bounties.bounties {
		if b.criminal == e.Perpetrator() {
			b.crimes[e.Crime()] = struct{}{}
			b.value += e.Value()
			return
		}
	}
	bounties.bounties = append(bounties.bounties, bounty{e.Perpetrator(), e.PerpetratorName(), map[string]struct{}{e.Crime(): struct{}{}}, e.Value()})
}

func (bounties *Bounties) RemoveBounty(criminal string) (int, string) {
	for i, b := range bounties.bounties {
		if b.criminal == criminal {
			bounties.bounties = append((bounties.bounties)[:i], (bounties.bounties)[i+1:]...)
			return b.value, b.criminalName
		}
	}
	return 0, ""
}

func (bounties *Bounties) Bounties() []bounty {
	return bounties.bounties
}

func (bounties Bounties) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	bountiesValue, err := json.Marshal(bounties.bounties)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Bounties\":%s,", bountiesValue))

	seenCrimesValue, err := json.Marshal(bounties.seenCrimes)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"SeenCrimes\":%s", seenCrimesValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (bounties *Bounties) UnmarshalJSON(data []byte) error {

	type bountiesJson struct {
		Bounties   []bounty
		SeenCrimes []string
	}

	var v bountiesJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	bounties.bounties = v.Bounties
	bounties.seenCrimes = v.SeenCrimes

	return nil
}
