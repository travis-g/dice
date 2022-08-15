package dice

import (
	"context"

	"go.uber.org/atomic"
)

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "dice context value " + k.name
}

var (
	CtxKeyRoll       = &contextKey{name: "roll"}
	CtxKeyTotalRolls = &contextKey{name: "total rolls"}
	CtxKeyMaxRolls   = &contextKey{name: "max rolls"}
)

// NewContextFromContext makes a child context with context keys from a given
// context.
func NewContextFromContext(ctx context.Context) context.Context {
	// ensure context has roll counter
	if _, ok := ctx.Value(CtxKeyTotalRolls).(*atomic.Uint64); !ok {
		ctx = context.WithValue(ctx, CtxKeyTotalRolls, atomic.NewUint64(0))
	}
	return ctx
}
