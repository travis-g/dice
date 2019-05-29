package dice

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
)

// CompareOp is an comparison operator usable in modifiers.
type CompareOp string

// Comparison operators.
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

// RerollModifier is a modifier that rerolls a Die if a comparison is true.
type RerollModifier struct {
	Compare string `json:"compare"`
	Point   int    `json:"point"`
}

func (m *RerollModifier) String() string {
	var buf bytes.Buffer
	write := buf.WriteString
	write("r")
	// inferred equals if not specified
	if m.Compare != "=" {
		write(m.Compare)
	}
	write(strconv.Itoa(m.Point))
	return buf.String()
}

// Apply executes a RerollModifier against a Die. The modifier may be slightly
// modified the first time it is applied to ensure property consistency.
func (m *RerollModifier) Apply(ctx context.Context, d *Die) error {
	if m.Compare == "" {
		m.Compare = CompareEquals
	}
	switch m.Compare {
	case CompareEquals:
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
		return &ErrNotImplemented{
			fmt.Sprintf("uncaught case for reroll compare: %s", m.Compare),
		}
	}
	return nil
}
