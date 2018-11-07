package dice

import (
	"bytes"
	"fmt" // rand "math/rand"
	"regexp"
	"strconv"

	rand "github.com/NebulousLabs/fastrand"
)

func init() {
	// // seed PRNG
	// rand.Seed(time.Now().UTC().UnixNano())
}

const (
	// DiceNotationPattern is the RegEx string pattern that matches a dice
	// notation string in the format XdY, where X is the number of Y-sided dice
	// to roll. X may be omitted if it is 1, yielding dY instead of 1dY.
	DiceNotationPattern = `(?P<count>\d*)d(?P<size>\d{1,})`
)

var (
	// DiceNotationRegex is the compiled RegEx for parsing simple dice
	// notations.
	DiceNotationRegex = regexp.MustCompile(DiceNotationPattern)
)

// A Die represents a variable-sided die in memory, including the result of
// rolling it.
type Die struct {
	Size   int `json:"size"`
	Result int `json:"result"`
}

// NewDie creates and returns a rolled die of size [1, size]. It panics if size
// < 1.
func NewDie(size int) Die {
	if size < 1 {
		panic("dice: call to create a new Die with less than 1 side")
	}
	d := new(Die)
	d.setSize(size)
	d.roll()
	return *d
}

func (d *Die) setSize(size int) error {
	if size < 1 {
		return fmt.Errorf("dice: call to setSize with size < 1")
	}
	d.Size = size
	return nil
}

func (d Die) String() string {
	// Largest die is d2147483647
	return fmt.Sprintf("d%d", d.Size)
}

// Roll will roll a given Die and set the die's result. Results are in the range
// [1, size].
func (d *Die) roll() int {
	r := 1 + rand.Intn(d.Size)
	d.Result = r
	return r
}

// A Dice set is a group of like-sided dice from a dice notation string
type Dice struct {
	Size   int    `json:"size"`
	Count  uint   `json:"count"`
	Result int    `json:"result"`
	Dice   []*Die `json:"dice"`
}

// String returns the dice notation format of the dice group in the format XdY,
// where X is the count of dice to roll and Y is the size of the dice
func (d Dice) Notation() string {
	var s bytes.Buffer

	if l := len(d.Dice); l > 1 {
		fmt.Fprintf(&s, "%d", l)
	}

	fmt.Fprintf(&s, "d%d", d.Size)
	return s.String()
}

func NewDice(size int, count uint) *Dice {
	s := size
	c := count
	dice := make([]*Die, c)
	for i := range dice {
		die := NewDie(s)
		dice[i] = &die
	}
	total := sumDice(dice)
	return &Dice{s, c, total, dice}
}

func sumDice(dice []*Die) int {
	sum := 0
	for _, d := range dice {
		sum += d.Result
	}
	return sum
}

// sum returns and sets the total of a rolled dice set
func (d *Dice) Sum() int {
	sum := sumDice(d.Dice)
	d.Result = sum
	return sum
}

func quote(s string) string {
	return "\"" + s + "\""
}

func (e *ErrParseError) Error() string {
	if e.Message == "" {
		return "parsing dice string " +
			quote(e.Notation) + ": cannot parse " +
			quote(e.ValueElem) + " as " +
			quote(e.NotationElem)
	}
	return "parsing dice " +
		quote(e.Notation) + e.Message
}

// Parse parses a dice notation string and returns a Dice set representation.
func Parse(notation string) (*Dice, error) {
	return parse(notation)
}

// Parse sets a dice set's properties, given a notation. This recreates the dice.
func (d *Dice) Parse(notation string) error {
	d, err := parse(notation)
	if err != nil {
		return err
	}
	return nil
}

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
