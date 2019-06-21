package worldmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"

	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/ui"
)

type Attribute struct {
	value   int
	max     int
	effects []*item.Effect
}

func NewAttribute(initial, max int) *Attribute {
	return &Attribute{initial, max, make([]*item.Effect, 0)}
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

func (a *Attribute) updateWithEffect(e *item.Effect) {
	a.value, a.max = e.Update(a.value, a.max)
}

func (a *Attribute) AddEffect(e *item.Effect) {
	a.effects = append(a.effects, e)
	// When first added, apply the effect
	a.updateWithEffect(e)
	// Make sure value is correct if too high/low
	a.value = int(math.Max(0.0, math.Min(float64(a.max), float64(a.value))))
}

func (a *Attribute) Effects() []*item.Effect {
	return a.effects
}

func (a *Attribute) Update() {
	for i := 0; i < len(a.effects); i++ {
		effect := a.effects[i]
		if effect.Expired() {
			a.effects = append(a.effects[:i], a.effects[i+1:]...)
			i--
		}
		a.updateWithEffect(effect)
	}
	// Make sure value is correct if too high/low
	a.value = int(math.Max(0.0, math.Min(float64(a.max), float64(a.value))))
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

	buffer.WriteString(fmt.Sprintf("\"Maximum\":%s,", maxValue))

	effectsValue, err := json.Marshal(a.effects)
	if err != nil {
		return nil, err
	}

	buffer.WriteString(fmt.Sprintf("\"Effects\":%s", effectsValue))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

func (a *Attribute) UnmarshalJSON(data []byte) error {

	type attributeJson struct {
		Value   int
		Maximum int
		Effects []*item.Effect
	}

	var v attributeJson

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	a.value = v.Value
	a.max = v.Maximum
	a.effects = v.Effects

	return nil
}

type Skill int

const (
	Unarmed Skill = iota
)

// Interface shared by Player, Npc and Enemy
type Creature interface {
	CanSee
	Render() ui.Element
	GetInitiative() int
	MeleeAttack(Creature)
	TakeDamage(item.Damage, item.Effects, int)
	IsDead() bool
	IsCrouching() bool
	AttackHits(int) bool
	GetName() ui.Name
	GetAlignment() Alignment
	Update()
	GetID() string
	SetMap(*Map)
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

const (
	Player Alignment = iota
	Enemy
	Neutral
)
