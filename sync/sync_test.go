package sync

import (
	"sync"

	"github.com/travis-g/dice"
)

type rwMutexer interface {
	sync.Locker
	RLock()
	RUnlock()
}

// ensure RWMutexRoller can be implemented like an RWMutex for thread safety
var _ = rwMutexer(&RWMutexRoller{})

// ensure RWMutexRoller implements Roller
var _ dice.Roller = (*RWMutexRoller)(nil)
