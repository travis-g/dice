package dice

// DieType is the enum of types that a die or dice can be
type DieType string

// Types of dice/dice groups.
const (
	// Concrete dice types: these can be used to instantiate a new rollable.
	TypePolyhedron DieType = ""
	TypeFudge      DieType = "fudge"

	// Meta dice types: these are used to classify rollable groups and unknown
	// dice.
	TypeUnknown DieType = "unknown"
)

func (t DieType) String() string {
	switch t {
	case TypePolyhedron:
		return "polyhedron"
	case TypeFudge:
		return "fudge"
	default:
		return "unknown"
	}
}
