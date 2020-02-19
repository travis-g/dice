package math

import (
	"testing"
)

func TestDiceFunctions(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
		result     float64
	}{
		{"abs-neg", "abs(-1)", 1},
		{"abs-pos", "abs(1)", 1},
		{"abs-zero", "abs(0)", 0},
		{"ceil0.5", "ceil(0.5)", 1},
		{"ceil0", "ceil(0.0)", 0},
		{"floor0.5", "floor(0.5)", 0},
		{"floor0.6", "floor(0.6)", 0},
		{"max01", "max(0,1)", 1},
		{"min01", "min(0,1)", 0},
		{"round-down", "round(0.49)", 0},
		{"round-up", "round(0.5)", 1},
	}
	var de *ExpressionResult
	for _, tc := range testCases {
		de, err := EvaluateExpression(ctx, tc.expression)
		if err != nil {
			t.Fatalf("error evaluating %s: %s", tc.expression, err)
		}
		if de.Result != tc.result {
			t.Errorf("evaluated %s; got result %v, wanted %v", tc.expression, de.Result, tc.result)
		}
	}
	i = de
}
