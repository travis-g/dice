package dice

import (
	"context"
	"fmt"
)

var _ = Interface(&PolyhedralDie{})

// A PolyhedralDie represents a variable-sided die in memory, including the result of
// rolling it.
type PolyhedralDie struct {
	Type      string  `json:"type"`
	Result    float64 `json:"result"`
	Size      int     `json:"size"`
	Dropped   bool    `json:"dropped,omitempty"`
	Unrolled  bool    `json:"unrolled,omitempty"`
	Modifiers []Modifier
}

// String returns an expression-like representation of a rolled die or its type,
// if it has not been rolled.
func (d *PolyhedralDie) String() string {
	if !d.Unrolled {
		t, _ := d.Total(context.Background())
		return fmt.Sprintf("%v", t)
	}
	return d.Type
}

// GoString prints the Go syntax of a die.
func (d *PolyhedralDie) GoString() string {
	return fmt.Sprintf("%#v", *d)
}

// Total implements the dice.Interface Total method.
func (d *PolyhedralDie) Total(ctx context.Context) (float64, error) {
	if d.Dropped {
		return 0.0, nil
	}
	if d.Unrolled {
		d.Roll(ctx)
	}
	return d.Result, nil
}

// Roll implements the dice.Interface Roll method. Results for polyhedral dice
// are in the range [1, size].
func (d *PolyhedralDie) Roll(ctx context.Context) (float64, error) {
	if !d.Unrolled {
		return d.Result, nil
	}
	if d.Result == 0 {
		i, err := Intn(d.Size)
		if err != nil {
			return 0, err
		}
		d.Result = float64(1 + i)
		d.Unrolled = false
	}
	// TODO: process modifiers that may result in a reroll
	// for _, mod := range d.Modifiers {
	// 	mod.Apply(ctx, d)
	// }
	return d.Result, nil
}
