package math

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/travis-g/dice"
)

// These functions must take interfaces as arguments since they must be valid
// govaluate.ExpressionFunctions.
func max(args ...interface{}) (interface{}, error) {
	sort.Slice(args[:], func(i, j int) bool {
		return args[i].(float64) < args[j].(float64)
	})
	return args[len(args)-1], nil
}
func min(args ...interface{}) (interface{}, error) {
	sort.Slice(args[:], func(i, j int) bool {
		return args[i].(float64) < args[j].(float64)
	})
	return args[0], nil
}

var (
	// DiceFunctions are functions usable in dice arithmetic operations, such as
	// `round()`, `min()`, and `max()`.
	DiceFunctions = map[string]govaluate.ExpressionFunction{
		"abs": func(args ...interface{}) (interface{}, error) {
			return math.Abs(args[0].(float64)), nil
		},
		"ceil": func(args ...interface{}) (interface{}, error) {
			return math.Ceil(args[0].(float64)), nil
		},
		"floor": func(args ...interface{}) (interface{}, error) {
			return math.Floor(args[0].(float64)), nil
		},
		"max": max,
		"min": min,
		"round": func(args ...interface{}) (interface{}, error) {
			return math.Round(args[0].(float64)), nil
		},
	}
)

// A DiceExpression is a representation of a dice roll that must be evaluated.
// This may be a simple expression like `d20` or more complex, like
// `floor(max(d20,d12)/2+3)`.
type DiceExpression struct {
	Original string             `json:"original"`
	Rolled   string             `json:"rolled"`
	Result   float64            `json:"result"`
	Dice     []dice.RollableSet `json:"dice"`
}

func (de *DiceExpression) String() string {
	// HACK(tssde71): since de.Result is a float Sprint is the easiest way to
	// auto-truncate any unnecessary decimals
	return strings.Join([]string{de.Rolled, "=", fmt.Sprint(de.Result)}, " ")
}

// Evaluate evaluates a string expression of dice and math, returning a synopsis
// of the various stages of evaluation and/or an error. The evaluation order
// needs to follow order of operations:
//
//  1. Roll all dice by matching for regex and substituting the roll values,
//  2. Perform any function-based operations (adv, dis, floor),
//  3. Multiplication/division,
//  4. Addition/subtraction
//
// Evaluate can likely benefit immensely from optimization and more fine-grained
// unit tests/benchmarks.
func Evaluate(expression string) (*DiceExpression, error) {
	de := &DiceExpression{
		Original: expression,
		Dice:     make([]dice.RollableSet, 0),
	}
	// systematically parse the DiceExpression for dice notation substrings,
	// evaluate and expand the rolls, replace the notation strings with their
	// fully-rolled and expanded counterparts, and save the expanded expression
	// to the object.
	rolledBytes := dice.DiceNotationRegex.ReplaceAllFunc([]byte(de.Original), func(matchBytes []byte) []byte {
		d, err := dice.Parse(string(matchBytes))
		// record dice:
		de.Dice = append(de.Dice, d)
		if err != nil {
			return []byte(``)
		}
		// write expanded result back as bytes
		var buf bytes.Buffer
		buf.WriteString(`(`)
		buf.WriteString(dice.Expand(d))
		buf.WriteString(`)`)
		return buf.Bytes()
	})
	de.Rolled = string(rolledBytes)

	// populate the expression object with the roll and function data
	exp, err := govaluate.NewEvaluableExpressionWithFunctions(de.Rolled, DiceFunctions)
	if err != nil {
		return nil, err
	}

	// get and set the result
	result, err := exp.Evaluate(nil)
	if err != nil {
		return nil, err
	}
	// result should be a float
	var ok bool
	if de.Result, ok = result.(float64); ok {
		return de, nil
	}

	return nil, fmt.Errorf("error evaluating roll")
}
