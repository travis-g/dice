package dice

import (
	"context"
)

type contextKey string

func (k contextKey) String() string {
	return "dice context value " + string(k)
}

// CtxKeyTotalRolls is the context key for the total number of rolls made during
// the request. NOTE: It is a pointer to a uint64, allowing parent contexts to
// respond properly to additional rolls made by child contexts.
var CtxKeyTotalRolls = contextKey("total rolls")

// CtxKeyMaxRolls is the context key for the maximum number of rolls allowed for
// the context.
var CtxKeyMaxRolls = contextKey("max rolls")

var (
	CtxKeyParameters = contextKey("parameters")
	CtxKeyFunctions  = contextKey("functions")
)

// NewContextFromContext makes a child context from a given context, including
// setting the context's maximum rolls and adding a roll counter if not present.
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

// CtxTotalRolls returns the pointer to count of rolls made within the context.
// It returns a pointer to a new uint64 and a non-nil error if the value is
// missing.
func CtxTotalRolls(ctx context.Context) (*uint64, error) {
	if count, ok := ctx.Value(CtxKeyTotalRolls).(*uint64); ok {
		return count, nil
	}
	return new(uint64), ErrContextKeyMissing
}

// CtxMaxRolls returns the context's maximum allowed number of rolls. If
// missing, it returns the default and a non-nil error. The default is defined
// by [MaxRolls].
func CtxMaxRolls(ctx context.Context) (uint64, error) {
	if max, ok := ctx.Value(CtxKeyMaxRolls).(uint64); ok {
		return max, nil
	}
	return MaxRolls, ErrContextKeyMissing
}

// CtxParameters returns the context's arbitrary parameters. If missing, it
// returns an empty map and a non-nil error.
func CtxParameters(ctx context.Context) (map[string]string, error) {
	if params, ok := ctx.Value(CtxKeyParameters).(map[string]string); ok {
		return params, nil
	}
	return make(map[string]string), ErrContextKeyMissing
}
