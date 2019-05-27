package dice

type rwMutexer interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

// ensure Die can be implemented like a mutex for thread safety
var _ = rwMutexer(&Die{})

// ensure Die is a rollable Interface
var _ = Interface(&Die{})
