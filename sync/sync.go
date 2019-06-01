package sync

import (
	"context"
	"sync"

	"github.com/travis-g/dice"
)

// RWMutexRoller is a Roller wrapped with a sync.RWMutex.
type RWMutexRoller struct {
	l   sync.RWMutex
	die dice.Roller
}

// Wrap creates an RWMutexRoller by wrapping a Roller with a sync.RWMutex.
func Wrap(die dice.Roller) *RWMutexRoller {
	return &RWMutexRoller{
		die: die,
	}
}

// Roll rolls the embedded Roller with thread safety. Rollers can call their
// Roll() function to calculate their result with thread safety.
func (r *RWMutexRoller) Roll(ctx context.Context) error {
	r.l.Lock()
	defer r.l.Unlock()
	return r.die.Roll(ctx)
}

// Reroll rolls the embedded Roller with thread safety.
func (r *RWMutexRoller) Reroll(ctx context.Context) error {
	r.l.Lock()
	defer r.l.Unlock()
	return r.die.Reroll(ctx)
}

// Total read-locks the embedded Roller and returns the total.
func (r *RWMutexRoller) Total(ctx context.Context) (float64, error) {
	r.l.RLock()
	defer r.l.RUnlock()
	return r.die.Total(ctx)
}

// String read-locks the embedded Roller and returns the Roller's string
// representation.
func (r *RWMutexRoller) String() string {
	r.l.RLock()
	defer r.l.RUnlock()
	return r.die.String()
}

// Lock locks the mutex of RWMutexRoller.
func (r *RWMutexRoller) Lock() {
	r.l.Lock()
}

// Unlock unlocks the mutex of RWMutexRoller.
func (r *RWMutexRoller) Unlock() {
	r.l.Unlock()
}

// RLock read-locks the mutex of RWMutexRoller.
func (r *RWMutexRoller) RLock() {
	r.l.RLock()
}

// RUnlock read-unlocks the mutex of RWMutexRoller.
func (r *RWMutexRoller) RUnlock() {
	r.l.RUnlock()
}
