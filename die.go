package dice

import (
	"context"
	"fmt"
)

// Die represents an internally-typed die. If Result is a non-nil pointer, it
// is considered rolled.
type Die struct {
	// Generic properties
	Type DieType `json:"type,omitempty" mapstructure:"type"`
	Size int     `json:"size" mapstructure:"size"`

	rolls   []*Result
	*Result `json:"result,omitempty" mapstructure:"result"`

	Modifiers ModifierList `json:"modifiers,omitempty" mapstructure:"modifiers"`
}

// NewDie creates a new die off of a properties list. It will tweak the
// properties list to better suit reuse.
func NewDie(props *RollerProperties) (Roller, error) {
	// If the property set was for a default fudge die set, set a default size
	// of 1.
	if props.Type == TypeFudge && props.Size == 0 {
		props.Size = 1
	}

	// Check if size was zero and it's not a fudge die
	if props.Size == 0 {
		return nil, ErrSizeZero
	}

	die := &Die{
		Type:      props.Type,
		Size:      props.Size,
		Result:    props.Result,
		Modifiers: props.DieModifiers,
	}
	return die, nil
}

// Roll rolls a die based on the die's size and type and calculates a value.
func (d *Die) Roll(ctx context.Context) error {
	if d == nil {
		return ErrNilDie
	}

	// Check if rolled too many times already
	if len(d.rolls) >= MaxRerolls {
		return ErrMaxRolls
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

// reset resets a Die's properties so that it can be re-rolled.
func (d *Die) reset() {
	d.Result = nil
	d.Dropped = false
	d.rolls = []*Result{}
}

// FullRoll rolls the Die. The die will be reset if it had been rolled
// previously.
func (d *Die) FullRoll(ctx context.Context) error {
	if d == nil {
		return ErrNilDie
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
	// mark the current result as dropped, move it to the roll history, and
	// unset the result
	d.Result.Drop(ctx, true)
	d.rolls = append(d.rolls, d.Result)
	d.Result = nil
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
		return fmt.Sprintf("%v", total)
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
