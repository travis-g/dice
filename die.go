package dice

import (
	"context"
	"fmt"

	"github.com/matryer/resync"
)

var _ = Interface(&Die{})

// Die represents a polyhedral die (possibly a Fate die) that uses mutexes to
// prevent unintended re-rolling.
type Die struct {
	Type      DieType `json:"type"`
	Size      int     `json:"size"`
	Result    float64 `json:"result"`
	Dropped   bool    `json:"dropped"`
	Modifiers ModifierList
	rolled    *resync.Once
}

// rolls a die based on the die's Size
func roll(d *Die) func() {
	return func() {
		i, err := Intn(d.Size)
		if err == nil {
			d.Result = float64(1 + i)
		}
	}
}

// NewDie create a new Die to roll
func NewDie(size int) *Die {
	return &Die{
		Size:   size,
		Type:   TypePolyhedron,
		rolled: nil,
	}
}

// Roll implements the dice.Interface Roll method. Results for polyhedral dice
// are in the range [1, size].
func (d *Die) Roll(ctx context.Context) (float64, error) {
	// create a mutex if needed
	if d.rolled == nil {
		d.rolled = new(resync.Once)
	}
	d.rolled.Do(roll(d))
	for _, mod := range d.Modifiers {
		mod.Apply(ctx, d)
	}
	return d.Result, nil
}

// Reroll unsets the mutex of the Die and rerolls it
func (d *Die) Reroll(ctx context.Context) (float64, error) {
	// reset the mutex before rerolling
	d.rolled.Reset()
	return d.Roll(ctx)
}

// String returns an expression-like representation of a rolled die or its type,
// if it has not been rolled.
func (d *Die) String() string {
	var ctx = context.TODO()
	if d.rolled != nil {
		t, _ := d.Total(ctx)
		return fmt.Sprintf("%v", t)
	}
	switch d.Type {
	case TypePolyhedron:
		return fmt.Sprintf("d%d%s", d.Size, d.Modifiers)
	case TypeFate:
		return "dF"
	default:
		return d.Type.String()
	}
}

// GoString prints the Go syntax of a die.
func (d *Die) GoString() string {
	return fmt.Sprintf("%#v", *d)
}

// Total implements the dice.Interface Total method.
func (d *Die) Total(ctx context.Context) (float64, error) {
	if d.Dropped {
		return 0.0, nil
	}
	return d.Result, nil
}
