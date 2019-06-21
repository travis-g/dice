package dice

import "errors"

var (
	// ErrSizeZero is returned when an attempt to create or roll a 0-sized die
	// is made.
	ErrSizeZero = errors.New("die cannot have 0 sides")

	// ErrNilDie is returned when a die's passed reference is nil.
	ErrNilDie = errors.New("nil die passed")

	// ErrMaxRolls is returned when a maximum number of rolls/rerolls has been
	// met or surpassed.
	ErrMaxRolls = errors.New("max rolls reached")

	// ErrUnrolled is returned when an operation that requires a rolled die is
	// preformed on an unrolled die
	ErrUnrolled = errors.New("die is unrolled")

	// ErrRolled is returned when an attempt is made to roll a die that had been
	// rolled already
	ErrRolled = errors.New("die already rolled")

	// ErrRerolled is returned when it must be noted that a die's value has
	// changed due to being rerolled.
	ErrRerolled = errors.New("die rerolled")

	// ErrImpossibleRoll is returned when a given dice roll is deemed
	// impossible, illogical, or will never be able to yield a result. As an
	// example, a roll of "3d6r<6" should return this error at some stage in its
	// evaluation, as no die in the set will ever be able to settle with the
	// given reroll modifier.
	ErrImpossibleRoll = errors.New("dice roll impossible")
)

// ErrNotImplemented is an error returned when a feature is not yet implemented.
type ErrNotImplemented struct {
	message string
}

// NewErrNotImplemented returns a new not implemented error.
func NewErrNotImplemented(message string) *ErrNotImplemented {
	return &ErrNotImplemented{
		message: message,
	}
}

func (e *ErrNotImplemented) Error() string {
	return e.message
}

// ErrParseError is an error encountered when parsing a string into dice
// notation.
type ErrParseError struct {
	Notation     string
	NotationElem string
	ValueElem    string
	Message      string
}

func (e *ErrParseError) Error() string {
	if e.Message == "" {
		return "parsing dice string " +
			quote(e.Notation) + ": cannot parse " +
			quote(e.ValueElem) + " as " +
			quote(e.NotationElem)
	}
	return "parsing dice " +
		quote(e.Notation) + e.Message
}
