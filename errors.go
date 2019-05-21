package dice

import "errors"

var (
	// ErrSizeZero is returned when an attempt to create or roll a 0-sized die
	// is made.
	ErrSizeZero = errors.New("die cannot have 0 sides")

	// ErrUnrolled is returned when an operation that requires a rolled die is
	// preformed on an unrolled die
	ErrUnrolled = errors.New("die is unrolled")
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
