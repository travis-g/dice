package dice

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"
)

// Die represents an internally-typed die. If Result is a non-nil pointer, it
// is considered rolled.
type Die struct {
	// Generic properties
	Type DieType `json:"type,omitempty" mapstructure:"type"`
	Size int     `json:"size" mapstructure:"size"`

	Rerolls int `json:"rerolls" mapstructure:"rerolls"`
	*Result `json:"result,omitempty" mapstructure:"result"`

	Modifiers ModifierList `json:"modifiers,omitempty" mapstructure:"modifiers"`

	parent Roller
}

// NewDie creates a new die off of a properties list. It will tweak the
// properties list to better suit reuse.
func NewDie(props *RollerProperties, parent Roller) (Roller, error) {
	return NewDieWithParent(props, parent)
}

// NewDieWithParent creates a new die off of a properties list. It will tweak the
// properties list to better suit reuse.
func NewDieWithParent(props *RollerProperties, parent Roller) (Roller, error) {
	// If the property set was for a default fudge die set, set a default size
	// of 1.
	if props.Type == TypeFudge && props.Size == 0 {
		props.Size = 1
	}

	die := &Die{
		Type:      props.Type,
		Size:      props.Size,
		Result:    props.Result,
		Modifiers: props.DieModifiers,
	}

	if parent != nil {
		die.SetParent(parent)
	}
	return die, nil
}

// Roll rolls a die based on the die's size and type and calculates a value.
func (d *Die) Roll(ctx context.Context) error {
	if d == nil {
		return ErrNilDie
	}

	// Check if rolled too many times already
	var maxRolls = int64(MaxRolls)
	if ctxMaxRolls, ok := ctx.Value(CtxKeyMaxRolls).(int64); ok {
		maxRolls = ctxMaxRolls
	}

	if *TotalRolls(ctx) >= uint64(maxRolls) {
		return ErrMaxRolls
	}

	// bump context roll count
	atomic.AddUint64(TotalRolls(ctx), 1)

	if d.Size == 0 {
		d.Result = NewResult(0)
		return nil
	}

	switch d.Type {
	case TypeFudge:
		d.Result = NewResult(float64(Source.Intn(int(d.Size*2+1)) - int(d.Size)))
		if d.Result.Value == -float64(d.Size) {
			d.CritFailure = true
		}
	default:
		d.Result = NewResult(float64(1 + Source.Intn(int(d.Size))))
		if d.Result.Value == 1 {
			d.CritFailure = true
		}
	}
	// default critical success on max roll; override via modifiers
	if d.Result.Value == float64(d.Size) {
		d.CritSuccess = true
	}

	return nil
}

// reset resets a Die's properties so that it can be re-rolled from scratch.
func (d *Die) reset() {
	d.Result = nil
	d.Dropped = false
}

// FullRoll rolls the Die. The die will be reset if it had been rolled
// previously.
func (d *Die) FullRoll(ctx context.Context) error {
	if d == nil {
		return ErrNilDie
	}

	// Check if rolled too many times already
	var maxRolls = MaxRolls
	if ctxMaxRolls, ok := ctx.Value(CtxKeyMaxRolls).(uint64); ok {
		maxRolls = ctxMaxRolls
	}

	if *TotalRolls(ctx) >= uint64(maxRolls) {
		return ErrMaxRolls
	}

	if err := d.Roll(ctx); err != nil {
		return err
	}

	// Apply modifiers
	for i := 0; i < len(d.Modifiers); i++ {
		err := d.Modifiers[i].Apply(ctx, d)
		switch {
		// die rerolled, so restart validation checks with new modifiers
		case err == ErrRerolled:
			i = -1
			// i++ => 0 to restart from first modifier
			break
		case err != nil:
			return err
		}
	}
	return nil
}

// Reroll performs a reroll after resetting a Die.
func (d *Die) Reroll(ctx context.Context) error {
	if d == nil {
		return ErrNilDie
	}
	if d.Result == nil {
		return ErrUnrolled
	}

	d.Result = nil
	d.Rerolls++
	// reroll without reapplying all modifiers
	return d.Roll(ctx)
}

// String returns an expression-like representation of a rolled die or its type,
// if it has not been rolled.
func (d *Die) String() string {
	if d == nil {
		return ""
	}
	if d.Result != nil {
		total, _ := d.Total(context.Background())
		return fmt.Sprintf("%.0f", total)
	}
	switch d.Type {
	case TypePolyhedron:
		return fmt.Sprintf("d%d%s", d.Size, d.Modifiers)
	case TypeFudge:
		if d.Size == 1 {
			return fmt.Sprintf("dF%s", d.Modifiers)
		}
		return fmt.Sprintf("f%d%s", d.Size, d.Modifiers)
	default:
		return d.Type.String()
	}
}

// Total implements the Total method. An ErrUnrolled error will be returned if
// the die has not been rolled.
func (d *Die) Total(ctx context.Context) (float64, error) {
	if d == nil {
		return 0.0, ErrNilDie
	}
	if d.Result == nil {
		return 0.0, ErrUnrolled
	}
	return d.Result.Total(ctx)
}

// Value returns the Result.Value of a Die, regardless of whether the Die was
// dropped.
func (d *Die) Value(ctx context.Context) (float64, error) {
	if d == nil {
		return 0.0, ErrNilDie
	}
	if d.Result == nil {
		return 0.0, ErrUnrolled
	}
	return d.Result.Value, nil
}

// Parent returns the Die's parent, which will be nil if an orphan.
func (d *Die) Parent() Roller {
	return d.parent
}

func (d *Die) SetParent(r Roller) {
	d.parent = r
}

// Add causes a panic as a single Die cannot have a descendent.
func (d *Die) Add(r Roller) {
	panic("impossible action")
}

func (d *Die) ToGraphviz() string {
	if d == nil {
		return ""
	}

	var b bytes.Buffer
	write := fmt.Fprintf
	write(&b, "\"%p\" [label=\"%v\"];\n", d, d.String())
	if d.Parent() != nil {
		fmt.Fprintf(&b, "\"%p\" -> \"%p\" [dir=back style=dotted color=red];\n", d.Parent(), d)
	}
	return b.String()
}
