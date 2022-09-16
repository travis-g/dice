package dice

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// A Modifier is a dice modifier that can apply to a set or a single die.
type Modifier interface {
	// Apply executes a modifier against a Die.
	Apply(context.Context, Roller) error
	fmt.Stringer
}

// ModifierList is a slice of modifiers that implements Stringer.
type ModifierList []Modifier

func (m ModifierList) String() string {
	var b strings.Builder
	for _, mod := range m {
		b.WriteString(mod.String())
	}
	return b.String()
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
	Once bool `json:"once,omitempty"`
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
	var b strings.Builder
	write := b.WriteString
	write("r")
	// inferred equals if not specified
	if m.Compare != EQL {
		write(m.Compare.String())
	}
	write(strconv.Itoa(m.Target))
	return b.String()
}

// Apply executes a RerollModifier against a Roller. The modifier may be
// slightly modified the first time it is applied to ensure property
// consistency.
//
// The full roll needs to be recalculated in the event that one result may be
// acceptable for one reroll criteria, but not for one that was already
// evaluated. An ErrRerolled error will be returned if the die was rerolled in
// case other modifiers on the die need to be reapplied. Impossible rerolls and
// impossible combinations of rerolls may cause a stack overflow from recursion
// without a safeguard like MaxRerolls.
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
	// if once, do only once
	if m.Once {
		return r.Reroll(ctx)
	}
	// reroll until valid
	return rerollApplyTail(ctx, m, r)
}

// rerollApplyTail is a tail-recursive function to reroll a die based on a
// modifier. The error returned must be an ErrRerolled to indicate the die was
// changed via rerolling. ErrRerolled may need to bubble up to the rollable's
// core rolling functions to indicate other modifiers must be reapplied.
//
// Tail recursion is used here as the stack has the potential to grow quite
// large if the recursive calls are not optimized.
func rerollApplyTail(ctx context.Context, m *RerollModifier, r Roller) error {
	if m == nil {
		return errors.New("nil modifier")
	}
	if err := r.Reroll(ctx); err != nil {
		return err
	}
	ok, err := m.Valid(ctx, r)
	if err != nil {
		return err
	}
	// Now that die has settled, return rerolled error
	if ok {
		return ErrRerolled
	}
	return rerollApplyTail(ctx, m, r)
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

	// TODO: do these dice need to be sorted by their result value/should
	// already-dropped dice be filtered and excluded?
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
		for i := len(dice) - d.Num; i < len(dice); i++ {
			dice[i].Drop(ctx, true)
		}
	case DropKeepMethodKeepLowest:
		for i := d.Num; i < len(dice); i++ {
			dice[i].Drop(ctx, true)
		}
	default:
		return &ErrNotImplemented{"unknown drop/keep method"}
	}
	return nil
}

// A CriticalSuccessModifier shifts or sets the compare point/range used to
// classify a die's result as a critical success.
type CriticalSuccessModifier struct {
	*CompareTarget
}

func (m *CriticalSuccessModifier) String() string {
	var b strings.Builder
	write := b.WriteString
	write("cs")
	// inferred equals if not specified
	if m.Compare != EQL {
		write(m.Compare.String())
	}
	write(strconv.Itoa(m.Target))
	return b.String()
}

// A CriticalFailureModifier shifts or sets the compare point/range used to
// classify a die's result as a critical failure.
type CriticalFailureModifier struct {
	*CompareTarget
}

func (m *CriticalFailureModifier) String() string {
	var b strings.Builder
	write := b.WriteString
	write("cf")
	// inferred equals if not specified
	if m.Compare != EQL {
		write(m.Compare.String())
	}
	write(strconv.Itoa(m.Target))
	return b.String()
}

// SortDirection is a possible direction for sorting dice.
type SortDirection uint8

// Sort directions for sorting modifiers.
const (
	SortDirectionAscending SortDirection = iota
	SortDirectionDescending
)

// SortModifier is a modifier that will sort the Roller group.
type SortModifier struct {
	Direction SortDirection `json:"direction"`
}

func (s *SortModifier) String() string {
	if s.Direction == SortDirectionDescending {
		return "sd"
	}
	return "s"
}

// Apply applies a sort to a Roller.
func (s *SortModifier) Apply(ctx context.Context, r Roller) error {
	group, ok := r.(*RollerGroup)
	if !ok {
		return errors.New("target for modifier not a dice group")
	}

	switch s.Direction {
	case SortDirectionAscending:
		sort.Sort(group.Group)
	case SortDirectionDescending:
		sort.Sort(sort.Reverse(group.Group))
	}
	return nil
}

type ExplodeModifier struct {
	*CompareTarget
	Once bool `json:"once,omitempty"`
}

func (m *ExplodeModifier) String() string {
	var b strings.Builder
	write := b.WriteString
	write("!")
	if m.Once {
		write("o")
	}
	// inferred equals if not specified
	if m.Compare != EQL {
		write(m.Compare.String())
	}
	if m.Target > 0 {
		write(strconv.Itoa(m.Target))
	}
	return b.String()
}

func (m *ExplodeModifier) Apply(ctx context.Context, r Roller) error {
	die, ok := r.(*Die)
	if !ok {
		return errors.New("roller not a die")
	}

	group, ok := die.Parent().(*Group)
	if !ok {
		return errors.New("parent is not a Group")
	}
	if group == nil {
		return errors.New("parent is nil")
	}
	if ok, err := m.Valid(ctx, r); err == nil && ok {
		group.Add(&Die{
			Type:      die.Type,
			Size:      die.Size,
			Modifiers: die.Modifiers, // FIXME: these need to be a copy
			parent:    group,
		})
	}

	return nil
}

func (m *ExplodeModifier) Valid(ctx context.Context, r Roller) (bool, error) {
	if m == nil {
		return false, errors.New("nil modifier")
	}
	var (
		result float64
		err    error
	)

	result, err = r.Value(ctx)
	if err != nil {
		return false, err
	}

	if m.Compare == EMPTY {
		m.Compare = EQL
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
			fmt.Sprintf("uncaught case for exploding compare: %s", m.Compare),
		}
		return false, err
	}
}
