package dice

import (
	"fmt"
	"strings"
)

var _ Interface = (*Group)(nil)
var _ RollableSet = (*Group)(nil)

// Type is the enum of types that a die or dice can be
type Type uint

const (
	// TypeInvalid is any invalid type
	TypeInvalid Type = 0

	// TypePolyhedron indicates a die is any standard polyhedron
	TypePolyhedron Type = iota

	// TypeFate indicates the die is a Fate/Fudge die
	TypeFate

	// TypeMultiple indicates a dice group is a mix of types
	TypeMultiple
)

func (t Type) String() string {
	switch t {
	case TypePolyhedron:
		return "polyhedron"
	case TypeFate:
		return "fate"
	case TypeMultiple:
		return "multiple"
	default:
		return "invalid"
	}
}

// A Interface is any kind of rollable object. A Interface could be a single die
// or many dice of any type.
type Interface interface {
	// Roll should be used to also set the object's Result
	Roll() (float64, error)
	Total() float64

	fmt.Stringer
	fmt.GoStringer
}

// A Group is a slice of dice interfaces.
type Group []Interface

// GroupProperties describes a die.
type GroupProperties struct {
	Type       interface{} `json:"type,omitempty"`
	Size       int         `json:"size,omitempty"`
	Count      int         `json:"count"`
	Result     float64     `json:"result"`
	Expression string      `json:"expression,omitempty"`

	// Dice is any dice rolled as part of the group.
	Dice Group `json:"dice,omitempty"`

	// Unrolled indicates the die has not been rolled. If the die has been
	// rolled Unrolled will be false and omitted from marshaled JSON.
	Unrolled bool `json:"unrolled,omitempty"`

	// Dropped indicates the die should be excluded from totals.
	Dropped bool `json:"dropped,omitempty"`

	// Drop indicates how many child dice should be dropped (and from which
	// direction) if describing a set.
	Drop int `json:"drop,omitempty"`
}

func (g *GroupProperties) String() string {
	return g.Dice.String()
}

// GoString returns the Go syntax for the object.
func (g *GroupProperties) GoString() string {
	return fmt.Sprintf("%#v", *g)
}

// Total sums a group of Rollables.
func (g Group) Total() float64 {
	sum := 0.0
	for _, dice := range g {
		dice.Roll()
		sum += dice.Total()
	}
	return sum
}

func (g *Group) String() string {
	temp := make([]string, len(*g))
	for i, dice := range *g {
		temp[i] = fmt.Sprintf("%v", dice.String())
	}
	return fmt.Sprintf("%s => %v", strings.Join(temp, "+"), g.Total())
}

// GoString returns the Go syntax for a group.
func (g Group) GoString() string {
	return fmt.Sprintf("%#v", g.Self())
}

// Self returns the group as a slice of interfaces
func (g *Group) Self() []Interface {
	self := make([]Interface, len(*g))
	for i, k := range *g {
		self[i] = k
	}
	return self
}

// Roll rolls each dice interface within the group.
func (g *Group) Roll() (float64, error) {
	for _, dice := range *g {
		dice.Roll()
	}
	return g.Total(), nil
}

// Expression returns an expression to represent the group's total. Dice in the
// group that are unrolled are replaced with their roll notations
func (g Group) Expression() string {
	dice := make([]string, len(g))
	for i, die := range g {
		dice[i] = die.String()
	}
	// simplify the expression
	return strings.Replace(strings.Join(dice, "+"), "+-", "-", -1)
}

// Properties calculates properties from a given group.
func Properties(g *Group) GroupProperties {
	if len(*g) == 0 {
		return GroupProperties{}
	}
	return GroupProperties{
		Count:      len(*g),
		Result:     g.Total(),
		Type:       g.Type(),
		Dice:       *g,
		Expression: g.Expression(),
	}
}

// Type determine the type of a group of dice, if consistent.
func (g Group) Type() interface{} {
	// HACK(tssde71): replace with All()
	var kind interface{}
	switch t := g[0]; t.(type) {
	// TODO(tssde71): add fate type
	case *Die:
		kind = TypePolyhedron
	default:
		kind = TypeInvalid
	}
	return kind
}

// Roll rolls a group of dice and returns the total.
func Roll(g Group) (float64, error) {
	return g.Roll()
}

// NewGroup creates a new group based on provided properties.
func NewGroup(props GroupProperties) Group {
	if props.Count == 0 {
		return Group{}
	}
	group := make(Group, props.Count)

	for i := range group {
		group[i] = &Die{
			Type:     fmt.Sprintf("d%d", props.Size),
			Size:     props.Size,
			Unrolled: true,
		}
	}
	return group
}

// A RollableSet are sets of Rollables.
type RollableSet interface {
	Interface

	Expression() string
}

// Expand returns the expanded representation of a set based on the set's type.
func Expand(set RollableSet) string {
	switch t := set.(type) {
	case *DieSet:
		return t.Expanded
	case *FateDieSet:
		return t.Expanded
	default:
		return ""
	}
}
