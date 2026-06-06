package main

import (
	"testing"
)

func TestEvaluateRoot(t *testing.T) {
	testCases := []struct {
		expression string
		want       float64
		wantErr    bool
	}{
		{"1+1", 2, false},
		{"3+5*2", 13, false},
		{"(3+5)*2", 16, false},
		{"2*3**2", 18, false},
		{"(2*3)**2", 36, false},
		{"10-(3+2)", 5, false},
		{"1/2", 0.5, false},
		{"min(1,2)", 1, false},
		{"min(1)", 1, false},
		{"max(1,2)", 2, false},
		{"round(1.5)", 2, false},
		{"abs(-1)", 1, false},
		{"1d1", 1, false},
		{"2d0", 0, false},
		{"?{foo}", 1, false},
		{"?{bar}", 3, false},
		{"?{foo}+?{bar}", 4, false},
		{"?{unset}", 0, true},    // won't resolve
		{"?{unset|1}", 1, false}, // default option
		{"min(?{foo}, 2)", 1, false},
		{"min(?{unset|1}, 2)", 1, false},
		{"max(?{foo}, 2)", 2, false},
		{"max(?{roll}, 41)", 41, false},
		{"?{macro}", 2, false},
		{"?{2d20}", 1, false},
	}

	for _, tc := range testCases {
		ctx := newContextWithParams(testParams)
		t.Run(tc.expression, func(t *testing.T) {
			root, err := ParseString(tc.expression, false)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			got, err := root.Evaluate(ctx)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Evaluate() error = %v, wantErr %v", err, tc.wantErr)
			}
			if err == nil && got != tc.want {
				t.Fatalf("Evaluate() = %v, want %v", got, tc.want)
			}
		})
	}
}
