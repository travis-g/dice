package dice

import "sync"

type rwMutexer interface {
	sync.Locker
	RLock()
	RUnlock()
}

// ensure Die can be implemented like an RWMutex for thread safety
var _ = rwMutexer(&Die{})

// ensure Die implements Roller
var _ Roller = (*Die)(nil)
