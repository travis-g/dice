package dice

import "context"

var (
	// MaxRequestRolls is the max number of Rolls per request context allowed
	MaxRequestRolls uint32 = 128
)

// contextKey is a value for use with context.WithValue.
type contextKey string

var (
	contextKeyTotalRolls = contextKey("request rolls")
)

func (k contextKey) String() string {
	return "github.com/travis-g/dice context value " + string(k)
}

func ContextTotalRollCount(ctx context.Context) (count uint32, ok bool) {
	count, ok = ctx.Value(contextKeyTotalRolls).(uint32)
	return
}
