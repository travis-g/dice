package dice

import (
	"context"
	"fmt"
)

const fateDieNotation = "dF"

// A FudgeDie is a die with six sides, {-1, -1, 0, 0, 1, 1}. A FudgeDie can be
// emulated with a traditional polyhedral die by evaluating "1d3-2".
type FudgeDie struct {
	Result    *int         `json:"result"`
	Dropped   bool         `json:"dropped,omitempty"`
	Modifiers ModifierList `json:"modifiers,omitempty" mapstructure:"modifiers"`
}

// NewFudgeDie creates a new fudge die from a list of properties.
func NewFudgeDie(props *RollerProperties) (Roller, error) {
	var result *int
	if props.Result != nil {
		*result = int(*props.Result)
	}
	return &FudgeDie{
		Result:    result,
		Dropped:   props.Dropped,
		Modifiers: props.DieModifiers,
	}, nil
}

func (f *FudgeDie) String() string {
	if f.Result != nil {
		total, _ := f.Total()
		return fmt.Sprintf("%v", total)
	}
	return fateDieNotation
}

// GoString prints a viable golang code representation of a FateDie.
func (f *FudgeDie) GoString() string {
	return fmt.Sprintf("%#v", *f)
}

func (f *FudgeDie) roll(ctx context.Context) (err error) {
	i := Source.Intn(3) - 1
	f.Result = &i
	return
}

func (f *FudgeDie) reset() {
	f.Result = nil
	f.Dropped = false
}

// Roll implements the dice.Interface Roll method. Fate dice can have integer
// results in [-1, 1].
func (f *FudgeDie) Roll(ctx context.Context) error {
	if f.Result != nil {
		return ErrRolled
	}
	err := f.roll(ctx)
	if err != nil {
		return err
	}

	// Apply modifiers
	for _, mod := range f.Modifiers {
		mod.Apply(ctx, f)
	}
	return nil
}

// Reroll implements the Roller interaface's Reroll method be recalculating the
// die's result.
func (f *FudgeDie) Reroll(ctx context.Context) error {
	if f.Result == nil {
		return ErrUnrolled
	}
	f.reset()
	return f.roll(ctx)
}

// Total implements the dice.Interface Total method. If dropped, 0 is returned.
func (f *FudgeDie) Total() (float64, error) {
	var err error
	if f.Result == nil {
		return 0.0, ErrUnrolled
	}
	if f.Dropped {
		return 0.0, nil
	}
	return float64(*f.Result), err
}

// Drop marks a FudgeDie as dropped.
func (f *FudgeDie) Drop(_ context.Context, dropped bool) {
	f.Dropped = dropped
}
