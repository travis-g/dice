package dice

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
