package dice

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

var _ Interface = (*Group)(nil)

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
	return fmt.Sprintf("%#v", g.Pointers())
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

// Drop marks a dice within a group as dropped based on an input integer.
func (g *Group) Drop(drop int) {
	dice := g.Copy()
	sort.Slice(dice, func(i, j int) bool {
		return (dice[i]).Total() < (dice[j]).Total()
	})
	// fmt.Println(dice)
	// drop lowest to highest
	if drop > 0 {
		for i := 0; i < drop; i++ {
			switch t := (dice[i]).(type) {
			case *Die:
				t.Dropped = true
			case *FateDie:
				t.Dropped = true
			}
		}
	} else if drop < 0 {
		for i := len(dice) - 1; i >= len(dice)+drop; i-- {
			switch t := (dice[i]).(type) {
			case *Die:
				t.Dropped = true
			case *FateDie:
				t.Dropped = true
			}
		}
	}
}

// Properties calculates properties from a given group.
func Properties(g *Group) GroupProperties {
	props := GroupProperties{
		Count: len(*g),
		Dice:  *g,
	}
	dice := g.Pointers()

	switch len(*g) {
	// No dice; set unrolled by default
	case 0:
		props.Unrolled = true
		return props
	// only one die: use its properties
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
	props.Result = g.Total()
	switch t := (*dice[0]).(type) {
	case *Die:
		props.Size = t.Size
	}
	return props

GROUP_INCONSISTENT:
	props.Expression = g.Expression()
	props.Result = g.Total()
	return props
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

	switch props.Type {
	case TypeFate:
		for i := range group {
			group[i] = &FateDie{
				Type:     fateDieNotation,
				Unrolled: true,
			}
		}
	default:
		for i := range group {
			group[i] = &Die{
				Type:     fmt.Sprintf("d%d", props.Size),
				Size:     props.Size,
				Unrolled: true,
			}
		}
	}
	return group
}
