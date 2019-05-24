package dice

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
)

type CompareOp string

const (
	CompareEquals  = "="
	CompareLess    = "<"
	CompareGreater = ">"
)

// A Modifier is a dice modifier that can apply to a set or a single die
type Modifier interface {
	// Apply executes a modifier against a Die.
	Apply(context.Context, *Die) error
	fmt.Stringer
}

// ModifierList is a slice of modifiers that implements Stringer.
type ModifierList []Modifier

func (m ModifierList) String() string {
	var buf bytes.Buffer
	for _, mod := range m {
		buf.WriteString(mod.String())
	}
	return buf.String()
}

var _ = Modifier(&RerollModifier{})

// RerollModifier is a modifier that rerolls a Die if a comparison is true, and
// possibly once and only once.
type RerollModifier struct {
	Compare string `json:"compare"`
	Point   int    `json:"point"`
	Once    bool   `json:"once"`
}

// func (m *RerollModifier) UnmarshalJSON(data []byte) error {
// 	// alias type to prevent an infinite loop
// 	type Alias RerollModifier
// 	return nil
// }

func (m *RerollModifier) String() string {
	var buf bytes.Buffer
	buf.WriteString("r")
	if m.Once {
		buf.WriteString("o")
	}
	// inferred equals if not specified
	if m.Compare != "=" {
		buf.WriteString(m.Compare)
	}
	buf.WriteString(strconv.Itoa(m.Point))
	return buf.String()
}

// Apply executes a RerollModifier against a Die
func (m *RerollModifier) Apply(ctx context.Context, d *Die) error {
	if m.Compare == "" {
		m.Compare = CompareEquals
	}
	switch m.Compare {
	case "", CompareEquals:
		for d.Result == float64(m.Point) {
			d.reroll(ctx)
		}
	case CompareLess:
		for d.Result <= float64(m.Point) {
			d.reroll(ctx)
		}
	case CompareGreater:
		for d.Result > float64(m.Point) {
			d.reroll(ctx)
		}
	default:
		return &ErrNotImplemented{"uncaught case for reroll"}
	}
	return nil
}
