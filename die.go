package dice

import (
	"context"
	"fmt"
)

// Die represents an internally-typed die.
type Die struct {
	// Generic properties
	Type      DieType      `json:"type,omitempty" mapstructure:"type"`
	Size      uint         `json:"size" mapstructure:"size"`
	Result    *float64     `json:"result,omitempty" mapstructure:"result"`
	Dropped   bool         `json:"dropped,omitempty" mapstructure:"dropped"`
	Modifiers ModifierList `json:"modifiers,omitempty" mapstructure:"modifiers"`
}

// NewDie TODO
func NewDie(props *RollerProperties) (Roller, error) {
	die := &Die{
		Type:      props.Type,
		Size:      props.Size,
		Result:    props.Result,
		Dropped:   props.Dropped,
		Modifiers: props.DieModifiers,
	}
	return die, nil
}

// rolls a die based on the die's size and type.
func (d *Die) roll(ctx context.Context) error {
	switch d.Type {
	case TypeFudge:
		i := float64(Source.Intn(int(d.Size*2+1)) - int(d.Size))
		d.Result = &i
	default:
		i := float64(1 + Source.Intn(int(d.Size)))
		d.Result = &i
	}
	return nil
}

// reset resets a Die's properties so that it can be re-rolled.
func (d *Die) reset() {
	d.Result = nil
	d.Dropped = false
}

// Roll rolls the Die. The error returned will be an ErrRolled error if the die
// was already rolled.
func (d *Die) Roll(ctx context.Context) error {
	// Return an error if the Die had been rolled
	if d.Result != nil {
		return ErrRolled
	}

	err := d.roll(ctx)
	if err != nil {
		return err
	}

	// Apply modifiers
	for _, mod := range d.Modifiers {
		if err = mod.Apply(ctx, d); err != nil {
			return err
		}
	}
	return nil
}

// Reroll performs a reroll after resetting a Die.
func (d *Die) Reroll(ctx context.Context) error {
	if d.Result == nil {
		return ErrUnrolled
	}
	d.reset()
	return d.roll(ctx)
}

// String returns an expression-like representation of a rolled die or its type,
// if it has not been rolled.
func (d *Die) String() string {
	if d.Result != nil {
		total, _ := d.Total(context.TODO())
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

// Total implements the dice.Interface Total method. An ErrUnrolled error will
// be returned if the die has not been rolled.
func (d *Die) Total(_ context.Context) (float64, error) {
	if d.Result == nil {
		return 0.0, ErrUnrolled
	}
	if d.Dropped {
		return 0.0, nil
	}
	return *d.Result, nil
}

// Drop marks a Die as dropped.
func (d *Die) Drop(_ context.Context, dropped bool) {
	d.Dropped = dropped
}
