package math

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"sort"

	eval "github.com/Knetic/govaluate"
	"github.com/travis-g/dice"
)

// Possible error types for mathematical functions.
var (
	ErrNotEnoughArgs   = errors.New("not enough args")
	ErrInvalidArgCount = errors.New("invalid argument count")
)

// DiceFunctions are functions usable in dice arithmetic operations, such as
// round, min, and max.
//
// TODO: adv() and dis()
var DiceFunctions = map[string]eval.ExpressionFunction{
	"abs":   absExpressionFunction,
	"ceil":  ceilExpressionFunction,
	"floor": floorExpressionFunction,
	"max":   maxExpressionFunction,
	"min":   minExpressionFunction,
	"round": roundExpressionFunction,
}

func absExpressionFunction(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return 0, ErrInvalidArgCount
	}
	return math.Abs(args[0].(float64)), nil
}

func ceilExpressionFunction(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return 0, ErrInvalidArgCount
	}
	return math.Ceil(args[0].(float64)), nil
}

func floorExpressionFunction(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return 0, ErrInvalidArgCount
	}
	return math.Floor(args[0].(float64)), nil
}

func maxExpressionFunction(args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return 0, ErrNotEnoughArgs
	}
	sort.Slice(args[:], func(i, j int) bool {
		return args[i].(float64) < args[j].(float64)
	})
	return args[len(args)-1], nil
}

func minExpressionFunction(args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return 0, ErrNotEnoughArgs
	}
	sort.Slice(args[:], func(i, j int) bool {
		return args[i].(float64) < args[j].(float64)
	})
	return args[0], nil
}

func roundExpressionFunction(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return 0, ErrInvalidArgCount
	}
	return math.Round(args[0].(float64)), nil
}

// An ExpressionResult is a representation of a dice roll that has been
// evaluated.
type ExpressionResult struct {
	// Original is the original expression input.
	Original string `json:"original"`

	// Rolled is the original expression but with any dice expressions rolled
	// and expanded.
	Rolled string `json:"rolled"`

	// Result is the expression's evaluated total.
	Result float64 `json:"result"`

	// Dice is the list of dice groups rolled as part of the expression. As dice
	// are rolled, their GroupProperties are retrieved.
	Dice []*dice.RollerGroup `json:"dice,omitempty"`
}

// String implements fmt.Stringer.
func (de *ExpressionResult) String() string {
	if de == nil {
		return ""
	}
	return fmt.Sprintf("%s = %v", de.Rolled, de.Result)
}

// GoString implements fmt.GoStringer.
func (de *ExpressionResult) GoString() string {
	return fmt.Sprintf("%#v", *de)
}

/*
Evaluate evaluates a string expression of dice, math, or a combination of the
two, and returns the resulting ExpressionResult. The evaluation order needs to
follow order of operations.

The expression passed must evaluate to a float64 result. A parsable expression
could be a simple expression or more complex.

    d20
    2d20dl1+5
    4d6-3d5+30
    min(d20,d20)+1
    floor(max(d20,2d12k1)/2+3)

Evaluate can likely benefit immensely from optimization and a custom parser
implementation along with more fine-grained unit tests/benchmarks.
*/
func Evaluate(ctx context.Context, expression string) (*ExpressionResult, error) {
	de := &ExpressionResult{
		Original: expression,
		Dice:     make([]*dice.RollerGroup, 0),
	}

	var evalErrors = []error{}

	// systematically parse the DiceExpression for dice notation substrings,
	// evaluate and expand the rolls, replace the notation strings with their
	// fully-rolled and expanded counterparts, and save the expanded expression
	// to the object.
	rolledBytes := dice.DiceWithModifiersExpressionRegex.ReplaceAllFunc([]byte(de.Original), func(matchBytes []byte) []byte {
		props, err := dice.ParseExpression(ctx, string(matchBytes))
		if err != nil {
			evalErrors = append(evalErrors, err)
			return []byte{}
		}
		d, err := dice.NewRollerGroup(&props)
		if err != nil {
			evalErrors = append(evalErrors, err)
			return []byte{}
		}
		d.FullRoll(ctx)
		// record dice:
		de.Dice = append(de.Dice, d)

		// write expanded result back as bytes
		var buf bytes.Buffer
		write := buf.WriteString
		write(`(`)
		write(d.Expression())
		write(`)`)
		return buf.Bytes()
	})
	if len(evalErrors) != 0 {
		return nil, fmt.Errorf("errors during parsing: %v", evalErrors)
	}
	de.Rolled = string(rolledBytes)

	// populate the expression object with the roll and function data
	exp, err := eval.NewEvaluableExpressionWithFunctions(de.Rolled, DiceFunctions)
	if err != nil {
		return nil, err
	}

	// get and set the result
	result, err := exp.Evaluate(nil)
	if err != nil {
		return nil, dice.ErrInvalidExpression
	}
	// result should be a float
	var ok bool
	if de.Result, ok = result.(float64); !ok {
		return de, fmt.Errorf("result %v not a float", err)
	}

	return de, nil
}
