package dice

import (
	"bytes"
	"fmt"
	rand "math/rand"
	"regexp"
	"strconv"
	"strings"
)

const (
	// DiceNotationPattern is the RegEx string pattern that matches a dice
	// notation string in the format XdY, where X is the number of Y-sided dice
	// to roll. X may be omitted if it is 1, yielding dY instead of 1dY.
	DiceNotationPattern = `(?P<count>\d*)d(?P<size>\d{1,})`

	// DropKeepNotationPattern is the RegEx string pattern that matches a
	// drop/keep-style dice roll (unimplemented).
	DropKeepNotationPattern = `(?P<count>\d+)?d(?P<size>\d{1,})(?P<dropkeep>(?P<op>[dk][lh]?)(?P<num>\d{1,}))?`
)

var (
	// DiceNotationRegex is the compiled RegEx for parsing simple dice
	// notations.
	DiceNotationRegex = regexp.MustCompile(DiceNotationPattern)

	// DropKeepNotationRegex is the compiled RegEx for parsing drop/keep dice
	// notations.
	DropKeepNotationRegex = regexp.MustCompile(DropKeepNotationPattern)
)

// A Rollable is any kind of rollable object. A Rollable could be a single die
// or many dice of any type.
type Rollable interface {
	// Roll should be used to also set the object's Result
	Roll() float64
	String() string
	Type() string
}

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
func (d *Die) Roll() float64 {
	if !d.rolled {
		d.Result = 1 + rand.Intn(d.Size)
		d.rolled = true
	}
	return float64(d.Result)
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

func (f *FateDie) String() string {
	return string(f.Result)
}

// Type returns the FateDie's type
func (f FateDie) Type() string {
	return "dF"
}

// NewFateDie create and returns a new FateDie. The error will always be nil.
func NewFateDie() (*FateDie, error) {
	f := new(FateDie)
	f.Roll()
	return f, nil
}

// Roll will Roll a given FateDie and set the die's result. Fate dice can have
// results in [-1, 1].
func (f *FateDie) Roll() float64 {
	if !f.rolled {
		f.Result = rand.Intn(3) - 2
		f.rolled = true
	}
	return float64(f.Result)
}

// A Dice set is a group of like-sided dice from a dice notation string
type Dice struct {
	Count    uint   `json:"count"`
	Dice     []*Die `json:"dice,omitempty"`
	Drop     int    `json:"drop,omitempty"`
	Expanded string `json:"expanded"`
	Result   int    `json:"result"`
	Size     int    `json:"size"`
}

// Notation returns the dice notation format of the dice group in the format
// XdY, where X is the count of dice to roll and Y is the size of the dice
func (d Dice) Notation() string {
	var s bytes.Buffer

	if l := len(d.Dice); l > 1 {
		s.WriteString(strconv.Itoa(l))
	}
	s.WriteString(strings.Join([]string{"d", strconv.Itoa(d.Size)}, ""))

	return s.String()
}

func (d *Dice) String() string {
	return strings.Join([]string{d.Expanded, "=>", strconv.Itoa(d.Result)}, " ")
}

// Type returns the Dice type
func (d Dice) Type() string {
	return strings.Join([]string{"d", strconv.Itoa(d.Size)}, "")
}

// NewDice creates a new Dice object and returns its pointer
func NewDice(size int, count uint) *Dice {
	dice := make([]*Die, count)
	results := make([]int, count)
	for i := range dice {
		die, err := NewDie(size)
		if err != nil {
			continue
		}
		dice[i] = die
		results[i] = die.Result
	}
	expr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(results)), "+"), "[]")
	return &Dice{
		Count:    count,
		Dice:     dice,
		Expanded: expr,
		Result:   sumDice(dice),
		Size:     size,
	}
}

// Roll rolls the dice within a Dice set and sums the result with `Sum()`
func (d *Dice) Roll() float64 {
	for _, i := range d.Dice {
		i.Roll()
	}
	return float64(d.Sum())
}

func sumDice(dice []*Die) int {
	sum := 0
	for _, d := range dice {
		sum += d.Result
	}
	return sum
}

// Sum returns and sets the total of a rolled dice set
func (d *Dice) Sum() int {
	sum := sumDice(d.Dice)
	d.Result = sum
	return sum
}

// Parse parses a dice notation string and returns a Dice set representation.
func Parse(notation string) (*Dice, error) {
	return parse(notation)
}

// Parse sets a dice set's properties, given a notation. If properties of the
// dice set have already been set this recreates the dice based on the given
// notation.
func (d *Dice) Parse(notation string) error {
	d, err := parse(notation)
	if err != nil {
		return err
	}
	return nil
}

// parse is the real-deal notation parsing method.
func parse(notation string) (*Dice, error) {
	matches := DiceNotationRegex.FindStringSubmatch(notation)
	if len(matches) < 3 {
		return &Dice{}, &ErrParseError{notation, notation, "", ": failed to identify dice components"}
	}

	// Parse and cast dice properties from regex capture values
	count, err := strconv.ParseUint(matches[1], 10, 0)
	if err != nil {
		// either there was an implied count, ex 'd20', or count was invalid
		count = 1
	}
	size, err := strconv.ParseUint(matches[2], 10, 0)
	if err != nil {
		return &Dice{}, &ErrParseError{notation, matches[2], "size", ": non-uint size"}
	}

	return NewDice(int(size), uint(count)), nil
}
