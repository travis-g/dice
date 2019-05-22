package dice

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

var _ Interface = (*Group)(nil)

// DieType is the enum of types that a die or dice can be
type DieType string

// Types of dice/dice groups
const (
	// TypeUnknown is any invalid/inconsistent type
	TypePolyhedron DieType = "polyhedron"
	TypeFate       DieType = "fate"
	TypeMultiple   DieType = "multiple"
)

func (t DieType) String() string {
	switch t {
	case TypePolyhedron:
		return "polyhedron"
	case TypeFate:
		return "fate"
	case TypeMultiple:
		return "multiple"
	default:
		return "unknown"
	}
}

// A Interface is any kind of rollable object. A Interface could be a single die
// or many dice of any type.
type Interface interface {
	// Roll will roll the object (if unrolled) and set the objects's result. If
	// the die already has a result it will not be rerolled.
	Roll(context.Context) (float64, error)

	// Total should return the totaled result. If the object is marked dropped 0
	// should be returned.
	Total(context.Context) (float64, error)

	// String/printing methods
	fmt.Stringer
	fmt.GoStringer
}

// A Group is a slice of dice interfaces.
type Group []Interface

// GroupProperties describes a die.
type GroupProperties struct {
	// Type is the type of dice within the group. If the properties object will
	// be used to create a new Group of dice, Type should be provided as a
	// dice.Type/uint.
	Type       interface{} `json:"type,omitempty"`
	Count      int         `json:"count"`
	Size       int         `json:"size,omitempty"`
	Result     float64     `json:"result"`
	Expression string      `json:"expression,omitempty"`
	Original   string      `json:"original,omitempty"`

	// Dice is any dice rolled as part of the group.
	Dice Group `json:"dice,omitempty"`

	// Unrolled indicates the die has not been rolled. If the die has been
	// rolled Unrolled will be false and omitted from marshaled JSON.
	Unrolled bool `json:"unrolled,omitempty"`

	// Dropped indicates the die should be excluded from totals.
	Dropped bool `json:"dropped,omitempty"`

	// DropKeep indicates how many child dice should be dropped (and from which
	// direction) if describing a set.
	DropKeep int `json:"drop,omitempty"`

	// Modifiers is the string of modifiers added to a given Group
	Modifiers []Modifier `json:"modifiers,omitempty"`
}

func (g *GroupProperties) String() string {
	return g.Dice.String()
}

// GoString returns the Go syntax for the object.
func (g *GroupProperties) GoString() string {
	return fmt.Sprintf("%#v", *g)
}

// Total implements the dice.Interface Total method and sums a group of
// Rollables' totals.
func (g *Group) Total(ctx context.Context) (total float64, err error) {
	total = 0.0
	for _, dice := range *g {
		result, err := dice.Total(ctx)
		if err != nil {
			return total, err
		}
		total += result
	}
	return
}

func (g *Group) String() string {
	temp := make([]string, len(*g))
	for i, dice := range *g {
		temp[i] = fmt.Sprintf("%v", dice.String())
	}
	t, _ := g.Total(context.Background())
	return fmt.Sprintf("%s => %v", strings.Join(temp, "+"), t)
}

// GoString returns the Go syntax for a group.
func (g Group) GoString() string {
	return fmt.Sprintf("%#v", g.Copy())
}

// Pointers returns the group as a slice of pointers to its dice.
func (g *Group) Pointers() []*Interface {
	self := make([]*Interface, len(*g))
	for i, k := range *g {
		self[i] = &k
	}
	return self
}

// Copy returns a copy of the dice within the group
func (g *Group) Copy() []Interface {
	self := make([]Interface, len(*g))
	for i, k := range *g {
		self[i] = k
	}
	return self
}

// Roll implements the dice.Interface Roll method by rolling each
// object/Interface within the group.
func (g *Group) Roll(ctx context.Context) (float64, error) {
	for _, dice := range *g {
		dice.Roll(ctx)
	}
	return g.Total(ctx)
}

