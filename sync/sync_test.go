package sync

// ensure RWMutexRoller can be implemented like a Roller and an RWMutex for
// thread safety
var _ = RWLockerRoller(&RWMutexRoller{})
