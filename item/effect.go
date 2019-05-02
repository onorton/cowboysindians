package item

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Effect struct {
	effect     int
	onMax      bool
	duration   int
	activated  bool
	compounded bool
}

func NewEffect(effect, duration int, onMax bool) *Effect {
	return &Effect{effect, onMax, duration, false, false}
}

func NewInstantEffect(effect int) *Effect {
	return &Effect{effect, false, 1, false, false}
}

func NewOngoingEffect(effect int) *Effect {
	return &Effect{effect, false, -1, false, true}
}

func (e *Effect) Update(value, max int) (int, int) {
	if e.duration != 0 {
		if e.duration > 0 {
			e.duration--
		}

		if !e.activated || e.compounded {
			e.activated = true
			if e.onMax {
				return value, max + e.effect
			} else {
				return value + e.effect, max
			}
		}
	} else if e.onMax {
		// Return maximum to original value if applicable
		return value, max - e.effect
	}
	return value, max
}

func (e *Effect) Expired() bool {
	return e.duration == 0
}

func (e *Effect) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	effectValue, err := json.Marshal(e.effect)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Effect\":%s,", effectValue))

	onMaxValue, err := json.Marshal(e.onMax)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"OnMax\":%s,", onMaxValue))

	durationValue, err := json.Marshal(e.duration)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Duration\":%s,", durationValue))

	activatedValue, err := json.Marshal(e.activated)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Activated\":%s,", activatedValue))

	compoundedValue, err := json.Marshal(e.compounded)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Compounded\":%s", compoundedValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (e *Effect) UnmarshalJSON(data []byte) error {

	type effectJson struct {
		Effect     int
		OnMax      bool
		Duration   int
		Activated  bool
		Compounded bool
	}

	var v effectJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	e.effect = v.Effect
	e.onMax = v.OnMax
	e.duration = v.Duration
	e.activated = v.Activated
	e.compounded = v.Compounded

	return nil
}
