package main

import (
	"testing"
)

// global variable to prevent compiler optimizations
var global any

type ExpressionTestCase struct {
	expression string
	// if expression's mathematic result is deterministic result is a pointer to
	// the evaluated expression's result, ex. "1+1" -> `ptr(2.0)`.
	result  *float64
	wantErr bool
}

// Deterministic returns whether a test case has a defined result, indicating
// any evaluation of the expression should result in the same value.
func (e *ExpressionTestCase) Deterministic() bool {
	return e.result != nil
}

// ptr returns a pointer to the passed value.
func ptr[T any](v T) *T {
	return &v
}

var benchmarkCases = []ExpressionTestCase{
	{"", nil, true},
	{"1d20", nil, false},
	{"d20", nil, false},
	{"dF", nil, false},
	{"1d20 + 1", nil, false},
	{"1d20+1", nil, false},
	{"1+1d20", nil, false},
	{"1+1", ptr(2.0), false},
	{"3+5*2", ptr(13.0), false},
	{"3+5*2", ptr(13.0), false},
	{"(3+5)*2", ptr(16.0), false},
}

var moreCases = []ExpressionTestCase{
	{"D20", nil, false},
	{"d0", ptr(0.0), false},
	{"d1", ptr(1.0), false},
	{" d20 ", nil, false},
	{"\td20", nil, false},
	{"1 d20", nil, false},
	{"0d20", ptr(0.0), false},
	{"(0d20)", ptr(0.0), false},
	{"1d20d1", ptr(0.0), false},
	{"1d20[foo]", nil, false},
	{"1d20d1[foo]", ptr(0.0), false},
	{"8/2*4", ptr(16.0), false},
	{"8/(2*4)", ptr(1.0), false},
	{"4+6/2", ptr(7.0), false},
	{"(4+6)/2", ptr(5.0), false},
	{"2*3**2", ptr(18.0), false},
	{"(2*3)**2", ptr(36.0), false},
	{"10-(3+2)", ptr(5.0), false},
	{"10 - 3 + 2", ptr(5.0), false},
	{"10- 3+2", ptr(9.0), false},
	{"10-3+2", ptr(9.0), false},
	{"1.1", ptr(1.1), false},
	{".1", ptr(0.1), false},
	{"0+.1", ptr(0.1), false},
	{"0-.1", ptr(-0.9), false},
	{"?{foo}", nil, false},
	{"1+-1", ptr(0.0), false},
	{"1+(-1)", ptr(0.0), false},
	{"1--1", ptr(2.0), false},
	{"1- -1", ptr(2.0), false},

	{"\n", nil, true},
	{"# comment", nil, true},
	{"// comment", nil, true},
	{`\ comment`, nil, true},
	{` \ comment`, nil, true},

	// TODO: fix these cases
	{"\nd20", nil, true},
	{"d20\n", nil, true},
	{"1 // comment", ptr(1.0), true},
	{"// comment\n1", ptr(1.0), true},
	{"-.3", ptr(-0.3), true},
	{"+.3", ptr(0.3), true},
	{"1[foo]", ptr(1.0), true},
}

var expressionParseCases = append(benchmarkCases, moreCases...)

func TestParse(t *testing.T) {
	for _, tt := range expressionParseCases {
		t.Run(tt.expression, func(t *testing.T) {
			got, err := ParseString(tt.expression, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			global = got
		})
	}
}

func BenchmarkParse(b *testing.B) {
	for _, tt := range benchmarkCases {
		b.Run(tt.expression, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				got, err := ParseString(tt.expression, false)
				if (err != nil) != tt.wantErr {
					b.Errorf("error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				global = got
			}
		})
	}
}
