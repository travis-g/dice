package dice

import "context"

// A Result is a value a die has rolled.
type Result struct {
	Value    float64 `json:"value"`
	Dropped  bool    `json:"dropped,omitempty"`
	Critical bool    `json:"critical,omitempty"`
}

// NewResult returns a new un-dropped Result.
func NewResult(result float64) *Result {
	return &Result{
		Value: result,
	}
}

// Drop marks a Result as dropped.
func (r *Result) Drop(ctx context.Context, drop bool) {
	r.Dropped = drop
}
