package dice

import (
	"context"
	"fmt"
)

const fateDieNotation = "dF"

var _ = Interface(&FateDie{})

// A FateDie is a die with six sides, {-1, -1, 0, 0, 1, 1}. A FateDie can be
// emulated with a traditional polyhedral die by evaluating "1d3-2".
type FateDie struct {
	Result   int    `json:"result"`
	Type     string `json:"type"`
	Dropped  bool   `json:"dropped,omitempty"`
	Unrolled bool   `json:"unrolled,omitempty"`
}

func (f *FateDie) String() string {
	if !f.Unrolled {
		return fmt.Sprintf("%v", f.Result)
	}
	return fateDieNotation
}

// GoString prints a viable golang code representation of a FateDie.
func (f *FateDie) GoString() string {
	return fmt.Sprintf("%#v", *f)
}

// Roll implements the dice.Interface Roll method. Fate dice can have integer
// results in [-1, 1].
func (f *FateDie) Roll(ctx context.Context) (float64, error) {
	if !f.Unrolled {
		t, err := f.Total(ctx)
		return float64(t), err
	}
	i, err := Intn(3)
	if err != nil {
		return 0, err
	}
	f.Result = i - 1
	f.Unrolled = false
	return float64(f.Result), nil
}

// Total implements the dice.Interface Total method. If dropped, 0 is returned.
// Note that the Dropped bool itself should be checked to ensure the fate die
// was indeed dropped, and did not simply roll a 0.
func (f *FateDie) Total(ctx context.Context) (float64, error) {
	if f.Dropped {
		return 0.0, nil
	}
	var err error
	if f.Unrolled {
		_, err = f.Roll(ctx)
	}
	return float64(f.Result), err
}
