package dice

import (
	"context"
	"fmt"
)

// Die represents an internally-typed die.
type Die struct {
	// Generic properties
	Type      DieType      `json:"type,omitempty"`
	Size      uint         `json:"size"`
	Result    *float64     `json:"result,omitempty"`
	Dropped   bool         `json:"dropped,omitempty"`
	Modifiers ModifierList `json:"modifiers,omitempty"`
}

// A DieProperties object is the set of properties (usually extracted from a
// notation) that should be used to define a Die or group of like dice (a slice
// of multiple Die).
type DieProperties struct {
	Type    DieType  `json:"type,omitempty"`
	Size    uint     `json:"size,omitempty"`
	Result  *float64 `json:"result,omitempty"`
	Rolled  bool     `json:"rolled,omitempty"`
	Dropped bool     `json:"dropped,omitempty"`

	// Modifiers for the dice or parent set
	DieModifiers   ModifierList `json:"die_modifiers,omitempty"`
	GroupModifiers ModifierList `json:"group_modifiers,omitempty"`
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
		mod.Apply(ctx, d)
	}
	return nil
}

// rolls a die based on the die's Size.
func (d *Die) roll(ctx context.Context) error {
	if d.Result != nil {
		return ErrRolled
	}
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

// Reroll performs a reroll after resetting a Die.
func (d *Die) Reroll(ctx context.Context) error {
	d.reset()
	return d.roll(ctx)
}

// reset resets a Die's properties so that it can be re-rolled.
func (d *Die) reset() {
	d.Result = nil
	d.Dropped = false
}

// String returns an expression-like representation of a rolled die or its type,
// if it has not been rolled.
func (d *Die) String() string {
	if d.Result != nil {
		return fmt.Sprintf("%v", *d.Result)
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
