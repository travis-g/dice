package dice

import (
	"context"
	"fmt"
)

const fateDieNotation = "dF"

// A FudgeDie is a die with six sides, {-1, -1, 0, 0, 1, 1}. A FudgeDie can be
// emulated with a traditional polyhedral die by evaluating "1d3-2".
type FudgeDie struct {
	Result  *int   `json:"result"`
	Type    string `json:"type"`
	Dropped bool   `json:"dropped,omitempty"`
}

func (f *FudgeDie) String() string {
	if f.Result != nil {
		return fmt.Sprintf("%v", *f.Result)
	}
	return fateDieNotation
}

// GoString prints a viable golang code representation of a FateDie.
func (f *FudgeDie) GoString() string {
	return fmt.Sprintf("%#v", *f)
}

// Roll implements the dice.Interface Roll method. Fate dice can have integer
// results in [-1, 1].
func (f *FudgeDie) Roll(ctx context.Context) error {
	if f.Result != nil {
		return ErrRolled
	}
	i := Source.Intn(3) - 1
	f.Result = &i
	return nil
}

// Reroll implements the Roller interaface's Reroll method be recalculating the
// die's result.
func (f *FudgeDie) Reroll(ctx context.Context) error {
	if f.Result == nil {
		return ErrUnrolled
	}
	i := Source.Intn(3) - 1
	f.Result = &i
	return nil
}

// Total implements the dice.Interface Total method. If dropped, 0 is returned.
func (f *FudgeDie) Total(ctx context.Context) (float64, error) {
	var err error
	if f.Result == nil {
		err = f.Roll(ctx)
	}
	if f.Dropped {
		return 0.0, nil
	}
	return float64(*f.Result), err
}
