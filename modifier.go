package dice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/pkg/errors"
)

// MaxRerolls is the maximum number of rerolls allowed due to a single
// modifier's application.
var MaxRerolls = 1000

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

// Initialize the compareStringMap for LookupCompareOp
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

// CompareTarget is the base comparison
type CompareTarget struct {
	Compare CompareOp `json:"compare,omitempty"`
	Target  int       `json:"target"`
}

// RerollModifier is a modifier that rerolls a Die if a comparison against the
// compare target is true.
type RerollModifier struct {
	*CompareTarget
	// TODO: fix Once after recursion tracking is solved
	// Once bool `json:"once"`
}

// MarshalJSON marshals the modifier into JSON and includes an internal type
// property.
func (m *RerollModifier) MarshalJSON() ([]byte, error) {
	type Faux RerollModifier
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Faux
	}{
		Type: "reroll",
		Faux: (*Faux)(m),
	})
}

func (m *RerollModifier) String() string {
	var buf bytes.Buffer
	write := buf.WriteString
	write("r")
	// inferred equals if not specified
	if m.Compare != EQL {
		write(m.Compare.String())
	}
	write(strconv.Itoa(m.Target))
	return buf.String()
}

// Apply executes a RerollModifier against a Roller. The modifier may be
// slightly modified the first time it is applied to ensure property
// consistency.
//
// The full roll needs to be recalculated in the event that one result may be
// acceptable for one reroll criteria, but not for one that was already
// evaluated. Impossible rerolls and impossible combinations of rerolls may
// cause a stack overflow from recursion.
func (m *RerollModifier) Apply(ctx context.Context, r Roller) error {
	if m == nil {
		return errors.New("nil modifier")
	}
	if m.Compare == EMPTY {
		m.Compare = EQL
	}
	ok, err := m.Valid(ctx, r)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return rerollApplyTail(ctx, m, r, 0)
}

// rerollApplyTail is a tail-recursive function to reroll a die based on a
// modifier.
func rerollApplyTail(ctx context.Context, m *RerollModifier, r Roller, rerolls int) error {
	fmt.Println("rerollApplyTail", rerolls, r)
	if err := r.Reroll(ctx); err != nil {
		return err
	}
	rerolls++
	ok, err := m.Valid(ctx, r)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	// if m.Once && rerolls >= 1 {
	// 	return nil
	// }
	return rerollApplyTail(ctx, m, r, rerolls)
}

// Valid checks if the supplied die is valid against the modifier. If not valid
// the reroll modifier should be applied, unless there is an error.
func (m *RerollModifier) Valid(ctx context.Context, r Roller) (bool, error) {
	if m == nil {
		return false, errors.New("nil modifier")
	}
	var (
		result float64
		err    error
	)
	if m.Compare == EMPTY {
		m.Compare = EQL
	}
	if result, err = r.Total(ctx); err != nil {
		// return invalid if error
		return false, err
	}
	switch m.Compare {
	// until the comparison operation succeeds and the reroll passes, keep
	// rerolling.
	case EQL:
		return result != float64(m.Target), nil
	case LSS, LEQ:
		return !(result <= float64(m.Target)), nil
	case GTR, GEQ:
		return !(result >= float64(m.Target)), nil
	default:
		err = &ErrNotImplemented{
			fmt.Sprintf("uncaught case for reroll compare: %s", m.Compare),
		}
		return false, err
	}
}

// A DropKeepMethod is a method to use when evaluating a drop/keep modifier
// against a dice group.
type DropKeepMethod string

// Drop/keep methods.
const (
	DropKeepMethodUnknown     DropKeepMethod = ""
	DropKeepMethodDrop        DropKeepMethod = "d"
	DropKeepMethodDropLowest  DropKeepMethod = "dl"
	DropKeepMethodDropHighest DropKeepMethod = "dh"
	DropKeepMethodKeep        DropKeepMethod = "k"
	DropKeepMethodKeepLowest  DropKeepMethod = "kl"
	DropKeepMethodKeepHighest DropKeepMethod = "kh"
)

// A DropKeepModifier is a modifier to drop the highest or lowest Num dice
// within a group by marking them as Dropped. The Method used to apply the
// modifier defines if the dice are dropped or kept (meaning the Num highest
// dice are not dropped).
type DropKeepModifier struct {
	Method DropKeepMethod `json:"op,omitempty"`
	Num    int            `json:"num"`
}

func (d *DropKeepModifier) String() string {
	return string(d.Method)
}

// Apply executes a DropKeepModifier against a Roller. If the Roller is not a
// Group an error is returned.
func (d *DropKeepModifier) Apply(ctx context.Context, r Roller) error {
	group, ok := r.(*RollerGroup)
	if !ok {
		return errors.New("target for modifier not a dice group")
	}

	// create a duplicate of the slice to sort
	dice := group.Copy()

	sort.Slice(dice, func(i, j int) bool {
		ti, _ := (dice[i]).Total(ctx)
		tj, _ := (dice[j]).Total(ctx)
		return ti < tj
	})

	switch d.Method {
	case DropKeepMethodDrop, DropKeepMethodDropLowest:
		// drop lowest Num
		for i := 0; i < d.Num && i < len(dice); i++ {
			dice[i].Drop(ctx, true)
		}
	case DropKeepMethodKeep, DropKeepMethodKeepHighest:
		// drop all but highest Num
		for i := 0; i < len(dice)-d.Num && i < len(dice); i++ {
			dice[i].Drop(ctx, true)
		}
	case DropKeepMethodDropHighest:
		for i := len(dice) - d.Num; i < len(dice) && i < len(dice); i++ {
			dice[i].Drop(ctx, true)
		}
	case DropKeepMethodKeepLowest:
		for i := d.Num; i < len(dice) && i < len(dice); i++ {
			dice[i].Drop(ctx, true)
		}
	default:
		return &ErrNotImplemented{"unknown drop/keep method"}
	}
	return nil
}
