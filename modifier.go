package dice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

// A Modifier is a dice modifier that can apply to a set or a single die
type Modifier interface {
	// Apply executes a modifier against a Die.
	Apply(context.Context, Roller) error
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

// CompareOp is an comparison operator usable in modifiers.
type CompareOp int

// Comparison operators.
const (
	EMPTY CompareOp = iota

	compareOpStart
	EQL // =
	LSS // <
	GTR // >
	LEQ // <=
	GEQ // >=
	compareOpEnd
)

var compares = [...]string{
	EMPTY: "",

	EQL: "=",
	LSS: "<",
	GTR: ">",
	LEQ: "<=",
	GEQ: ">=",
}
var compareStringMap map[string]CompareOp

func init() {
	compareStringMap = make(map[string]CompareOp)
	for i := compareOpStart + 1; i < compareOpEnd; i++ {
		compareStringMap[compares[i]] = i
	}
}

// LookupCompareOp returns the CompareOp that is represented by a given string.
func LookupCompareOp(s string) CompareOp {
	return compareStringMap[s]
}

func (c CompareOp) String() string {
	s := ""
	if 0 <= c && c < CompareOp(len(compares)) {
		s = compares[c]
	}
	return s
}

// MarshalJSON ensures the CompareOp is encoded as its string representation.
func (c *CompareOp) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// UnmarshalJSON enables JSON encoded string versions of CompareOps to be
// converted to their appropriate counterparts.
func (c *CompareOp) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return errors.Wrap(err, "error unmarshaling json to CompareOp")
	}
	*c = LookupCompareOp(str)
	return nil
}

// ComparePoint is the base comparison
type ComparePoint struct {
	Compare CompareOp `json:"compare"`
	Point   int       `json:"point"`
}

// RerollModifier is a modifier that rerolls a Die if a comparison is true.
type RerollModifier struct {
	ComparePoint
}

func (m *RerollModifier) String() string {
	var buf bytes.Buffer
	write := buf.WriteString
	write("r")
	// inferred equals if not specified
	if m.Compare != EQL {
		write(m.Compare.String())
	}
	write(strconv.Itoa(m.Point))
	return buf.String()
}

// Apply executes a RerollModifier against a Roller. The modifier may be slightly
// modified the first time it is applied to ensure property consistency.
func (m *RerollModifier) Apply(ctx context.Context, r Roller) (err error) {
	var result float64
	if m.Compare == EMPTY {
		m.Compare = EQL
	}
	result, err = r.Total(ctx)
	if err != nil {
		return
	}
	reroll := func() (err error) {
		err = r.Reroll(ctx)
		if err != nil {
			return
		}
		result, err = r.Total(ctx)
		if err != nil {
			return
		}
		return
	}
	switch m.Compare {
	case EQL:
		for result == float64(m.Point) {
			err = reroll()
			if err != nil {
				return
			}
		}
	case LSS:
		for result <= float64(m.Point) {
			err = reroll()
			if err != nil {
				return
			}
		}
	case GTR:
		for result > float64(m.Point) {
			err = reroll()
			if err != nil {
				return
			}
		}
	default:
		err = &ErrNotImplemented{
			fmt.Sprintf("uncaught case for reroll compare: %s", m.Compare),
		}
	}
	return err
}
