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

// params are parameters that should be referenced when using test roll queries.
// These should be set as values in test contexts.
var params map[string]string = map[string]string{
	"foo":     "1",
	"baz":     "",
	"quz":     "2+1",
	"foo bar": "4",
	// bar: nil
}

var benchmarkCases = []ExpressionTestCase{
	{"", nil, true},
	{"1d20", nil, false},
	{"d20", nil, false},
	{"1dF", nil, false},
	{"dF", nil, false},
	{"1d20 + 1", nil, false},
	{"1d20+1", nil, false},
	{"1+1d20", nil, false},
	{"1+1", ptr(2.0), false},
	{"3+5*2", ptr(13.0), false},
	{"(3+5)*2", ptr(16.0), false},
	{"1d20ro1", nil, false},
	{"3d6", nil, false},
	{"3d6s", nil, false},
	{"?{foo}", ptr(1.0), false},
	{"?{bar|1}", ptr(1.0), false},
	{"foo", nil, true},
}

var moreCases = []ExpressionTestCase{
	{"D20", nil, false},
	{"d0", ptr(0.0), false},
	{"d1", ptr(1.0), false},
	{" d20 ", nil, false},
	{"\td20", nil, false},
	{"1 d20", nil, false},
	{"1d 20", nil, true},
	{"0d20", ptr(0.0), false},
	{"(0d20)", ptr(0.0), false},
	{"1d20d1", ptr(0.0), false},
	{"1d20d", ptr(0.0), false},
	{"1dFd", ptr(0.0), false},
	{"1d20sd", ptr(0.0), false},
	{"1d20[foo]", nil, false},
	{"1d20d1[foo]", ptr(0.0), false},
	{"1d20!", nil, false},
	{"1d20r", nil, false},
	{"1d20sa", nil, false},
	{"1d20sd", nil, false},
	{"1d20ssd", nil, false},
	{"1d20sad", nil, false}, // sort + drop
	{"1d20cf", nil, false},
	{"1d20cs", nil, false},
	{"1d20cs>2", nil, false},
	{"1d20!p", nil, false},
	{"1d20!p>8", nil, false},
	{"1d20!!", nil, false},
	{"1d20r1r2", nil, false},
	{"1d20r1r2>4", nil, false},
	{"1d20r1r2>0", ptr(1.0), false},
	{"1dF>4", ptr(0.0), false},
	{"2dF>-1", ptr(2.0), false},
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
	{"1/2", ptr(0.5), false},
	{"1.5*2", ptr(3.0), false},
	{"3%2", ptr(1.0), false},
	{"3^3", ptr(9.0), false},
	{"1.1", ptr(1.1), false},
	{".1", ptr(0.1), false},
	{"0+.1", ptr(0.1), false},
	{"0-.1", ptr(-0.1), false},
	{"?{foo|a,1}", nil, false},
	{"1+?{foo|a,1}", nil, false},
	{"?{foo|a,1|b, 2}", nil, false},
	{"?{baz|1|2}", nil, false},
	{"?{qux}", ptr(3.0), false},
	{"?{foo bar}", ptr(4.0), false},
	{"1+-1", ptr(0.0), false},
	{"1+(-1)", ptr(0.0), false},
	{"1--1", ptr(2.0), false},
	{"1-+1", ptr(0.0), false},
	{"1*+2", ptr(2.0), false},
	{"1- -1", ptr(2.0), false},
	{"-.3", ptr(-0.3), false},
	{"+.3", ptr(0.3), false},
	{"min(1,2)", ptr(1.0), false},
	{"min(1)", nil, false},
	{"round(1.2)", ptr(1.0), false},
	{"round(1,2)", nil, false},
	{"round()", nil, false},
	{"1 // comment", ptr(1.0), false},
	{"1// comment", ptr(1.0), false},
	{"1//comment", ptr(1.0), false},
	// Fail comments-only
	{"# comment", nil, true},
	{"// comment", nil, true},
	{`\ comment`, nil, true},
	{` \ comment`, nil, true},

	// TODO: fix these cases
	{"1d20d>2", nil, false},
	{"\nd20", nil, true},
	{"d20\n", nil, true},
	{"// comment\n1", ptr(1.0), true},
	{"1[foo]", ptr(1.0), true},
	{"(3)d(1)", ptr(3.0), true},
	{"(3)d(1)k2", ptr(2.0), true},
	{"{3,4}k1", ptr(4.0), true},
	{"{3,4}>=2", ptr(2.0), true},
	{"{1,1}d1>=2", ptr(0.0), true},
	{"{1,3}d1>=2", ptr(1.0), true},
	{"[[[[2]]d1]]+1", ptr(3.0), true},
	// {"[[[[2]]d1]]", nil, true}, // TBD

	// should fail always
	{"1+--1", nil, true},
	{"1+*1", nil, true},
	{"1=1", nil, true},
	{"1+=1", nil, true},
	{"\n", nil, true},
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
