package math

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/travis-g/dice"
)

// package-level variable to prevent optimizations
var (
	i   interface{}
	ctx = context.Background()
)

// Check implements
var (
	_ = fmt.Stringer(&ExpressionResult{})
	_ = fmt.GoStringer(&ExpressionResult{})
)

// function to set math/rand as the RNG source
func sourceMathRand() {
	seed, _ := dice.CryptoInt64()
	dice.Source = rand.New(rand.NewSource(seed))
}

func init() {
	sourceMathRand()
}

func BenchmarkEvaluate(b *testing.B) {
	b.ReportAllocs()
	benchmarks := []struct {
		expression string
	}{
		{""},
		{"d6"},
		{"d20"},
		{"1d20"},
		{"3d20"},
		{"1d20+1d20+1d20"},
		{"3d20+1"},
		{"3d20+2d4"},
		{"100d6"},
	}
	var de *ExpressionResult
	for _, bmark := range benchmarks {
		b.Run(bmark.expression, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				de, _ = EvaluateExpression(ctx, bmark.expression)
			}
		})
	}
	i = de
}

func BenchmarkEvaluateCount(b *testing.B) {
	b.ReportAllocs()
	benchmarks := []struct {
		expression string
	}{
		{"1d20"},
		{"2d20"},
		{"3d20"},
		{"4d20"},
		{"5d20"},
		{"10d20"},
		{"15d20"},
		{"20d20"},
		{"25d20"},
		{"50d20"},
		{"100d20"},
	}
	var de *ExpressionResult
	for _, bmark := range benchmarks {
		b.Run(bmark.expression, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				de, _ = EvaluateExpression(ctx, bmark.expression)
			}
		})
	}
	i = de
}

func BenchmarkEvaluateSize(b *testing.B) {
	b.ReportAllocs()
	benchmarks := []struct {
		expression string
	}{
		{"1d1"},
		{"1d2"},
		{"1d3"},
		{"1d4"},
		{"1d5"},
		{"1d10"},
		{"1d15"},
		{"1d20"},
		{"1d25"},
		{"1d50"},
		{"1d100"},
	}
	var de *ExpressionResult
	for _, bmark := range benchmarks {
		b.Run(bmark.expression, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				de, _ = EvaluateExpression(ctx, bmark.expression)
			}
		})
	}
	i = de
}

func BenchmarkEvaluateDiceFunctions(b *testing.B) {
	b.ReportAllocs()
	benchmarks := []struct {
		name       string
		expression string
	}{
		{"min", "min(0,1)"},
		{"max", "max(0,1)"},
		{"floor", "floor(0.5)"},
		{"ceil", "ceil(0.5)"},
		{"round", "round(0.5)"},
	}
	var de *ExpressionResult
	for _, bmark := range benchmarks {
		b.Run(bmark.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				de, _ = EvaluateExpression(ctx, bmark.expression)
			}
		})
	}
	i = de
}

func TestEvaluate(t *testing.T) {
	testCases := []struct {
		expression string
		result     float64
	}{
		{"1", 1},
		{"d1", 1},
		{"d0", 0},
	}
	var de *ExpressionResult
	for _, tc := range testCases {
		de, err := EvaluateExpression(ctx, tc.expression)
		t.Logf("evaluating %s; got %v", tc.expression, de)
		if err != nil {
			t.Fatalf("error evaluating \"%s\": %s", tc.expression, err)
		}
		if de.Result != tc.result {
			t.Errorf("evaluated %s; got result %v, wanted %v", tc.expression, de.Result, tc.result)
		}
	}
	i = de
}
