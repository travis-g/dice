package dice

import (
	"context"
	"fmt"
	"strings"
)

var _ Roller = (*Group)(nil)

// Roller must be implemented for an object to be considered rollable.
type Roller interface {
	// Roll rolls the object and records results appropriately.
	Roll(context.Context) error

	// Reroll resets the object and re-rolls it.
	Reroll(context.Context) error

	// Total returns the summed results.
	Total(context.Context) (float64, error)

	// Drop marks the object dropped based on a provided boolean.
	Drop(context.Context, bool)

	// Must implement a String method; if the object has not been rolled String
	// should return a stringified representation of that can be re-parsed to
	// yield an equivalent property set.
	fmt.Stringer
}

// A RollerProperties object is the set of properties (usually extracted from a
// notation) that should be used to define a Die or group of like dice (a slice
// of multiple Die).
//
// This may be best broken into two properties types, a RollerProperties and a
// RollerGroupProperties.
type RollerProperties struct {
	Type    DieType `json:"type,omitempty" mapstructure:"type"`
	Size    int     `json:"size,omitempty" mapstructure:"size"`
	Result  *Result `json:"result,omitempty" mapstructure:"result"`
	Dropped bool    `json:"dropped,omitempty" mapstructure:"dropped"`
	Count   int     `json:"count,omitempty" mapstructure:"count"`

	// Modifiers for the dice or parent set
	DieModifiers   ModifierList `json:"die_modifiers,omitempty" mapstructure:"die_modifiers"`
	GroupModifiers ModifierList `json:"group_modifiers,omitempty" mapstructure:"group_modifiers"`
}

// A RollerFactory is a function that takes a properties object and returns a
// valid rollable die based off of the properties list. If there is an error
// creating a die off of the properties list an error should be returned.
type RollerFactory func(*RollerProperties) (Roller, error)

// RollerFactoryMap is the package-wide mapping of die types and the function to
// use to create a new die of that type. This map can be modified to create dice
// using different functions or to implement new die types.
var RollerFactoryMap = map[DieType]RollerFactory{
	TypePolyhedron: NewDie,
	TypeFudge:      NewDie,
}

// NewRoller creates a new Die to roll off of a supplied property set. The
// property set is modified/linted to better suit defaults in the event a
// properties list is reused.
//
// New dice created with this function are created by the per-DieType factories
// declared within the package-level RollerFactoryMap.
func NewRoller(props *RollerProperties) (Roller, error) {
	// Retrieve the factory function out of the package-wide map and use it to
	// create the new die.
	f, ok := RollerFactoryMap[props.Type]
	if !ok {
		return nil, fmt.Errorf("cannot create Die of type %s", props.Type)
	}
	return f(props)
}

// A Group is a slice of rollables.
type Group []Roller

// Total implements the Total method and sums a group of dice's totals.
func (g Group) Total(ctx context.Context) (total float64, err error) {
	for _, dice := range g {
		result, err := dice.Total(ctx)
		if err != nil {
			return total, err
		}
		total += result
	}
	return
}

func (g Group) String() string {
	temp := make([]string, len(g))
	for i, dice := range g {
		temp[i] = fmt.Sprintf("%v", dice.String())
	}
	t, _ := g.Total(context.TODO())
	return fmt.Sprintf("%s => %v", expression(strings.Join(temp, "+")), t)
}

// Drop is (presently) a noop on the group.
func (g Group) Drop(_ context.Context, _ bool) {
	// noop
}

// Copy returns a copy of the dice within the group
func (g Group) Copy() []Roller {
	self := make([]Roller, len(g))
	for i, k := range g {
		self[i] = k
	}
	return self
}

// Roll implements the Roller interface's Roll method by rolling each
// object/Roller within the group.
func (g Group) Roll(ctx context.Context) (err error) {
	for _, dice := range g {
		err = dice.Roll(ctx)
		if err != nil {
			break
		}
	}
	return err
}

// Reroll implements the dice.Reroll method by rerolling each object in it.
func (g Group) Reroll(ctx context.Context) (err error) {
	for _, dice := range g {
		err = dice.Reroll(ctx)
		if err != nil {
			break
		}
	}
	return err
}

// Expression returns an expression to represent the group's total. Dice in the
// group that are unrolled are replaced with their roll notations
func (g Group) Expression() string {
	dice := make([]string, 0)
	for _, die := range g {
		dice = append(dice, die.String())
	}
	// simplify the expression
	return strings.Replace(strings.Join(dice, "+"), "+-", "-", -1)
}

// RollerGroup is a wrapper around a Group that implements Roller. The Modifiers
// supplied at this level should be group-level modifiers, like drop/keep
// modifiers.
type RollerGroup struct {
	Group     `json:"group" mapstructure:"group"`
	Modifiers ModifierList `json:"modifiers,omitempty" mapstructure:"modifiers"`
}

// NewRollerGroup creates a new dice group with the count provided by the
// properties list. If a count of dice was not specified within the properties
// list it will default to a count of 1 and tweak the provided properties object
// accordingly.
func NewRollerGroup(props *RollerProperties) (*RollerGroup, error) {
	if props.Count <= 0 {
		props.Count = 1
	}
	dice := make([]Roller, props.Count)
	for i := range dice {
		die, err := NewRoller(props)
		if err != nil {
			return nil, err
		}
		dice[i] = die
	}

	return &RollerGroup{
		Group:     dice,
		Modifiers: props.GroupModifiers,
	}, nil
}

// Roll rolls each die embedded in the dice group.
func (d *RollerGroup) Roll(ctx context.Context) error {
	if err := d.Group.Roll(ctx); err != nil {
		return err
	}
	for _, mod := range d.Modifiers {
		err := mod.Apply(ctx, d)
		if err != nil {
			return err
		}
	}
	return nil
}

// Reroll re-rolls each die within the dice group.
func (d *RollerGroup) Reroll(ctx context.Context) error {
	if err := d.Group.Reroll(ctx); err != nil {
		return err
	}
	for _, mod := range d.Modifiers {
		err := mod.Apply(ctx, d)
		if err != nil {
			return err
		}
	}
	return nil
}

// All is a helper function that returns true if all dice.Interfaces of a slice
// match a predicate. All will return false on the first failure.
func All(vs []*Roller, f func(*Roller) bool) bool {
	for _, v := range vs {
		if !f(v) {
			return false
		}
	}
	return true
}
