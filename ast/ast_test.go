package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

// write-only global variable to prevent compiler optimizations
var global any

// Ensure that all AST types implement the Stringer interface.
var (
	_ Stringer = (*Args)(nil)
	_ Stringer = (*Dice)(nil)
	_ Stringer = (*Expr)(nil)
	_ Stringer = (*Func)(nil)
	_ Stringer = (*GroupModifier)(nil)
	_ Stringer = (*Modifier)(nil)
	_ Stringer = (*OpTerm)(nil)
	_ Stringer = (*Query)(nil)
	_ Stringer = (*QueryOption)(nil)
	_ Stringer = (*Root)(nil)
	_ Stringer = (*Term)(nil)
)

func TestRoot_String(t *testing.T) {
	tests := []struct {
		r       *Root
		want    string
		wantErr bool
	}{
		{&Root{}, "", false},
		{&Root{Comment: ptr(" test ")}, "#  test ", false},
		{&Root{Comment: ptr("test")}, "# test", false},
		{&Root{Expr: &Expr{L: &Term{Number: ptr(1.0)}}, Comment: ptr("test")}, "1 # test", false},
		{&Root{Expr: &Expr{L: &Term{Number: ptr(1.0)}}}, "1", false},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.r.String()
			if (got != tt.want) != tt.wantErr {
				t.Errorf("Root.String() = %v, want %v", got, tt.want)
			}
			global = got
		})
	}
}

func TestTerm_String(t *testing.T) {
	tests := []struct {
		f       *Term
		want    string
		wantErr bool
	}{
		// TODO: more test cases!
		{&Term{Dice: &Dice{Count: ptr(1), Size: ptr(6)}}, "d6", false},
		{&Term{Dice: &Dice{Count: ptr(2), Fudge: true}}, "2dF", false},
		{&Term{Dice: &Dice{Count: ptr(2), Size: ptr(6)}}, "2d6", false},
		{&Term{Dice: &Dice{Count: ptr(2), Size: ptr(8)}, Label: ptr("test")}, "2d8[test]", false},
		{&Term{Func: &Func{Name: "foo", Args: &Args{Arg: []*Term{{Number: ptr(1.0)}}}}}, "foo(1)", false},
		{&Term{Func: &Func{Name: "foo"}}, "foo()", false},
		{&Term{GroupModifier: []*GroupModifier{{Failure: true, CompareOp: ptr("<"), Value: ptr(19)}}}, "f<19", false},
		{&Term{Label: ptr("test")}, "[test]", false},
		{&Term{Modifiers: []*Modifier{{Type: "r", CompareOp: ptr(">")}}}, "r>", false},
		{&Term{Modifiers: []*Modifier{{Type: "foo"}, {Type: "bar", Value: ptr(1)}}}, "foobar1", false},
		{&Term{Number: ptr(-0.000001)}, "-0.000001", false},
		{&Term{Number: ptr(0.0)}, "0", false},
		{&Term{Number: ptr(0.000001)}, "0.000001", false},
		{&Term{Number: ptr(10.0), Label: ptr("test")}, "10[test]", false},
		{&Term{Number: ptr(10.0)}, "10", false},
		{&Term{Query: &Query{Name: "foo", Options: []*QueryOption{{OptionLabel: "test", OptionValue: "3"}}}}, "?{foo|test, 3}", false},
		{&Term{Query: &Query{Name: "foo", Options: []*QueryOption{{OptionValue: "3"}}}}, "?{foo|3}", false},
		{&Term{Query: &Query{Name: "foo"}}, "?{foo}", false},
		{&Term{Subexpr: &Expr{}}, "()", false},
		{&Term{Subexpr: &Expr{L: &Term{Number: ptr(0.0)}, R: []*OpTerm{{Op: "+", Term: &Term{Number: ptr(1.0)}}}}}, "(0 + 1)", false},
		{&Term{Subexpr: &Expr{L: &Term{Number: ptr(0.0)}}}, "(0)", false},
		{&Term{Subexpr: &Expr{L: &Term{Subexpr: &Expr{L: &Term{Number: ptr(0.0)}}}}}, "((0))", false},

		// TODO(travis-g): decide the results of the below cases
		{&Term{Dice: &Dice{Count: ptr(0), Size: ptr(6)}}, "0d6", false},
		{&Term{Query: &Query{Name: "foo", Options: []*QueryOption{{OptionLabel: "test"}}}}, "?{foo}", true},
		{&Term{Query: &Query{}}, "?{}", false}, // panic, or print nothing?
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.f.String()
			if (got != tt.want) != tt.wantErr {
				t.Errorf("Term.String() = %v, want %v", got, tt.want)
			}
			global = got
		})
	}
}

