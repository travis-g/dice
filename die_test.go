package dice

import "sync"

type rwMutexer interface {
	sync.Locker
	RLock()
	RUnlock()
}

// ensure Die can be implemented like a mutex for thread safety
var _ = rwMutexer(&Die{})

// ensure Die is a rollable Interface
var _ Roller = (*Die)(nil)
