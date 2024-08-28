package main

import (
	"encoding/json"
	"testing"
)

// write-only global variable to prevent compiler optimizations
var global any

type ExpressionTestCase struct {
	expression string
	// if expression's mathematic result is deterministic then result is a
	// pointer to the evaluated expression's result, ex. "1+1" => `ptr(2.0)`.
	result *float64
	// test cases include whether the parsing of the string should fail
	wantParseErr bool
}

// Deterministic returns whether a test case has a defined result, indicating
// any evaluation of the expression should result in the same value.
func (e *ExpressionTestCase) Deterministic() bool {
	return e.result != nil
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
	{"?{undefined|1}", ptr(1.0), false},
	{"foo", nil, true},
}

// More test cases. Many of these should parse correctly, but may throw errors
// when evaluated.
var moreCases = []ExpressionTestCase{
	// TODO: break these into more case "sets" for maintainability, ex. case
	// sensitivity set vs. future features set vs. eval error set
	{"D20", nil, false},
	{"df", nil, false},
	{"Df", nil, false},
	{"d0", ptr(0.0), false},
	{"d1", ptr(1.0), false},
	{" d20 ", nil, false},
	{"\td20", nil, false},
	{"1 d20", nil, false},
	{"1d 20", nil, true},
	{"0d20", ptr(0.0), false},
	{"(0d20)", ptr(0.0), false},
	{"1d20k", nil, false},
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
	{"1d20R", nil, false}, // case insensitivity tests
	{"1d20Sa", nil, false},
	{"1d20Sd", nil, false},
	{"1d20sD", nil, false},
	{"1d20ssd", nil, false},      // double sort
	{"1d20sad", ptr(0.0), false}, // sort + drop
	{"1d20cf", nil, false},
	{"1d20cs", nil, false},
	{"1d20cs>2", nil, false},
	{"1d20!p", nil, false},
	{"1d20!p>8", nil, false},
	{"1d20!!", nil, false},
	{"1d20!!>8", nil, false},
	{"1d20r1r2", nil, false},
	{"1d20>1", ptr(1.0), false}, // success/failure modifiers
	{"1d20r1r2>4", nil, false},
	{"1d20r1r2>0", ptr(1.0), false},
	{"1d20f>1", ptr(0.0), false},
	{"3d6f>2<3", nil, false}, // "reversed" success modifiers
	{"1d20>10f<5", nil, false},
	{"1d20>1f<20", ptr(0.0), false}, // "Schrödinger's roll"
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
	{"5%2", ptr(1.0), false},
	{"3^3", ptr(9.0), false},
	{"1.1", ptr(1.1), false},
	{".1", ptr(0.1), false},
	{"0+.1", ptr(0.1), false},
	{"0-.1", ptr(-0.1), false},
	{"?{foo|a,1}", nil, false}, // roll queries
	{"?{foo|a,1|b, 2}", nil, false},
	{"?{baz|1}", nil, false},
	{"?{baz|1|2}", nil, false},
	{"?{bar}", ptr(3.0), false},
	{"?{foo bar}", ptr(4.0), false},
	{"1+?{foo|a,1}", nil, false},
	{"1+-1", ptr(0.0), false},
	{"1+(-1)", ptr(0.0), false},
	{"1-+1", ptr(0.0), false},
	{"1--1", ptr(2.0), false},
	{"1- -1", ptr(2.0), false},
	{"1*+3", ptr(3.0), false},
	{"-.3", ptr(-0.3), false},
	{"+.3", ptr(0.3), false},
	{"min(1,2)", ptr(1.0), false},
	{"min(1, 2)", ptr(1.0), false},
	{"min(1, 2 )", ptr(1.0), false},
	{"min(1)", ptr(1.0), false},
	{"min()", nil, false}, // eval error?
	{"min(1, 2, 3)", ptr(1.0), false},
	{"max(1, 2, 3)", ptr(3.0), false},
	{"abs(-1)", ptr(1.0), false},
	{"abs(-1, 2)", nil, false}, // eval error
	{"round(1.2)", ptr(1.0), false},
	{"round(1,2)", nil, false}, // eval error
	{"round()", nil, false},    // eval error?
	{"1 // comment", ptr(1.0), false},
	{"1// comment", ptr(1.0), false},
	{"1//comment", ptr(1.0), false},
	{"1//comment ", ptr(1.0), false},
	{"?{undefined}", nil, false}, // eval error
	{"((((((1))))))", ptr(1.0), false},

	// comment should still parse but throw eval errors (nil results)
	{"# comment", nil, false},
	{"// comment", nil, false},
	{`\ comment`, nil, false},
	{` \ comment`, nil, false},

	// should fail always
	{"+", nil, true},
	{"-", nil, true},
	{"1+--1", nil, true},
	{"1+*1", nil, true},
	{"1=1", nil, true},
	{"1+=1", nil, true},
	{"1+(1", nil, true},
	{"1+)1", nil, true},
	{"1(2)", nil, true},
	{"\nd20", nil, true}, // unsanitized roll strings
	{"\n", nil, true},
	{"d20\n", nil, true},

	// TODO: fix the below cases cases?
	{"3d6d>2", nil, false},            // no comparisons on drop/keep modifiers
	{"// comment\n1", ptr(1.0), true}, // support single multiline roll
	{"1[foo]", ptr(1.0), true},        // allow labels on Numbers/Factors
	{"1d20f1", nil, true},             // allow optional equals for failures?

	// TODO: future features
	// computed dice
	{"d(1)", ptr(1.0), true},
	{"3d(1)", ptr(1.0), true},
	{"(3)d1", ptr(3.0), true},
	{"(3)d(1)", ptr(3.0), true},
	{"(3d1)d1", ptr(3.0), true},
	{"(3)d(1)k2", ptr(2.0), true},
	{"?{bar}d(1)", ptr(3.0), true},
	// dice groups
	{"{3,4}k1", ptr(4.0), true},
	{"{3,4}>=2", ptr(2.0), true},
	{"{3,4}s", ptr(7.0), true},
	{"{1,3}d1>=2", ptr(1.0), true},
	{"{1,1}d1>=2", ptr(0.0), true},
	{"{3+4}k1", nil, true}, // eval error?
	// inline rolls
	{"[[[[2]]d1]]+1", nil, true},
	// {"[[[[2]]d1]]", nil, true}, // TBD
}

