package dice

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var _ Interface = (*RollableDie)(nil)
var _ Interface = (*DieSet)(nil)
var _ = Interface(&Die{})
var _ RollableSet = (*DieSet)(nil)

// A Die represents a variable-sided die in memory, including the result of
// rolling it.
type Die struct {
	Interface `json:"self,omitempty"`
	Type      string  `json:"type"`
	Result    float64 `json:"result"`
	Size      int     `json:"size"`
	Dropped   bool    `json:"dropped,omitempty"`
	Unrolled  bool    `json:"unrolled,omitempty"`
}

// RollableDie is a wrapper around Die that implements Rollable.
type RollableDie struct {
	*Die
}

// Get returns the wrapped Die.
func (r *RollableDie) Get() *Die {
	return r.Die
}

// Result returns the wrapped Die's result. If the Die had not been rolled an
// error will be returned.
func (r *RollableDie) Result() (float64, error) {
	die := r.Get()
	if !die.Unrolled {
		return (float64)(die.Result), nil
	}
	return 0, errors.New("unrolled die")
}

func (r *RollableDie) String() string {
	die := r.Get()
	if !die.Unrolled {
		return fmt.Sprintf("%f", die.Result)
	}
	return die.Type
}

// Type returns the kind of Die a RollableDie is.
func (r *RollableDie) Type() string {
	return r.Get().Type
}

// Total returns the RollableDie's total (its Result)
func (r *RollableDie) Total() float64 {
	t, _ := r.Result()
	return t
}

// NewDie creates and returns a rolled die of size [1, size]. It panics if size
// < 1.
func NewDie(size int) (RollableDie, error) {
	if size < 1 {
		return RollableDie{&Die{}}, fmt.Errorf("dice: call to setSize with size < 1")
	}
	d := &Die{
		Size:     size,
		Type:     strings.Join([]string{"d", strconv.Itoa(size)}, ""),
		Unrolled: true,
	}
	d.Roll()
	return RollableDie{Die: d}, nil
}

// String returns an expression-like representation of a rolled Die or the kind
// of die if it has not been rolled.
func (d *Die) String() string {
	if !d.Unrolled {
		return fmt.Sprintf("%v", d.Result)
	}
	return d.Type
}

// GoString prints the Go syntax of a Die.
func (d *Die) GoString() string {
	return fmt.Sprintf("%#v", *d)
}

// Total returns the result of a die. If dropped, 0 is returned.
func (d *Die) Total() float64 {
	if d.Dropped {
		return 0.0
	}
	if d.Unrolled {
		d.Roll()
	}
	return d.Result
}

// Roll will Roll a given Die (if unrolled) and set the die's result. Results
// are in the range [1, size]. If the die already has a result it will not be
// rerolled.
func (d *Die) Roll() (float64, error) {
	if !d.Unrolled {
		return d.Result, nil
	}
	if d.Result == 0 {
		i, err := Intn(d.Size)
		if err != nil {
			return 0, err
		}
		d.Result = (float64)(1 + i)
		d.Unrolled = false
	}
	return d.Result, nil
}

// A DieSet set is a group of like-sided dice from a dice notation string
type DieSet struct {
	Count    uint          `json:"count"`
	Dice     []RollableDie `json:"dice,omitempty"`
	Drop     int           `json:"drop,omitempty"`
	Expanded string        `json:"expanded"`
	Result   float64       `json:"result"`
	Size     int           `json:"size"`
}

// NewDieSet creates a new DieSet.
func NewDieSet(size int, count uint, drop int) DieSet {
	dice := make([]RollableDie, count)
	// create and roll dice
	for i := range dice {
		die, err := NewDie(size)
		if err != nil {
			continue
		}
		dice[i] = die
	}

	if drop != 0 {
		// sort the dice ascending
		//
		// HACK(travis-g): modify the function to sort
		// an array of pointers to the dice so that the original order is not
		// modified but the dice properties can but updated directly
		sort.Slice(dice, func(i, j int) bool {
			return dice[i].Get().Result < dice[j].Get().Result
		})
		// drop lowest to highest
		if drop > 0 {
			for i := 0; i < drop; i++ {
				dice[i].Get().Dropped = true
			}
			// drop highest to lowest
		} else if drop < 0 {
			for i := len(dice) - 1; i >= len(dice)+drop; i-- {
				dice[i].Get().Dropped = true
			}
		}
	}

	// total the undropped dice
	results := make([]float64, count)
	sum := 0.0
	for i, die := range dice {
		if die.Get().Dropped {
			continue
		}
		result := die.Get().Result
		results[i] = result
		sum += result
	}
	return DieSet{
		Count:    count,
		Dice:     dice,
		Expanded: expression(results),
		Result:   (float64)(sum),
		Size:     size,
		Drop:     drop,
	}
}

func (d DieSet) String() string {
	return fmt.Sprintf("%s => %v", d.Expanded, d.Result)
}

// GoString prints a viable golang code representation of a DieSet.
func (d *DieSet) GoString() string {
	return fmt.Sprintf("%#v", *d)
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

// Total returns the DieSet's total.
func (d DieSet) Total() float64 {
	return d.Result
}

// Notation returns the dice notation format of the dice group in the format
// XdY, where X is the count of dice to roll and Y is the size of the dice
func (d DieSet) Notation() string {
	var s bytes.Buffer

	if l := len(d.Dice); l > 1 {
		s.WriteString(strconv.Itoa(l))
	}
	s.WriteString(fmt.Sprintf("d%d", d.Size))

	return s.String()
}

// Expression returns the expanded expression of a DieSet.
func (d DieSet) Expression() string {
	return d.Expanded
}

func sumDice(dice []RollableDie) int {
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
func (d DieSet) Sum() float64 {
	d.Result = (float64)(sumDice(d.Dice))
	return d.Result
}
