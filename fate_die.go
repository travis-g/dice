package dice

import (
	"strconv"
	"strings"
)

var _ Rollable = (*FateDie)(nil)
var _ Rollable = (*FateDieSet)(nil)
var _ RollableSet = (*FateDieSet)(nil)

// A FateDie (a.k.a. "Fudge die") is a die with six sides, {-1, -1, 0, 0, 1, 1}.
// In a pinch, a FateDie can be emulated by evaluating `1d3-2`.
type FateDie struct {
	rolled bool
	Result int `json:"result"`
}

// NewFateDie create and returns a new FateDie. The error will always be nil.
func NewFateDie() (*FateDie, error) {
	f := new(FateDie)
	f.Roll()
	return f, nil
}

func (f FateDie) String() string {
	return (string)(f.Result)
}

// Roll will Roll a given FateDie and set the die's result. Fate dice can have
// results in [-1, 1].
func (f *FateDie) Roll() (float64, error) {
	if !f.rolled {
		i, err := Intn(3)
		if err != nil {
			return 0, err
		}
		f.Result = i - 2
		f.rolled = true
	}
	return (float64)(f.Result), nil
}

// Type returns the FateDie's type
func (f FateDie) Type() string {
	return "dF"
}

// A FateDieSet set is a group of fate/fudge dice from a notation
type FateDieSet struct {
	Count    uint      `json:"count"`
	Dice     []FateDie `json:"dice,omitempty"`
	Drop     int       `json:"drop,omitempty"`
	Expanded string    `json:"expanded"`
	Result   float64   `json:"result"`
}

// NewFateDieSet creates and returns a rolled FateDieSet.
func NewFateDieSet(count uint) FateDieSet {
	dice := make([]FateDie, count)
	results := make([]int, count)
	sum := 0
	for i := range dice {
		die, err := NewFateDie()
		if err != nil {
			continue
		}
		dice[i] = *die
		results[i] = die.Result
		sum += die.Result
	}
	return FateDieSet{
		Count:    count,
		Dice:     dice,
		Expanded: expression(results),
		Result:   (float64)(sum),
	}
}

func (d FateDieSet) String() string {
	return strings.Join([]string{d.Expanded, "=>", strconv.FormatFloat(d.Result, 'f', -1, 64)}, " ")
}

// Roll rolls the dice within a FateDieSet.
func (d *FateDieSet) Roll() (float64, error) {
	for _, d := range d.Dice {
		_, err := d.Roll()
		if err != nil {
			return 0, err
		}
	}
	return d.Sum(), nil
}

// Type returns the type of the dice within a FateDieSet, which will always be
// "dF".
func (d FateDieSet) Type() string {
	return "dF"
}

func sumFateDice(dice []FateDie) int {
	sum := 0
	for _, d := range dice {
		sum += d.Result
	}
	return sum
}

// Sum returns and sets the total of a rolled dice set
func (d FateDieSet) Sum() float64 {
	d.Result = (float64)(sumFateDice(d.Dice))
	return d.Result
}
