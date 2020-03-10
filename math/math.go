package math

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	eval "github.com/Knetic/govaluate"
	"github.com/travis-g/dice"
)

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
	// as there could be a float/decimal result, format the float properly
	return fmt.Sprintf("%s = %s", de.Rolled, strconv.FormatFloat(de.Result, 'f', -1, 64))
}

// GoString implements fmt.GoStringer.
func (de *ExpressionResult) GoString() string {
	return fmt.Sprintf("%#v", *de)
}

/*
EvaluateExpression evaluates a string expression of dice, math, or a combination of the
two, and returns the resulting ExpressionResult. The evaluation order needs to
follow order of operations.

The expression passed must evaluate to a float64 result. A parsable expression
could be a simple expression or more complex.

    d20
    2d20dl1+5
    4d6-3d5+30
    min(d20,d20)+1
    floor(max(d20,2d12k1)/2+3)

EvaluateExpression can likely benefit immensely from optimization and a custom parser
implementation along with more fine-grained unit tests/benchmarks.
*/
func EvaluateExpression(ctx context.Context, expression string) (*ExpressionResult, error) {
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
		props, err := dice.ParseNotation(ctx, string(matchBytes))
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
		var b strings.Builder
		write := b.WriteString
		write(`(`)
		write(d.Expression())
		write(`)`)
		return []byte(b.String())
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
	if exp == nil {
		return nil, ErrNilExpression
	}

	// get and set the result
	result, err := exp.Evaluate(nil)
	if err != nil {
		return nil, dice.ErrInvalidExpression
	}
	if result == nil {
		return de, ErrNilResult
	}

	// result should be a float
	var ok bool
	if de.Result, ok = result.(float64); !ok {
		return de, fmt.Errorf("result %v not a float", err)
	}

	return de, nil
}

// Math package errors.
var (
	ErrNilExpression = errors.New("nil expression")
	ErrNilResult     = errors.New("nil result")
)

// ParseExpressionWithFunc
//
// rolledBytes, err := ParseExpressionWithFunc(ctx, dice.DiceWithModifiersExpressionRegex, de.Original, )
func ParseExpressionWithFunc(ctx context.Context, regexp *regexp.Regexp, expression string, repl func([]byte) []byte) (string, error) {
	bytes := regexp.ReplaceAllFunc([]byte(expression), repl)
	return string(bytes), nil
}
