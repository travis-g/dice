package dice

import (
	"context"
	"fmt"
)

const fateDieNotation = "dF"

var _ = Roller(&FudgeDie{})

// A FudgeDie is a die with six sides, {-1, -1, 0, 0, 1, 1}. A FudgeDie can be
// emulated with a traditional polyhedral die by evaluating "1d3-2".
type FudgeDie struct {
	Result   *int   `json:"result"`
	Type     string `json:"type"`
	Dropped  bool   `json:"dropped,omitempty"`
	Unrolled bool   `json:"unrolled,omitempty"`
}

func (f *FudgeDie) String() string {
	if !f.Unrolled {
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
	if !f.Unrolled {
		return nil
	}
	i := Source.Intn(3) - 1
	f.Result = &i
	f.Unrolled = false
	return nil
}

// Total implements the dice.Interface Total method. If dropped, 0 is returned.
// Note that the Dropped bool itself should be checked to ensure the fate die
// was indeed dropped, and did not simply roll a 0.
func (f *FudgeDie) Total(ctx context.Context) (float64, error) {
	if f.Dropped {
		return 0.0, nil
	}
	var err error
	if f.Unrolled {
		err = f.Roll(ctx)
	}
	return float64(*f.Result), err
}
