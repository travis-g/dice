package dice

import (
	"encoding/json"
	"testing"
)

// package-level variable to prevent optimizations
var i interface{}

func BenchmarkParse3d20(b *testing.B) {
	b.ReportAllocs()
	var d *Dice
	for n := 0; n < b.N; n++ {
		d, _ = Parse("3d20")
	}
	i = d
}

func BenchmarkNewFateDie(b *testing.B) {
	b.ReportAllocs()
	var f *FateDie
	for n := 0; n < b.N; n++ {
		f, _ = NewFateDie()
	}
	i = f
}

func TestRollableInterfaces(t *testing.T) {
	var rollables = []Rollable{
		&Die{},
		&Dice{},
		&FateDie{},
	}
	t.Log(rollables)
}

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
		if dice.Result < int(dice.Count) {
			t.Errorf("parsed notation %s; got result %d which is less than count %d", tc.notation, dice.Result, dice.Count)
		}
	}
}

func TestSumDice(t *testing.T) {
	testCases := []struct {
		dice  []*Die
		total int
	}{
		{[]*Die{
			&Die{Size: 20, Result: 1},
			&Die{Size: 20, Result: 2},
			&Die{Size: 20, Result: 3},
		}, 6},
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
			if dice.Result < int(dice.Count) {
				t.Errorf("parsed notation %s; got result %d which is less than count %d", tc.notation, dice.Result, dice.Count)
			}
			if dice.Result > dice.Size*int(dice.Count) {
				t.Errorf("parsed notation %s; got result %d which should be impossible", tc.notation, dice.Result)
			}
		}
	}
}
