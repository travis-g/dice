package dice

import (
	"fmt"
	"testing"
)

// package-level variable to prevent optimizations
var i interface{}

func TestDieSetString(t *testing.T) {
	testCases := []struct {
		dice DieSet
		str  string
	}{
		{DieSet{
			Dice: []RollableDie{
				RollableDie{&Die{Size: 20, Result: 1}},
				RollableDie{&Die{Size: 20, Result: 2}},
				RollableDie{&Die{Size: 20, Result: 3}},
			},
			Expanded: expression(1, 2, 3),
			Result:   6.0,
		}, "1+2+3 => 6"},
		{DieSet{
			Dice: []RollableDie{
				RollableDie{&Die{Size: 6, Result: 1}},
				RollableDie{&Die{Size: 6, Result: 3}},
				RollableDie{&Die{Size: 6, Result: 4}},
			},
			Expanded: expression(1, 3, 4),
			Result:   8.0,
		}, "1+3+4 => 8"},
	}
	for _, tc := range testCases {
		str := fmt.Sprintf(tc.dice.String())
		if str != tc.str {
			t.Errorf("want result %s, got %s", tc.str, str)
		}
	}
}

// Benchmarks

var diceNotationStrings = []struct {
	notation string
}{
	{"d20"},
	{"1d20"},
}
