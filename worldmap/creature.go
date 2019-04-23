package worldmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"

	"github.com/onorton/cowboysindians/ui"
)

type Attribute struct {
	value int
	max   int
}

func NewAttribute(initial, max int) *Attribute {
	return &Attribute{initial, max}
}

func (a *Attribute) Modify(amount int) {
	a.value = int(math.Max(0.0, math.Min(float64(a.max), float64(a.value+amount))))
}

func (a *Attribute) Status() string {
	return fmt.Sprintf("%d/%d", a.value, a.max)
}

func (a *Attribute) Value() int {
	return a.value
}

func (a *Attribute) Maximum() int {
	return a.max
}

func (a *Attribute) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	value, err := json.Marshal(a.value)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Value\":%s,", value))

	maxValue, err := json.Marshal(a.max)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Maximum\":%s", maxValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (a *Attribute) UnmarshalJSON(data []byte) error {

	type attributeJson struct {
		Value   int
		Maximum int
	}

	var v attributeJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	a.value = v.Value
	a.max = v.Maximum

	return nil
}

// Interface shared by Player, Npc and Enemy
type Creature interface {
	CanSee
	Render() ui.Element
	GetInitiative() int
	MeleeAttack(Creature)
	TakeDamage(int)
	IsDead() bool
	IsCrouching() bool
	AttackHits(int) bool
	GetName() ui.Name
	GetAlignment() Alignment
	Update()
	GetID() string
}

type hasPosition interface {
	GetCoordinates() (int, int)
	SetCoordinates(int, int)
}

type CanSee interface {
	GetVisionDistance() int
	hasPosition
}

type CanCrouch interface {
	IsCrouching() bool
	Standup()
	Crouch()
}
