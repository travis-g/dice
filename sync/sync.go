/*
Package sync implements a thread-safe wrapper for rollable dice.
*/
package sync

import (
	"context"
	"sync"

	"github.com/travis-g/dice"
)

// RWLockerRoller is implemented by any value that implements dice.Roller,
// sync.Locker, and has an RLock and RUnlock method.
type RWLockerRoller interface {
	dice.Roller
	sync.Locker
	RLock()
	RUnlock()
}

// RWMutexRoller is a dice.Roller wrapped with a sync.RWMutex. The methods of
// RWMutexRoller call the embedded Roller's methods within a thread-safe
// context.
type RWMutexRoller struct {
	l   sync.RWMutex
	die dice.Roller
}

// Wrap creates an RWMutexRoller out of a Roller by wrapping it with a
// sync.RWMutex.
func Wrap(die dice.Roller) *RWMutexRoller {
	return &RWMutexRoller{
		die: die,
	}
}

// Roll rolls the embedded Roller with thread safety. It calls the embedded
// Roller's Roll() method to calculate its result.
func (r *RWMutexRoller) Roll(ctx context.Context) error {
	r.l.Lock()
	defer r.l.Unlock()
	return r.die.Roll(ctx)
}

// Reroll re-rolls the embedded Roller with thread safety.
func (r *RWMutexRoller) Reroll(ctx context.Context) error {
	r.l.Lock()
	defer r.l.Unlock()
	return r.die.Reroll(ctx)
}

// Total read-locks the embedded Roller and returns its total.
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
