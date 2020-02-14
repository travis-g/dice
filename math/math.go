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
	"abs": func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return 0, ErrInvalidArgCount
		}
		return math.Abs(args[0].(float64)), nil
	},
	"ceil": func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return 0, ErrInvalidArgCount
		}
		return math.Ceil(args[0].(float64)), nil
	},
	"floor": func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return 0, ErrInvalidArgCount
		}
		return math.Floor(args[0].(float64)), nil
	},
	"max": func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return 0, ErrNotEnoughArgs
		}
		sort.Slice(args[:], func(i, j int) bool {
			return args[i].(float64) < args[j].(float64)
		})
		return args[len(args)-1], nil
	},
	"min": func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return 0, ErrNotEnoughArgs
		}
		sort.Slice(args[:], func(i, j int) bool {
			return args[i].(float64) < args[j].(float64)
		})
		return args[0], nil
	},
	"round": func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return 0, ErrInvalidArgCount
		}
		return math.Round(args[0].(float64)), nil
	},
}

// An Expression is a representation of a dice roll that has been evaluated.
type Expression struct {
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
func (de *Expression) String() string {
	if de == nil {
		return ""
	}
	return fmt.Sprintf("%s = %v", de.Rolled, de.Result)
}

// GoString implements fmt.GoStringer.
func (de *Expression) GoString() string {
	return fmt.Sprintf("%#v", *de)
}

/*
Evaluate evaluates a string expression of dice, math, or a combination of the
two, and returns the resulting Expression. The evaluation order needs to follow
order of operations.

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
func Evaluate(ctx context.Context, expression string) (*Expression, error) {
	de := &Expression{
		Original: expression,
		Dice:     make([]*dice.RollerGroup, 0),
	}
	// systematically parse the DiceExpression for dice notation substrings,
	// evaluate and expand the rolls, replace the notation strings with their
	// fully-rolled and expanded counterparts, and save the expanded expression
	// to the object.
	rolledBytes := dice.DiceWithModifiersExpressionRegex.ReplaceAllFunc([]byte(de.Original), func(matchBytes []byte) []byte {
		props, _ := dice.ParseExpression(ctx, string(matchBytes))
		d, err := dice.NewRollerGroup(&props)
		if err != nil {
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
