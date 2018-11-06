package dice

type ErrNotImplemented struct {
	message string
}

func NewErrNotImplemented(message string) *ErrNotImplemented {
	return &ErrNotImplemented{
		message: message,
	}
}

func (e *ErrNotImplemented) Error() string {
	return e.message
}

type ErrParseError struct {
	Notation     string
	NotationElem string
	ValueElem    string
	Message      string
}

type ErrImpossibleDie struct {
	message string
}

func NewImpossibleDie(message string) *ErrImpossibleDie {
	return &ErrImpossibleDie{
		message: message,
	}
}

func (e *ErrImpossibleDie) Error() string {
	return e.message
}
