package dice

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

var _ Roller = (*Group)(nil)

// Roller must be implemented for an object to be considered rollable.
// Internally, a Roller and should maintain a "total rolls" count.
type Roller interface {
	// FullRoll rolls the object at the macro level, inclusive of testing and
	// applying modifiers.
	FullRoll(context.Context) error

	// Roll rolls and records the object's Result. Roll should not apply
	// modifiers.
	Roll(context.Context) error

	// Reroll resets the object and should re-roll the core die by calling Roll.
	// Methods used by Reroll should not call FullRoll without safeguards to
	// prevent a stack overflow.
	Reroll(context.Context) error

	// Total returns the summed results, omitting any dropped results.
	Total(context.Context) (float64, error)

	// Value returns the rolled face value of the Roller, regardless of whether
	// the Roller was dropped. The Value method should be used when sorting.
	Value(context.Context) (float64, error)

	// Drop marks the object dropped based on a provided boolean.
	Drop(context.Context, bool)

	// IsDropped returns the dropped status of the Roller.
	IsDropped(context.Context) bool

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
	Type    DieType   `json:"type,omitempty" mapstructure:"type"`
	Size    int       `json:"size,omitempty" mapstructure:"size"`
	Result  *Result   `json:"result,omitempty" mapstructure:"result"`
	Results []*Result `json:"results,omitempty" mapstructure:"results"`
	Count   int       `json:"count,omitempty" mapstructure:"count"`

	// Modifiers for the dice or parent set
	RollModifiers  ModifierList `json:"roll_modifiers,omitempty" mapstructure:"roll_modifiers"`
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
		return nil, fmt.Errorf("no factory for type %s", props.Type)
	}
	return f(props)
}

// A Group is a slice of Rollers. Groups can be comprised of any mix of
// individual dice and other nested Groups.
type Group []Roller

// Total implements the Total method and sums a dice group's totals, excluding
// values of dropped dice.
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

// Value returns the total value of a Group for sorting purposes. It should
// return the Group's Total still.
func (g Group) Value(ctx context.Context) (float64, error) {
	return g.Total(ctx)
}

func (g Group) String() string {
	// TODO: use buffers/[]byte
	if len(g) == 0 {
		// short circuit
		return "0 => 0"
	}
	temp := make([]string, len(g))
	for i, dice := range g {
		temp[i] = fmt.Sprintf("%v", dice.String())
	}
	t, _ := g.Total(context.Background())
	return fmt.Sprintf("%s => %.0f", SimplifyExpression(strings.Join(temp, "+")), t)
}

// Drop is (presently) a no op on the group.
func (g Group) Drop(_ context.Context, _ bool) {
	// noop
}

// IsDropped returns whether the Group is dropped. It always returns false.
func (g Group) IsDropped(_ context.Context) bool {
	return false
}

// Copy returns a copy of the dice within the group
func (g Group) Copy() []Roller {
	self := make([]Roller, len(g))
	for i, k := range g {
		self[i] = k
	}
	return self
}

// FullRoll implements the Roller interface's FullRoll method by rolling each
// object/Roller within the group.
func (g Group) FullRoll(ctx context.Context) (err error) {
	for _, dice := range g {
		err = dice.FullRoll(ctx)
		if err != nil {
			break
		}
	}
	return err
}

// Roll rolls each of the dice in the group without applying their modifiers.
func (g Group) Roll(ctx context.Context) (err error) {
	for _, dice := range g {
		err = dice.Roll(ctx)
		if err != nil {
			break
		}
	}
	return err
}

// Reroll implements the Reroll method by rerolling each object in the group.
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
// group that are unrolled are replaced with their roll notations and dropped
// dice results are omitted.
func (g Group) Expression() string {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// return 0 if the Group is empty.
	if len(g) == 0 {
		return "0"
	}

	dice := make([]string, 0)
	for _, die := range g {
		if !die.IsDropped(ctx) {
			dice = append(dice, die.String())
		}
	}
	// simplify the expression
	return strings.Replace(strings.Join(dice, "+"), "+-", "-", -1)
}

// Group should implement the sort interface.
var _ sort.Interface = (*Group)(nil)

// Len returns the number of elements in a Group.
func (g Group) Len() int {
	return len(g)
}

// Less determines the sort order of Rollers in a Group.
func (g Group) Less(i, j int) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// if i is dropped, sort i after
	if g[i].IsDropped(ctx) {
		return true
	}
	// if j is dropped, sort j after
	if g[j].IsDropped(ctx) {
		return false
	}

	// if i's total is less than j, sort i after
	it, _ := g[i].Total(ctx)
	jt, _ := g[j].Total(ctx)
	if it < jt {
		return true
	}
	return false
}

// Swap swaps the positions of two Rollers in a Group. This method is not thread
// safe.
func (g Group) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

// RollerGroup is a wrapper around a Group that implements Roller. The Modifiers
// supplied at this level should be group-level modifiers, like drop/keep
// modifiers.
type RollerGroup struct {
	Group         `json:"group" mapstructure:"group"`
	Modifiers     ModifierList `json:"modifiers,omitempty" mapstructure:"modifiers"`
	RollModifiers ModifierList `json:"roll_modifiers,omitempty" mapstructure:"roll_modifiers"`
}

// NewRollerGroup creates a new dice group with the count provided by the
// properties list. If a count of dice was not specified within the properties
// list it will default to a count of 1 and tweak the provided properties object
// accordingly.
func NewRollerGroup(props *RollerProperties) (*RollerGroup, error) {
	if props.Count == 0 {
		return &RollerGroup{
			Modifiers: props.GroupModifiers,
		}, nil
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

// FullRoll rolls each die embedded in the dice group.
func (d *RollerGroup) FullRoll(ctx context.Context) error {
	if err := d.Group.FullRoll(ctx); err != nil {
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

// All is a helper function that returns true if all Rollers of a slice match a
// predicate. All will return false on the first failure.
func All(vs []Roller, f func(Roller) bool) bool {
	for _, v := range vs {
		if !f(v) {
			return false
		}
	}
	return true
}

// Filter is a helper function that returns a slice of Rollers that match a
// predicate out of an input slice.
func Filter(vs []Roller, f func(Roller) bool) []Roller {
	var rolls = []Roller{}
	for _, v := range vs {
		if f(v) {
			rolls = append(rolls, v)
		}
	}
	return rolls
}
