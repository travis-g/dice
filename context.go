package dice

import (
	"context"
)

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "dice context value " + k.name
}

var (
	CtxKeyTotalRolls = &contextKey{name: "total rolls"}
	CtxKeyMaxRolls   = &contextKey{name: "max rolls"}
	CtxKeyParameters = &contextKey{name: "parameters"}
)

// NewContextFromContext makes a child context from a given
// context.
func NewContextFromContext(ctx context.Context) context.Context {
	// add a roll counter, if one doesn't exist
	if _, ok := ctx.Value(CtxKeyTotalRolls).(*uint64); !ok {
		return context.WithValue(ctx, CtxKeyTotalRolls, new(uint64))
	}
	return ctx
}

func TotalRolls(ctx context.Context) *uint64 {
	if count, ok := ctx.Value(CtxKeyTotalRolls).(*uint64); ok {
		return count
	}
	return new(uint64)
}

func Parameters(ctx context.Context) map[string]interface{} {
	if params, ok := ctx.Value(CtxKeyParameters).(map[string]interface{}); ok {
		return params
	}
	return make(map[string]interface{})
}
