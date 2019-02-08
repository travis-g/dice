package dice

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// DiceNotationRegex is the compiled RegEx for parsing supported dice
	// notations.
	DiceNotationRegex = regexp.MustCompile(`(?P<count>\d*)d(?P<size>(?:\d{1,}|F))`)

	// DropKeepNotationRegex is the compiled RegEx for parsing drop/keep dice
	// notations (unimplemented).
	DropKeepNotationRegex = regexp.MustCompile(`(?P<count>\d+)?d(?P<size>\d{1,})(?P<dropkeep>(?P<op>[dk][lh]?)(?P<num>\d{1,}))?`)
)

// A Rollable is any kind of rollable object. A Rollable could be a single die
// or many dice of any type.
type Rollable interface {
	// Roll should be used to also set the object's Result
	Roll() (float64, error)
	String() string
	Type() string
}

var ( // Validate that die types are Rollable
	_ Rollable = (*Die)(nil)
	_ Rollable = (*FateDie)(nil)
	_ Rollable = (*DieSet)(nil)
)

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

// String returns an expression-like representation of a rolled Die or the kind
// of die if it has not been rolled.
func (d *Die) String() string {
	if d.rolled {
		return strconv.Itoa(d.Result)
	}
	return d.kind
}

// A FateDie (a.k.a. "Fudge die") is a die with six sides, {-1, -1, 0, 0, 1, 1}.
// In a pinch, a FateDie can be emulated by evaluating `1d3-2`.
type FateDie struct {
	rolled bool
	Result int `json:"result"`
}

func (f FateDie) String() string {
	return string(f.Result)
}

// Type returns the FateDie's type
func (f FateDie) Type() string {
	return "dF"
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
	return float64(f.Result), nil
}

// NewFateDie create and returns a new FateDie. The error will always be nil.
func NewFateDie() (*FateDie, error) {
	f := new(FateDie)
	f.Roll()
	return f, nil
}

// A FateDieSet set is a group of fate/fudge dice from a notation
type FateDieSet struct {
	Count    uint       `json:"count"`
	Dice     []*FateDie `json:"dice,omitempty"`
	Drop     int        `json:"drop,omitempty"`
	Expanded string     `json:"expanded"`
	Result   float64    `json:"result"`
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

func (d DieSet) String() string {
	return strings.Join([]string{d.Expanded, "=>", strconv.FormatFloat(d.Result, 'f', -1, 64)}, " ")
}

// Type returns the Dice type
func (d DieSet) Type() string {
	return strings.Join([]string{"d", strconv.Itoa(d.Size)}, "")
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

// Roll rolls a set of rollables and returns the total.
func Roll(rollables ...Rollable) (float64, error) {
	for _, r := range rollables {
		r.Roll()
	}
	return sumRollables(rollables...)
}

func sumDice(dice []*Die) int {
	sum := 0
	for _, d := range dice {
		sum += d.Result
	}
	return sum
}

func sumRollables(rollables ...Rollable) (float64, error) {
	sum := 0.0
	for _, r := range rollables {
		i, err := r.(Rollable).Roll()
		if err != nil {
			return 0, err
		}
		sum += i
	}
	return sum, nil
}

// Sum returns and sets the total of a rolled dice set
func (d DieSet) Sum() float64 {
	d.Result = (float64)(sumDice(d.Dice))
	return d.Result
}

// Parse parses a dice notation string and returns a Dice set representation.
func Parse(notation string) (DieSet, error) {
	return parse(notation)
}

// parse is the real-deal notation parsing method.
func parse(notation string) (DieSet, error) {
	matches := DiceNotationRegex.FindStringSubmatch(notation)
	if len(matches) < 3 {
		return DieSet{}, &ErrParseError{notation, notation, "", ": failed to identify dice components"}
	}

	// Parse and cast dice properties from regex capture values
	count, err := strconv.ParseUint(matches[1], 10, 0)
	if err != nil {
		// either there was an implied count, ex 'd20', or count was invalid
		count = 1
	}
	size, err := strconv.ParseUint(matches[2], 10, 0)
	if err == nil {
		// valid size, so build the dice set
		return NewDieSet(int(size), uint(count)), nil
	}

	// Check for special dice types
	if matches[2] == "F" {
		return DieSet{}, errors.New("fudge dice not yet implemented")
	}

	// Couldn't parse the "size" as a uint and it's not a special die type
	return DieSet{}, &ErrParseError{notation, matches[2], "size", ": invalid size"}
}
