package dice

import (
	"context"

	"github.com/pkg/errors"
)

// NewResult returns a new un-dropped, non-critical Result's pointer.
func NewResult(result float64) *Result {
	return &Result{
		Value: result,
	}
}

// Drop marks a Result as dropped.
func (r *Result) Drop(ctx context.Context, dropped bool) {
	if r != nil {
		r.Dropped = dropped
	}
}

// Total returns the Result's value or 0 if the result was dropped.
func (r *Result) Total(ctx context.Context) (float64, error) {
	if r == nil {
		return 0.0, errors.New("nil result")
	}
	if r.Dropped {
		return 0.0, nil
	}
	return r.Value, nil
}
