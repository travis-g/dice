package dice

import (
	"context"
	"fmt"
)

// Die represents a typed die. If Result is a non-nil pointer, it
// is considered rolled.
type Die struct {
	// Generic properties
	Type    DieType   `json:"type,omitempty" mapstructure:"type"`
	Size    int       `json:"size" mapstructure:"size"`
	Rerolls int       `json:"rerolls" mapstructure:"rerolls"`
	Results []*Result `json:"results,omitempty" mapstructure:"results"`
	Dropped bool      `json:"dropped,omitempty" mapstructure:"dropped"`

	RollModifiers ModifierList `json:"roll_modifiers,omitempty" mapstructure:"roll_modifiers"`
	Modifiers     ModifierList `json:"modifiers,omitempty" mapstructure:"modifiers"`
}

// NewDie creates a new die off of a properties list. It will tweak the
// properties list to better suit reuse.
func NewDie(props *RollerProperties) (Roller, error) {
	// If the property set was for a default fudge die set, set a default size
	// of 1.
	if props.Type == TypeFudge && props.Size == 0 {
		props.Size = 1
	}

	die := &Die{
		Type:      props.Type,
		Size:      props.Size,
		Modifiers: props.DieModifiers,
	}
	if props.Result != nil {
		die.Results = []*Result{props.Result}
	}
	if len(props.Results) > 0 {
		die.Results = props.Results
	}
	return die, nil
}

// Roll rolls a die based on the die's size and type and calculates a value.
func (d *Die) Roll(ctx context.Context) error {
	if d == nil {
		return ErrNilDie
	}

	// Check if rolled too many times already
	if d.Rerolls >= MaxRerolls {
		return ErrMaxRolls
	}

	if d.Size == 0 {
		d.Results = []*Result{NewResult(0)}
		return nil
	}

	var r *Result
	switch d.Type {
	case TypeFudge:
		r = NewResult(float64(Source.Intn(int(d.Size*2+1)) - int(d.Size)))
		if r.Value == -float64(d.Size) {
			r.CritFailure = true
		}
		d.Results = append(d.Results, r)
	default:
		r = NewResult(float64(1 + Source.Intn(int(d.Size))))
		if r.Value == 1 {
			r.CritFailure = true
		}
		d.Results = append(d.Results, r)
	}
	// default critical success on max roll; override via modifiers
	if r.Value == float64(d.Size) {
		r.CritSuccess = true
	}

	d.Rerolls++
	return nil
}

// reset resets a Die's properties so that it can be re-rolled.
func (d *Die) reset() {
	d.Results = nil
	d.Dropped = false
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
	if d.Results == nil {
		return ErrUnrolled
	}

	d.Results = nil
	// reroll without reapplying all modifiers
	return d.Roll(ctx)
}

// String returns an expression-like representation of a rolled die or its type,
// if it has not been rolled.
func (d *Die) String() string {
	if d == nil {
		return ""
	}
	if len(d.Results) != 0 {
		total, _ := d.Total(context.Background())
		return fmt.Sprintf("%.0f", total)
	}
	switch d.Type {
	case TypePolyhedron:
		return fmt.Sprintf("d%d%s%s", d.Size, d.RollModifiers, d.Modifiers)
	case TypeFudge:
		if d.Size == 1 {
			return fmt.Sprintf("dF%s%s", d.RollModifiers, d.Modifiers)
		}
		return fmt.Sprintf("f%d%s%s", d.Size, d.RollModifiers, d.Modifiers)
	default:
		return d.Type.String()
	}
}

// Total implements the Total method. An ErrUnrolled error will be returned if
// the die has no Results.
func (d *Die) Total(ctx context.Context) (float64, error) {
	if d == nil {
		return 0.0, ErrNilDie
	}
	if len(d.Results) == 0 {
		return 0.0, ErrUnrolled
	}
	if d.IsDropped(ctx) {
		return 0.0, nil
	}
	sum := 0.0
	for _, r := range d.Results {
		total, err := r.Total(ctx)
		if err != nil {
			return sum, err
		}
		sum += total
	}
	return sum, nil
}

// Value returns the sum of undropped Result.Values of a Die, regardless of
// whether the Die was dropped.
func (d *Die) Value(ctx context.Context) (float64, error) {
	if d == nil {
		return 0.0, ErrNilDie
	}
	if len(d.Results) == 0 {
		return 0.0, ErrUnrolled
	}
	sum := 0.0
	for _, r := range d.Results {
		total, err := r.Total(ctx)
		if err != nil {
			return sum, err
		}
		sum += total
	}
	return sum, nil
}

// Drop marks a die as dropped, indicating all of its Results should be ignored
// from totals.
func (d *Die) Drop(_ context.Context, drop bool) {
	if d != nil {
		d.Dropped = drop
	}
}

// IsDropped returns whether a Die was dropped.
func (d *Die) IsDropped(_ context.Context) bool {
	// If there's no die, the die can't have been dropped, it's unrolled.
	if d == nil {
		return false
	}
	return d.Dropped
}
