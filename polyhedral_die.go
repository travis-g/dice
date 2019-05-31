package dice

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

var _ = Roller(&PolyhedralDie{})

// A PolyhedralDie represents a variable-sided die in memory, including the result of
// rolling it.
type PolyhedralDie struct {
	sync.RWMutex

	// Rolled state. Handle changes atomically.
	rolled uint32

	// Generic properties
	Result    *int         `json:"result"`
	Size      int          `json:"size"`
	Dropped   bool         `json:"dropped,omitempty"`
	Modifiers ModifierList `json:"modifiers,omitempty"`
}

// String returns an expression-like representation of a rolled die or its
// notation/type, if it has not been rolled.
func (d *PolyhedralDie) String() string {
	d.RLock()
	defer d.RUnlock()
	if d.rolled == 1 || d.Result != nil {
		return fmt.Sprintf("%v", *d.Result)
	}
	return fmt.Sprintf("d%d%s", d.Size, d.Modifiers)
}

// GoString prints the Go syntax of a die.
func (d *PolyhedralDie) GoString() string {
	d.RLock()
	defer d.RUnlock()
	return fmt.Sprintf("%#v", d)
}

// Total implements the dice.Interface Total method.
func (d *PolyhedralDie) Total(ctx context.Context) (float64, error) {
	d.RLock()
	defer d.RUnlock()
	if d.rolled == 0 && d.Result == nil {
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
	d.Lock()
	defer d.Unlock()

	// Return an error if the Die had been rolled
	if d.rolled == 1 {
		return ErrRolled
	}

	err := d.roll()
	if err != nil {
		return err
	}

	// for _, mod := range d.Modifiers {
	// 	err := mod.Apply(ctx, d)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

func (d *PolyhedralDie) roll() error {
	if ok := atomic.CompareAndSwapUint32(&d.rolled, 0, 1); !ok {
		return ErrRolled
	}
	i := 1 + Source.Intn(int(d.Size))
	d.Result = &i
	return nil
}
