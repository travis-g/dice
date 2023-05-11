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

// NewContextFromContext makes a child context from a given context, including
// setting the context's maximum rolls and adding a roll counter.
func NewContextFromContext(ctx context.Context) context.Context {
	// ensure a maximum roll value is present
	if _, ok := ctx.Value(CtxKeyMaxRolls).(uint64); !ok {
		ctx = context.WithValue(ctx, CtxKeyMaxRolls, MaxRolls)
	}
	// add a roll counter, if one doesn't exist
	if _, ok := ctx.Value(CtxKeyTotalRolls).(*uint64); !ok {
		return context.WithValue(ctx, CtxKeyTotalRolls, new(uint64))
	}
	return ctx
}

// CtxTotalRolls returns the pointer to total number of rolls made by the
// context.
func CtxTotalRolls(ctx context.Context) *uint64 {
	if count, ok := ctx.Value(CtxKeyTotalRolls).(*uint64); ok {
		return count
	}
	return new(uint64)
}

// CtxMaxRolls returns the context's maximum allowed number of rolls, or the
// default.
func CtxMaxRolls(ctx context.Context) uint64 {
	if max, ok := ctx.Value(CtxKeyMaxRolls).(uint64); ok {
		return max
	}
	return MaxRolls
}

func CtxParameters(ctx context.Context) map[string]interface{} {
	if params, ok := ctx.Value(CtxKeyParameters).(map[string]interface{}); ok {
		return params
	}
	return make(map[string]interface{})
}
