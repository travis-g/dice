package math

import (
	"errors"
	"math"
	"sort"

	eval "github.com/Knetic/govaluate"
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

func ListDiceFunctions() []string {
	funcs := make([]string, 0, len(DiceFunctions))
	for name := range DiceFunctions {
		funcs = append(funcs, name)
	}
	return funcs
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
