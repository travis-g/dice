package dice

import (
	"encoding/json"
	"fmt"
	"testing"
)

// package-level variable to prevent optimizations
var i interface{}

func TestParse(t *testing.T) {
	testCases := []struct {
		notation string
		count    uint
		size     int
		output   string
	}{
		{"1d20", 1, 20, "d20"},
		{"d20", 1, 20, "d20"},
		{"3d20", 3, 20, "3d20"},
	}
	for _, tc := range testCases {
		dice, err := Parse(tc.notation)
		dice.Sum()
		json, _ := json.Marshal(dice)
		t.Logf("parsed %s; got %s", tc.notation, string(json))
		if err != nil {
			t.Fatalf("failed to parse %q", tc.notation)
		}
		if dice.Size != tc.size {
			t.Errorf("parsed %s; want size %d, got size %d", tc.notation, tc.size, dice.Size)
		}
		if dice.Count != tc.count {
			t.Errorf("parsed %s; want count %d, got count %d", tc.notation, tc.count, dice.Count)
		}
		if output := dice.Notation(); output != tc.output {
			t.Errorf("parsed notation %s; want %s, got %s", tc.notation, tc.output, output)
		}
		if dice.Result < float64(dice.Count) {
			t.Errorf("parsed notation %s; got result %f which is less than count %d", tc.notation, dice.Result, dice.Count)
		}
	}
}

func TestDieSetString(t *testing.T) {
	testCases := []struct {
		dice DieSet
		str  string
	}{
		{DieSet{
			Dice: []*Die{
				&Die{Size: 20, Result: 1},
				&Die{Size: 20, Result: 2},
				&Die{Size: 20, Result: 3},
			},
			Expanded: expression(1, 2, 3),
			Result:   6.0,
		}, "1+2+3 => 6"},
		{DieSet{
			Dice: []*Die{
				&Die{Size: 6, Result: 1},
				&Die{Size: 6, Result: 3},
				&Die{Size: 6, Result: 4},
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

func TestSumDieSet(t *testing.T) {
	testCases := []struct {
		dice  []*Die
		total int
	}{
		{[]*Die{
			&Die{Size: 20, Result: 1},
			&Die{Size: 20, Result: 2},
			&Die{Size: 20, Result: 3},
		}, 6},
		{[]*Die{
			&Die{Size: 6, Result: 1},
			&Die{Size: 6, Result: 3},
			&Die{Size: 6, Result: 4},
		}, 8},
	}
	for _, tc := range testCases {
		sum := sumDice(tc.dice)
		t.Logf("summed %v: got %d", tc.dice, sum)
		if sum != tc.total {
			t.Errorf("summed %v; want result %d, got %d", tc.dice, tc.total, sum)
		}
	}
}

func TestRoll(t *testing.T) {
	testCases := []struct {
		notation string
		count    int
		size     int
		output   string
	}{
		{"1d20", 1, 20, "d20"},
		{"d20", 1, 20, "d20"},
		{"3d20", 3, 20, "3d20"},
		{"20d20", 20, 20, "20d20"},
	}
	for _, tc := range testCases {
		for i := 0; i < 100; i++ {
			dice, err := Parse(tc.notation)
			if err != nil {
				t.Fatalf("failed to parse %q", tc.notation)
			}
			t.Logf("parsed %s: got %v", tc.notation, dice)
			if dice.Result < float64(dice.Count) {
				t.Errorf("parsed notation %s; got result %f which is less than count %d", tc.notation, dice.Result, dice.Count)
			}
			if dice.Result > float64(dice.Size*int(dice.Count)) {
				t.Errorf("parsed notation %s; got result %f which should be impossible", tc.notation, dice.Result)
			}
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

func BenchmarkParse(b *testing.B) {
	b.ReportAllocs()
	for _, tc := range diceNotationStrings {
		b.Run(tc.notation, func(b *testing.B) {
			notation := tc.notation
			for n := 0; n < b.N; n++ {
				Parse(notation)
			}
		})
	}
}
