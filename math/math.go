package math

import (
	"fmt"
	"math"
	"sort"

	"github.com/Knetic/govaluate"
	"github.com/travis-g/dice"
)

// Advantage takes a slice of Dice pointers and returns the first Dice object
// that is the last occurrence of the highest roll in the slice.
func Advantage(rolls ...*dice.Dice) *dice.Dice {
	sort.Slice(rolls[:], func(i, j int) bool {
		return rolls[i].Result < rolls[j].Result
	})
	return rolls[len(rolls)-1]
}

// Disadvantage has the same functionality as Advantage, but returns the first
// occurrence of the lowest roll in the slice.
func Disadvantage(rolls ...*dice.Dice) *dice.Dice {
	sort.Slice(rolls[:], func(i, j int) bool {
		return rolls[i].Result < rolls[j].Result
	})
	return rolls[0]
}

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
		"int": func(args ...interface{}) (interface{}, error) {
			return args[0].(int), nil
		},
		"max": max,
		"min": min,
		"round": func(args ...interface{}) (interface{}, error) {
			return math.Round(args[0].(float64)), nil
		},
	}
)

// A DiceExpression is a representation of a dice roll that must be evaluated.
// This may be as simple as `d20` or as complex as `floor(max(d20,d12)/2+3)`.
type DiceExpression struct {
	Original string       `json:"original"`
	Rolled   string       `json:"rolled"`
	Result   float64      `json:"result"`
	Dice     []*dice.Dice `json:"dice"`
}

// Evaluate will calculate the result of dice expression.
func (de *DiceExpression) Evaluate() error {
	faux, err := Evaluate(de.Original)
	if err != nil {
		return err
	}
	de = faux
	return dice.NewErrNotImplemented("not implemented")
}

// Evaluate evaluates a string expression of dice and math, returning a synopsis of
// the various stages of evaluation and/or an error. The evaluation order needs
// to follow order of operations:
//
//     1. Roll all dice by matching for regex and substituting the roll values,
//     2. Perform any function-based operations (adv, dis, floor),
//     3. Multiplication/division,
//     4. Addition/subtraction,
func Evaluate(expression string) (*DiceExpression, error) {
	de := &DiceExpression{
		Original: expression,
		Dice:     make([]*dice.Dice, 0),
	}
	rolledBytes := dice.DiceNotationRegex.ReplaceAllFunc([]byte(de.Original), func(matchBytes []byte) []byte {
		d, err := dice.Parse(string(matchBytes))
		// record dice:
		de.Dice = append(de.Dice, d)
		if err != nil {
			return []byte(``)
		}
		return []byte(fmt.Sprintf("%d", d.Result))
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
