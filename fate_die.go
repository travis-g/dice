package dice

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const fateDieNotation = "dF"

var _ Rollable = (*RollableFateDie)(nil)
var _ Rollable = (*FateDieSet)(nil)
var _ RollableSet = (*FateDieSet)(nil)

// A FateDie (a.k.a. "Fudge die") is a die with six sides, {-1, -1, 0, 0, 1, 1}.
// In a pinch, a FateDie can be emulated by evaluating `1d3-2`.
type FateDie struct {
	rolled bool
	Result int    `json:"result"`
	Type   string `json:"type"`
}

// RollableFateDie is a wrapper around FateDie that implements Rollable.
type RollableFateDie struct {
	*FateDie
}

// Get returns the wrapped FateDie.
func (r *RollableFateDie) Get() *FateDie {
	return r.FateDie
}

// Result returns the wrapped FateDie's result. If the FateDie had not been
// rolled an error will be returned.
func (r *RollableFateDie) Result() (float64, error) {
	die := r.Get()
	if die.rolled {
		return (float64)(die.Result), nil
	}
	return 0, errors.New("unrolled die")
}

func (r *RollableFateDie) String() string {
	die := r.Get()
	if die.rolled {
		return strconv.Itoa(die.Result)
	}
	return fateDieNotation
}

// Type returns the Fate die's type, which will always be "dF".
func (r *RollableFateDie) Type() string {
	return fateDieNotation
}

// NewFateDie create and returns a new FateDie.
func NewFateDie() (RollableFateDie, error) {
	f := &FateDie{
		Type: fateDieNotation,
	}
	f.Roll()
	return RollableFateDie{FateDie: f}, nil
}

func (f *FateDie) String() string {
	if f.rolled {
		return strconv.Itoa(f.Result)
	}
	return fateDieNotation
}

// GoString prints a viable golang code representation of a FateDie.
func (f *FateDie) GoString() string {
	return fmt.Sprintf("%#v", *f)
}

// Roll will Roll a given FateDie and set the die's result. Fate dice can have
// results in [-1, 1].
func (f *FateDie) Roll() (float64, error) {
	if !f.rolled {
		i, err := Intn(3)
		if err != nil {
			return 0, err
		}
		f.Result = i - 1
		f.rolled = true
	}
	return (float64)(f.Result), nil
}

// A FateDieSet set is a group of fate/fudge dice from a notation
type FateDieSet struct {
	Count    uint              `json:"count"`
	Dice     []RollableFateDie `json:"dice,omitempty"`
	Drop     int               `json:"drop,omitempty"`
	Expanded string            `json:"expanded"`
	Result   float64           `json:"result"`
}

// NewFateDieSet creates and returns a rolled FateDieSet.
func NewFateDieSet(count uint) FateDieSet {
	dice := make([]RollableFateDie, count)
	results := make([]int, count)
	sum := 0
	for i := range dice {
		die, err := NewFateDie()
		if err != nil {
			continue
		}
		dice[i] = die
		result := die.Get().Result
		results[i] = result
		sum += result
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

// GoString prints a viable golang code representation of a FateDieSet.
func (d *FateDieSet) GoString() string {
	return fmt.Sprintf("%#v", *d)
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

// Type returns the type of the dice within a FateDieSet, which will always be "dF".
func (d FateDieSet) Type() string {
	return "dF"
}

func sumFateDice(dice []RollableFateDie) int {
	sum := 0
	for _, d := range dice {
		result, err := d.Result()
		if err != nil {
			return 0
		}
		sum += (int)(result)
	}
	return sum
}

// Sum returns and sets the total of a rolled dice set
func (d FateDieSet) Sum() float64 {
	d.Result = (float64)(sumFateDice(d.Dice))
	return d.Result
}
