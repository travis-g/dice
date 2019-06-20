package dice

import "context"

// A Result is a value a die has rolled. By default, CritSuccess and CritFailure
// should be set to true if the maximum or minumum value of a die is rolled
// respectively, but the range in which a critical success/failure must be
// overridable through modifiers.
type Result struct {
	Value       float64 `json:"value"`
	Dropped     bool    `json:"dropped,omitempty"`
	CritSuccess bool    `json:"crit,omitempty"`
	CritFailure bool    `json:"fumble,omitempty"`
}

// NewResult returns a new un-dropped, non-critical Result.
func NewResult(result float64) *Result {
	return &Result{
		Value: result,
	}
}

// Drop marks a Result as dropped.
func (r *Result) Drop(ctx context.Context, drop bool) {
	r.Dropped = drop
}
