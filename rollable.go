package dice

// A Rollable is any kind of rollable object. A Rollable could be a single die
// or many dice of any type.
type Rollable interface {
	// Roll should be used to also set the object's Result
	Roll() (float64, error)
	String() string
	Type() string
	GoString() string
}

// Roll rolls a set of rollables and returns the total.
func Roll(rollables ...Rollable) (float64, error) {
	for _, r := range rollables {
		r.Roll()
	}
	return sumRollables(rollables...)
}

func sumRollables(rollables ...Rollable) (float64, error) {
	sum := 0.0
	for _, r := range rollables {
		i, err := r.Roll()
		if err != nil {
			return 0, err
		}
		sum += i
	}
	return sum, nil
}

// A RollableSet are sets of Rollables
type RollableSet interface {
	Rollable
}

// Expand returns the expanded representation of a set based on the set's type.
func Expand(set RollableSet) string {
	switch t := set.(type) {
	case *DieSet:
		return t.Expanded
	case *FateDieSet:
		return t.Expanded
	default:
		return ""
	}
}
