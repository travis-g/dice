package dice

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

// Die represents a typed die. A Die should use a mutex lock for thread safety
// and call `Reroll()` manually to prevent unintended re-rolling/possible wastes
// of system entropy.
//
// At its core, Die should be handled as an RWMutex: depending on state
// (rolled/settled, mid-roll, etc.) it may or may not be safe to read the die's
// properties at any specific time. When a Die is being read, RLock() it to
// prevent writes. When it's being modified, Lock() it to prevent race condition
// other writes as well as reads.
//
// The thread safety of any given Die should be left up to the implementer to
// check, as oftentimes an individual die is rolled once, returned, and printed
// synchronously. It is only when dice are cached, monitored, etc. that thread
// safety is required.
type Die struct {
	// embed an RWMutex's properties/methods
	sync.RWMutex

	// Rolled state and the count of total rolls. Handle changes atomically.
	rolled uint32
	rolls  uint32

	// Generic properties
	Type      DieType      `json:"type,omitempty"`
	Size      int          `json:"size"`
	Result    float64      `json:"result,omitempty"`
	Dropped   bool         `json:"dropped,omitempty"`
	Modifiers ModifierList `json:"-"`
}

// A DieProperties object is the set of properties (usually extracted from a
// notation) that should be used to define a Die or group of like dice (a slice
// of multiple Die).
type DieProperties struct {
	Type    DieType `json:"type,omitempty"`
	Size    int     `json:"size,omitempty"`
	Result  float64 `json:"result,omitempty"`
	Dropped bool    `json:"dropped,omitempty"`

	// Modifiers for the dice or parent set
	DieModifiers   ModifierList `json:"die_modifiers,omitempty"`
	GroupModifiers ModifierList `json:"group_modifiers,omitempty"`
}

// NewDie create a new Die to roll off of a supplied property set.
func NewDie(props *DieProperties) (*Die, error) {
	if props.Size == 0 {
		return nil, ErrSizeZero
	}
	if props.Type == TypeMultiple {
		return nil, fmt.Errorf("cannot create single die of type TypeMultiple")
	}
	return &Die{
		Type:      props.Type,
		Size:      props.Size,
		Result:    props.Result,
		Dropped:   props.Dropped,
		Modifiers: props.DieModifiers,
	}, nil
}

// Roll implements the Roller interface and is thread-safe. The error returned
// will be an ErrRolled error if the die was already rolled. The Roll function
// should be what checks any context maximums, as this is the function that
// gatekeeps entropy use (net new rolls, rerolls, etc.).
//
// Note: In order for the modifiers to mutate the die (ex. rerolls), the die
// must be unlocked, which may lead to thread safety issues.
func (d *Die) Roll(ctx context.Context) (float64, error) {
	// wait until we can safely roll the die, then re-lock the mutex
	d.Lock()
	// if die was already rolled, return its existing roll and an error and
	// defer unlocking the die
	if d.rolled == 1 {
		defer d.Unlock()
		return d.Result, ErrRolled
	}

	err := roll(ctx, d)
	if err != nil {
		fmt.Println(err)
		defer d.Unlock()
		return d.Result, err
	}

	d.Unlock()
	for _, mod := range d.Modifiers {
		mod.Apply(ctx, d)
	}
	return d.Result, nil
}

// rolls a die based on the die's Size. This does not ensure thread-safety: the
// die's mutex should be locked before attempting to roll.
func roll(ctx context.Context, d *Die) error {
	atomic.AddUint32(&d.rolls, 1)
	if ok := atomic.CompareAndSwapUint32(&d.rolled, 0, 1); !ok {
		return ErrRolled
	}

	switch d.Type {
	case TypeFudge:
		i, err := Intn(d.Size*2 + 1)
		if err != nil {
			return err
		}
		d.Result = float64(i - d.Size)
	default:
		i, err := Intn(d.Size)
		if err != nil {
			return err
		}
		d.Result = float64(1 + i)
	}
	return nil
}

// Reroll performs a thread-safe reroll after resetting a Die.
func (d *Die) Reroll(ctx context.Context) (float64, error) {
	d.Lock()
	defer d.Unlock()
	d.reset()
	err := roll(ctx, d)
	return d.Result, err
}

// reroll performs a thread unsafe reroll.
func (d *Die) reroll(ctx context.Context) (err error) {
	d.reset()
	err = roll(ctx, d)
	return
}

// reset resets a Die's properties so that it can be re-rolled.
func (d *Die) reset() {
	d.rolled = 0
	d.Result = 0
	d.Dropped = false
}

// String returns an expression-like representation of a rolled die or its type,
// if it has not been rolled.
func (d *Die) String() string {
	d.RLock()
	defer d.RUnlock()
	if d.rolled == 1 {
		return fmt.Sprintf("%v", d.Result)
	}
	switch d.Type {
	case TypePolyhedron:
		return fmt.Sprintf("d%d", d.Size)
	case TypeFudge:
		if d.Size == 1 {
			return "dF"
		}
		return fmt.Sprintf("f%d", d.Size)
	default:
		return d.Type.String()
	}
}

// Total implements the dice.Interface Total method. An ErrUnrolled error will
// be returned if the die has not been rolled.
func (d *Die) Total(ctx context.Context) (float64, error) {
	d.RLock()
	defer d.RUnlock()
	if d.rolled == 0 {
		return 0.0, ErrUnrolled
	}
	if d.Dropped {
		return 0.0, nil
	}
	return d.Result, nil
}
