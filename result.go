package dice

import (
	"context"

	"github.com/pkg/errors"
)

// Result-related errors.
var (
	ErrNilResult = errors.New("nil result")
)

// A Result is a value a die has rolled. CritSuccess and CritFailure should be
// set to true (by default) if the maximum or minimum value of a die is rolled
// respectively, but the range in which a critical success/failure must be
// overridable through modifiers.
type Result struct {
	Value       float64 `json:"value"`
	Dropped     bool    `json:"dropped,omitempty"`
	CritSuccess bool    `json:"crit,omitempty"`
	CritFailure bool    `json:"fumble,omitempty"`
}

// A ResultFactory is a function that generates and returns a new Result.
type ResultFactory func() *Result

// NewResult returns a new un-dropped, non-critical Result with a provided
// value.
func NewResult(result float64) *Result {
	return &Result{
		Value: result,
	}
}

// Drop marks a Result as dropped.
func (r *Result) Drop(_ context.Context, drop bool) {
	if r != nil {
		r.Dropped = drop
	}
}

// IsDropped returns whether a Result was dropped. Returns false if the Result
// is nil.
func (r *Result) IsDropped(_ context.Context) bool {
	// If there's no result, there's no die to roll.
	if r == nil {
		return false
	}
	return r.Dropped
}

// Total returns the Result's value or 0 if the result was dropped. If the
// Result is nil an ErrNilResult error is returned.
func (r *Result) Total(ctx context.Context) (float64, error) {
	if r == nil {
		return 0.0, ErrNilResult
	}
	if r.IsDropped(ctx) {
		return 0.0, nil
	}
	return r.Value, nil
}

// A ResultList is a list of Results.
type ResultList []*Result

// Filter returns a slice of pointers to the Results in a ResultList that match
// a predicate.
func (rl ResultList) Filter(f func(*Result) bool) []*Result {
	var results = []*Result{}
	for _, r := range rl {
		if f(r) {
			results = append(results, r)
		}
	}
	return results
}