func TestQuantity_String(t *testing.T) {
	tests := []struct {
		q    *Quantity
		want string
	}{
		{&Quantity{Number: ptr(1)}, "1"},
		{&Quantity{Subexpr: &Expr{L: &Term{Number: ptr(1.0)}, R: []*OpTerm{{Op: "+", Term: &Term{Number: ptr(2.0)}}}}}, "(1 + 2)"},
		{&Quantity{Query: &Query{Name: "foo", Options: []*QueryOption{{OptionLabel: "bar", OptionValue: "3"}}}}, "?{foo|bar, 3}"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.q.String(); got != tt.want {
				t.Errorf("Quantity.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

type expressionTestCase struct {
	// expression for the test case.
	expression string
	// if expression's mathematic result is deterministic then result is a
	// pointer to the evaluated expression's result, ex. "1+1" => `ptr(2.0)`.
	result *float64
	// whether the expression should parse correctly.
	wantParseErr bool
}

// Deterministic returns whether a test case has a defined result, indicating
// any evaluation of the expression should result in the same value.
func (e *expressionTestCase) Deterministic() bool {
	return e.result != nil
}

var benchmarkCases = []expressionTestCase{
	{"", nil, true},
	{"1d20", nil, false},
	{"d20", nil, false},
	{"D20", nil, false},
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

// More test cases. Many of these should parse correctly as they are
// syntactically valid but would throw errors when evaluated.
var moreCases = []expressionTestCase{
	// TODO: break these into more case "sets" for maintainability, ex. case
	// sensitivity set vs. future features set vs. eval error set
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
	{"1d20sad", ptr(0.0), false}, // sort ascending + drop
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
	{"?{foo|1}", nil, false}, // roll queries
	{"?{foo|a,1}", nil, false},
	{"?{bar}", ptr(3.0), false},
	{"?{foo bar}", ptr(4.0), false},
	{"1+?{foo|a,1}", nil, false},
	{"?{foo|a,1|b, 2}", nil, false}, // dropdowns
	{"?{baz|1|2}", nil, false},
	{"1+1", ptr(2.0), false},
	{"1+-1", ptr(0.0), false},
	{"1+(-1)", ptr(0.0), false},
	{"1-+1", ptr(0.0), false},
	{"1--1", ptr(2.0), false},
	{"1- -1", ptr(2.0), false},
	{"1*+3", ptr(3.0), false},
	{"-.3", ptr(-0.3), false},
	{"+.3", ptr(0.3), false},
	{"min(1,2)", ptr(1.0), false}, // functions
	{"min(1, 2)", ptr(1.0), false},
	{"min(1, 2 )", ptr(1.0), false},
	{"min(1)", ptr(1.0), false}, // eval error?
	{"min()", nil, false},       // eval error?
	{"min(1, 2, 3)", ptr(1.0), false},
	{"max(1, 2, 3)", ptr(3.0), false},
	{"abs(-1)", ptr(1.0), false},
	{"abs(-1, 2)", nil, false}, // eval error
	{"round(1.2)", ptr(1.0), false},
	{"round(1.5)", ptr(2.0), false},
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
	{"1[]", ptr(1.0), true}, // empty label

	// TODO: fix the below test cases
	{"3d6d>2", nil, false},            // no comparisons on drop/keep modifiers
	{"5d10s9#2s7f2#2f", nil, false},   // Fantasy Grounds
	{"// comment\n1", ptr(1.0), true}, // support single multiline roll
	{"1[foo]", ptr(1.0), true},        // allow labels on Numbers/Factors
	{"1 [foo]", ptr(1.0), true},       // allow spaces before labels
	{"1d20f1", nil, true},             // allow optional equals sign for failures?
	{"sling5", nil, true},             // custom regexp-based macro

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
	{"{3,4}", ptr(7.0), true},
	{"{3, 4}", ptr(7.0), true},
	{"{3,4}k1", ptr(4.0), true},
	{"{3,4}>=2", ptr(2.0), true},
	{"{3,4}s", ptr(7.0), true},
	{"{4,3}s", ptr(7.0), true},
	{"{1,3}d1>=2", ptr(1.0), true},
	{"{1,1}d1>=2", ptr(0.0), true},
	{"{3+4}k1", nil, true}, // eval error?
	// inline rolls
	{"[[[[2]]d1]]+1", ptr(1.0), true},
	// {"[[[[2]]d1]]", nil, true}, // TBD
}

var expressionParseCases = append(benchmarkCases, moreCases...)

// params are parameters referenced when using test roll queries. These should
// be passed into test contexts.
var params map[string]string = map[string]string{
	"foo":     "1",
	"bar":     "2+1",
	"foo bar": "4",
	"baz":     "", // empty string is an unexpected value, but valid
}

type ASTTestCase struct {
	expression string
	ast        *Root
	wantErr    bool
}

var astCases = []ASTTestCase{
	// TODO: ensure that defaults are tested as well
	{"", nil, true},
	{"1", &Root{Expr: &Expr{L: &Term{Number: ptr(1.0)}}}, false},
	{"1d20", &Root{Expr: &Expr{L: &Term{Dice: &Dice{Count: ptr(1), Size: ptr(20)}}}}, false},
	{"d20", &Root{Expr: &Expr{L: &Term{Dice: &Dice{Count: ptr(1), Size: ptr(20)}}}}, false},
	{"1d20+1", &Root{Expr: &Expr{L: &Term{Dice: &Dice{Count: ptr(1), Size: ptr(20)}}, R: []*OpTerm{{Op: "+", Term: &Term{Number: ptr(1.0)}}}}}, false},
	{"1 //test", &Root{Expr: &Expr{L: &Term{Number: ptr(1.0)}}, Comment: ptr("test")}, false},
	{"1 // te st ", &Root{Expr: &Expr{L: &Term{Number: ptr(1.0)}}, Comment: ptr("te st ")}, false}, // parser doesn't sanitize
	{"0+.1", &Root{Expr: &Expr{L: &Term{Number: ptr(0.0)}, R: []*OpTerm{{Op: "+", Term: &Term{Number: ptr(0.1)}}}}}, false},
	{"1--1", &Root{Expr: &Expr{L: &Term{Number: ptr(1.0)}, R: []*OpTerm{{Op: "-", Term: &Term{Number: ptr(-1.0)}}}}}, false},
	{"1d20[foo]", &Root{Expr: &Expr{L: &Term{Dice: &Dice{Count: ptr(1), Size: ptr(20)}, Label: ptr("foo")}}}, false},
}

func TestParseString(t *testing.T) {
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
			if !reflect.DeepEqual(got, tt.ast) {
				gotBytes, err := json.Marshal(got)
				if err != nil {
					panic(err)
				}
				wantBytes, err := json.Marshal(tt.ast)
				if err != nil {
					panic(err)
				}
				t.Errorf("got = %v, want %v", string(gotBytes), string(wantBytes))
			}
			global = got
		})
	}
}

func TestParseString_linting(t *testing.T) {
	tests := []struct {
		base     string
		compare  string
		wantSame bool // true is good
	}{
		{"3d6#bar", " 3 d6  //  bar", true},
		{`3d6\bar`, " 3 d6 // bar", true},

		// Fail cases:
		{"3d6//bar", "3 d6 ", false},       // no Comment
		{"3d6#bar", "3 d6#bar ", false},    // no space after Comment
		{`3d6\bar`, " 3 d6 # bar ", false}, // extra space in Comment
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			want, err := ParseString(tt.base, false)
			if err != nil {
				panic(err)
			}
			got, err := ParseString(tt.compare, false)
			if err != nil {
				panic(err)
			}
			if reflect.DeepEqual(want, got) != tt.wantSame {
				wantBytes, err := json.Marshal(want)
				if err != nil {
					panic(err)
				}
				gotBytes, err := json.Marshal(got)
				if err != nil {
					panic(err)
				}
				t.Errorf("got = %v, want %v", string(gotBytes), string(wantBytes))
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

func TestSanitize(t *testing.T) {
	type args struct {
		expression string
	}
	tests := []struct {
		name          string
		args          args
		wantSanitized string
	}{
		{"empty", args{""}, ""},
		{"whitespace", args{" \t\n "}, ""},
		{"trailing", args{"foo  "}, "foo"},
		{"comment", args{"// comment "}, "// comment"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSanitized := Sanitize(tt.args.expression); gotSanitized != tt.wantSanitized {
				t.Errorf("Sanitize() = %v, want %v", gotSanitized, tt.wantSanitized)
			}
		})
	}
}
