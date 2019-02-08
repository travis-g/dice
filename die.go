package dice

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

var _ Rollable = (*Die)(nil)
var _ Rollable = (*DieSet)(nil)
var _ RollableSet = (*DieSet)(nil)

// A Die represents a variable-sided die in memory, including the result of
// rolling it.
type Die struct {
	kind   string
	Result int `json:"result"`
	rolled bool
	Size   int `json:"size"`
}

// NewDie creates and returns a rolled die of size [1, size]. It panics if size
// < 1.
func NewDie(size int) (*Die, error) {
	if size < 1 {
		return nil, fmt.Errorf("dice: call to setSize with size < 1")
	}
	d := &Die{
		Size: size,
		kind: strings.Join([]string{"d", strconv.Itoa(size)}, ""),
	}
	d.Roll()
	return d, nil
}

// String returns an expression-like representation of a rolled Die or the kind
// of die if it has not been rolled.
func (d *Die) String() string {
	if d.rolled {
		return strconv.Itoa(d.Result)
	}
	return d.kind
}

// Roll will Roll a given Die (if unrolled) and set the die's result. Results
// are in the range [1, size].
func (d *Die) Roll() (float64, error) {
	if !d.rolled {
		i, err := Intn(d.Size)
		if err != nil {
			return 0, err
		}
		d.Result = 1 + i
		d.rolled = true
	}
	return float64(d.Result), nil
}

// Type returns the die type
func (d Die) Type() string {
	// Largest die is d2147483647
	return d.kind
}

// A DieSet set is a group of like-sided dice from a dice notation string
type DieSet struct {
	Count    uint    `json:"count"`
	Dice     []*Die  `json:"dice,omitempty"`
	Drop     int     `json:"drop,omitempty"`
	Expanded string  `json:"expanded"`
	Result   float64 `json:"result"`
	Size     int     `json:"size"`
}

// NewDieSet creates a new DieSet.
func NewDieSet(size int, count uint) DieSet {
	dice := make([]*Die, count)
	results := make([]int, count)
	sum := 0
	for i := range dice {
		die, err := NewDie(size)
		if err != nil {
			continue
		}
		dice[i] = die
		results[i] = die.Result
		sum += die.Result
	}
	return DieSet{
		Count:    count,
		Dice:     dice,
		Expanded: expression(results),
		Result:   (float64)(sum),
		Size:     size,
	}
}

func (d DieSet) String() string {
	return strings.Join([]string{d.Expanded, "=>", strconv.FormatFloat(d.Result, 'f', -1, 64)}, " ")
}

// Roll rolls the dice within a Dice set and sums the result with `Sum()`
func (d *DieSet) Roll() (float64, error) {
	for _, d := range d.Dice {
		_, err := d.Roll()
		if err != nil {
			return 0, err
		}
	}
	return float64(d.Sum()), nil
}

// Type returns the Dice type
func (d DieSet) Type() string {
	return strings.Join([]string{"d", strconv.Itoa(d.Size)}, "")
}

// Notation returns the dice notation format of the dice group in the format
// XdY, where X is the count of dice to roll and Y is the size of the dice
func (d DieSet) Notation() string {
	var s bytes.Buffer

	if l := len(d.Dice); l > 1 {
		s.WriteString(strconv.Itoa(l))
	}
	s.WriteString(strings.Join([]string{"d", strconv.Itoa(d.Size)}, ""))

	return s.String()
}

func sumDice(dice []*Die) int {
	sum := 0
	for _, d := range dice {
		sum += d.Result
	}
	return sum
}

// Sum returns and sets the total of a rolled dice set
func (d DieSet) Sum() float64 {
	d.Result = (float64)(sumDice(d.Dice))
	return d.Result
}
