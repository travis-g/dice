package dice

import (
	"fmt"
)

const fateDieNotation = "dF"

var _ = Interface(&FateDie{})

// A FateDie (a.k.a. "Fudge die") is a die with six sides, {-1, -1, 0, 0, 1, 1}.
// In a pinch, a FateDie can be emulated by evaluating `1d3-2`.
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

// Roll will Roll a given FateDie and set the die's result. Fate dice can have
// results in [-1, 1].
func (f *FateDie) Roll() (float64, error) {
	if !f.Unrolled {
		return float64(f.Result), nil
	}
	i, err := Intn(3)
	if err != nil {
		return 0, err
	}
	f.Result = i - 1
	f.Unrolled = false
	return float64(f.Result), nil
}

// Total returns the result of a Fate die. If dropped, 0 is returned.
func (f *FateDie) Total() float64 {
	if f.Dropped {
		return 0.0
	}
	if f.Unrolled {
		f.Roll()
	}
	return float64(f.Result)
}
