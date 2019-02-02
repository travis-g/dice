package dice

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	rand "gitlab.com/NebulousLabs/fastrand"
)

const (
	// DiceNotationPattern is the RegEx string pattern that matches a dice
	// notation string in the format XdY, where X is the number of Y-sided dice
	// to roll. X may be omitted if it is 1, yielding dY instead of 1dY.
	DiceNotationPattern = `(?P<count>\d*)d(?P<size>\d{1,})`

	// DropKeepNotationPattern is the RegEx string pattern that matches a
	// drop/keep-style dice roll.
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
	Roll()
	String() string
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
	d := new(Die)
	d.Size = size
	d.kind = fmt.Sprintf("d%d", d.Size)
	d.Roll()
	return d, nil
}

// Roll will Roll a given Die (if unrolled) and set the die's result. Results
// are in the range [1, size].
func (d *Die) Roll() {
	if !d.rolled {
		d.Result = 1 + rand.Intn(d.Size)
		d.rolled = true
	}
}

func (d Die) String() string {
	// Largest die is d2147483647
	return d.kind
}

type FateDie struct {
	Result int `json:"result"`
}

func (f FateDie) String() string {
	return "dF"
}

// Roll will Roll a given FateDie and set the die's result. Fate dice can have
// results in the range [-1, 1].
func (f FateDie) Roll() {
	f.Result = rand.Intn(2) - 1
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
		fmt.Fprintf(&s, "%d", l)
	}

	fmt.Fprintf(&s, "d%d", d.Size)
	return s.String()
}

func (d Dice) String() string {
	return d.Notation()
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
func (d *Dice) Roll() {
	for _, i := range d.Dice {
		i.Roll()
	}
	d.Sum()
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
	var (
		size  uint
		count uint
	)
	matches := DiceNotationRegex.FindStringSubmatch(notation)
	if len(matches) < 3 {
		return &Dice{}, &ErrParseError{notation, notation, "", ": failed to identify dice components"}
	}
	scount, ssize := matches[1], matches[2]

	// Parse and cast dice properties from regex capture values
	ucount, err := strconv.ParseUint(scount, 10, 0)
	count = uint(ucount)
	if err != nil {
		// either there was an implied count, ex 'd20', or count was invalid
		count = 1
	}
	usize, err := strconv.ParseUint(ssize, 10, 0)
	if err != nil {
		return &Dice{}, &ErrParseError{notation, ssize, "size", ": non-uint size"}
	}
	size = uint(usize)

	return NewDice(int(size), count), nil
}
