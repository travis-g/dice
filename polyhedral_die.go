package dice

import (
	"context"
	"fmt"
)

// A PolyhedralDie represents a variable-sided die in memory, including the result of
// rolling it.
type PolyhedralDie struct {
	// Generic properties
	Result    *int         `json:"result" mapstructure:"result"`
	Size      int          `json:"size" mapstructure:"size"`
	Dropped   bool         `json:"dropped,omitempty" mapstructure:"dropped"`
	Modifiers ModifierList `json:"modifiers,omitempty" mapstructure:"modifiers"`
}

// NewPolyhedralDie creates a new standard polyhedral die and returns it.
func NewPolyhedralDie(props *RollerProperties) (new Roller, err error) {
	var result *int
	if props.Result != nil {
		*result = int(*props.Result)
	}
	new = &PolyhedralDie{
		Result:    result,
		Size:      int(props.Size),
		Dropped:   props.Dropped,
		Modifiers: props.DieModifiers,
	}
	return
}

// String returns an expression-like representation of a rolled die or its
// notation/type, if it has not been rolled.
func (d *PolyhedralDie) String() string {
	if d.Result != nil {
		total, _ := d.Total()
		return fmt.Sprintf("%v", total)
	}
	return fmt.Sprintf("d%d%s", d.Size, d.Modifiers)
}

// GoString prints the Go syntax of a die.
func (d *PolyhedralDie) GoString() string {
	return fmt.Sprintf("%#v", d)
}

// Total implements the dice.Interface Total method.
func (d *PolyhedralDie) Total() (float64, error) {
	if d.Result == nil {
		return 0.0, ErrUnrolled
	}
	if d.Dropped {
		return 0.0, nil
	}
	return float64(*d.Result), nil
}

// Roll implements the dice.Interface Roll method. Results for polyhedral dice
// are in the range [1, size].
func (d *PolyhedralDie) Roll(ctx context.Context) error {
	// Return an error if the Die had been rolled
	if d.Result != nil {
		return ErrRolled
	}

	if err := d.roll(); err != nil {
		return err
	}

	for _, mod := range d.Modifiers {
		if err := mod.Apply(ctx, d); err != nil {
			return err
		}
	}
	return nil
}

// Reroll implements the Roller interaface's Reroll method be recalculating the
// die's result.
func (d *PolyhedralDie) Reroll(ctx context.Context) error {
	if d.Result == nil {
		return ErrUnrolled
	}
	d.Result = nil
	d.Dropped = false
	return d.Roll(ctx)
}

func (d *PolyhedralDie) roll() error {
	i := 1 + Source.Intn(d.Size)
	d.Result = &i
	return nil
}

// Drop marks a PolyhedralDie as dropped.
func (d *PolyhedralDie) Drop(_ context.Context, dropped bool) {
	d.Dropped = dropped
}