var expressionParseCases = append(benchmarkCases, moreCases...)

type ASTTestCase struct {
	expression string
	ast        *Root
	wantErr    bool
}

// params are parameters referenced when using test roll queries. These should
// be passed into test contexts.
var params map[string]string = map[string]string{
	"foo":     "1",
	"bar":     "2+1",
	"foo bar": "4",
	"baz":     "",
}

var astCases = []ASTTestCase{
	// TODO: ensure that defaults are tested as well
	{"", nil, true},
	{"1", &Root{Expr: &Expr{L: &Factor{Number: ptr(1.0)}}}, false},
	{"1d20", &Root{Expr: &Expr{L: &Factor{Dice: &Dice{Count: ptr(1), Size: ptr(20)}}}}, false},
	{"d20", &Root{Expr: &Expr{L: &Factor{Dice: &Dice{Count: ptr(1), Size: ptr(20)}}}}, false},
	{"1d20+1", &Root{Expr: &Expr{L: &Factor{Dice: &Dice{Count: ptr(1), Size: ptr(20)}}, R: []*OpFactor{{Op: "+", Factor: &Factor{Number: ptr(1.0)}}}}}, false},
	{"1 //test", &Root{Expr: &Expr{L: &Factor{Number: ptr(1.0)}}, Comment: ptr("test")}, false},
	{"1 // te st ", &Root{Expr: &Expr{L: &Factor{Number: ptr(1.0)}}, Comment: ptr("te st ")}, false},
	{"0+.1", &Root{Expr: &Expr{L: &Factor{Number: ptr(0.0)}, R: []*OpFactor{{Op: "+", Factor: &Factor{Number: ptr(0.1)}}}}}, false},
	{"1--1", &Root{Expr: &Expr{L: &Factor{Number: ptr(1.0)}, R: []*OpFactor{{Op: "-", Factor: &Factor{Number: ptr(-1.0)}}}}}, false},
	{"1d20[foo]", &Root{Expr: &Expr{L: &Factor{Dice: &Dice{Count: ptr(1), Size: ptr(20)}, Label: "foo"}}}, false},
}

func TestParseString(t *testing.T) {
	t.Parallel()

	// test that given expressions can be parsed or error out as expected
	for _, tt := range expressionParseCases {
		t.Run(tt.expression, func(t *testing.T) {
			got, err := ParseString(tt.expression, false)
			// TODO: confirm that err is desired type
			if (err != nil) != tt.wantParseErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantParseErr)
				return
			}
			global = got
		})
	}

	// test that expected ASTs are created from specific expressions
	for _, tt := range astCases {
		t.Run(tt.expression, func(t *testing.T) {
			got, err := ParseString(tt.expression, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !deepEqual(got, tt.ast) {
				gotBytes, _ := json.Marshal(got)
				astBytes, _ := json.Marshal(tt.ast)
				t.Errorf("got = %v, want %v", string(gotBytes), string(astBytes))
			}
			global = got
		})
	}
}

func BenchmarkParseString(b *testing.B) {
	for _, tt := range benchmarkCases {
		b.Run(tt.expression, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				got, err := ParseString(tt.expression, false)
				if (err != nil) != tt.wantParseErr {
					b.Errorf("error = %v, wantErr %v", err, tt.wantParseErr)
					return
				}
				global = got
			}
		})
	}
}