// Expression returns an expression to represent the group's total. Dice in the
// group that are unrolled are replaced with their roll notations
func (g *Group) Expression() string {
	dice := make([]string, len(*g))
	for i, die := range *g {
		dice[i] = die.String()
	}
	// simplify the expression
	return strings.Replace(strings.Join(dice, "+"), "+-", "-", -1)
}

// Drop marks a die/dice within a group as dropped based on an input integer. If
// n is positive it will drop the n objects with the lowest Totals; if n is
// negative, it will drop the n objects with the highest Totals.
func (g *Group) Drop(drop int) {
	if drop == 0 {
		return
	}
	// create a copy of the array to sort and forward dice updates rather than
	// modifying the original order of the dice
	dice := g.Copy()

	sort.Slice(dice, func(i, j int) bool {
		ti, _ := (dice[i]).Total(context.Background())
		tj, _ := (dice[j]).Total(context.Background())
		return ti < tj
	})
	// fmt.Println(dice)
	// drop lowest to highest
	if drop > 0 {
		for i := 0; i < drop; i++ {
			switch t := (dice[i]).(type) {
			case *PolyhedralDie:
				t.Dropped = true
			case *FateDie:
				t.Dropped = true
			}
		}
	} else if drop < 0 {
		for i := len(dice) - 1; i >= len(dice)+drop; i-- {
			switch t := (dice[i]).(type) {
			case *PolyhedralDie:
				t.Dropped = true
			case *FateDie:
				t.Dropped = true
			}
		}
	}
}

// Properties calculates properties from a given group.
func Properties(ctx context.Context, g *Group) GroupProperties {
	props := GroupProperties{
		Count: len(*g),
		Dice:  *g,
	}
	dice := g.Pointers()

	switch len(*g) {
	// No dice: set unrolled by default and return
	case 0:
		props.Unrolled = true
		return props
	// Only one die: use its properties
	case 1:
		goto GROUP_CONSISTENT
	// There are multiple dice, so check that they're all of the same type
	default:
		kind := reflect.TypeOf(dice[0]).String()
		consistent := All(dice[1:], func(die *Interface) bool {
			this := reflect.TypeOf(die).String()
			return this == kind
		})
		if !consistent {
			goto GROUP_INCONSISTENT
		}
		goto GROUP_CONSISTENT
	}

GROUP_CONSISTENT:
	props.Expression = g.Expression()
	props.Result, _ = g.Total(ctx)
	switch t := (*dice[0]).(type) {
	case *PolyhedralDie:
		props.Size = t.Size
	}
	return props

GROUP_INCONSISTENT:
	props.Expression = g.Expression()
	props.Result, _ = g.Total(ctx)
	return props
}

// Roll rolls an arbitrary group of dice and returns the total.
func Roll(ctx context.Context, g *Group) (float64, error) {
	return g.Roll(ctx)
}

// NewGroup creates a new group based on provided seed of properties.
func NewGroup(props GroupProperties) (Group, error) {
	if props.Count == 0 {
		return Group{}, nil
	}
	group := make(Group, props.Count)

	switch props.Type {
	case TypeFate:
		for i := range group {
			group[i] = &FateDie{
				Type:     fateDieNotation,
				Unrolled: true,
			}
		}
	case TypePolyhedron:
		for i := range group {
			group[i] = &PolyhedralDie{
				Type:     fmt.Sprintf("d%d", props.Size),
				Size:     props.Size,
				Unrolled: true,
			}
		}
	default:
		return Group{}, fmt.Errorf("type %s not a valid dice.Type", props.Type)
	}
	return group, nil
}

// All is a helper function that returns true if all dice.Interfaces of a slice
// match a predicate. All will return false on the first failure.
func All(vs []*Interface, f func(*Interface) bool) bool {
	for _, v := range vs {
		if !f(v) {
			return false
		}
	}
	return true
}
