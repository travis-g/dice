package dice

import (
	"encoding/json"
	"testing"
)

func BenchmarkParse3d20(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Parse("3d20")
	}
}

func TestParse(t *testing.T) {
	testCases := []struct {
		notation string
		count    int
		size     int
		output   string
	}{
		{"1d20", 1, 20, "d20"},
		{"d20", 1, 20, "d20"},
		{"3d20", 3, 20, "3d20"},
	}
	for _, tc := range testCases {
		dice, err := Parse(tc.notation)
		dice.sum()
		json, _ := json.Marshal(dice)
		if err != nil {
			t.Fatalf("failed to parse %q", tc.notation)
		}
		t.Logf("parsed %s: got %v", tc.notation, string(json))
		if dice.Size != tc.size {
			t.Errorf("parsed %s; want size %d, got size %d", tc.notation, tc.size, dice.Size)
		}
		if dice.Count != tc.count {
			t.Errorf("parsed %s; want count %d, got count %d", tc.notation, tc.count, dice.Count)
		}
		if output := dice.Notation(); output != tc.output {
			t.Errorf("parsed notation %s; want %s, got %s", tc.notation, tc.output, output)
		}
		if dice.Result < dice.Count {
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
			&Die{20, 1},
			&Die{20, 2},
			&Die{20, 3},
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
	}
	for _, tc := range testCases {
		for i := 0; i < 100; i++ {
			dice, err := Parse(tc.notation)
			if err != nil {
				t.Fatalf("failed to parse %q", tc.notation)
			}
			t.Logf("parsed %s: got %v", tc.notation, dice)
			if dice.Result < dice.Count {
				t.Errorf("parsed notation %s; got result %d which is less than count %d", tc.notation, dice.Result, dice.Count)
			}
			if dice.Result > dice.Size*dice.Count {
				t.Errorf("parsed notation %s; got result %d which should be impossible", tc.notation, dice.Result)
			}
		}
	}
}
