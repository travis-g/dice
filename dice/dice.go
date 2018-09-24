package dice

import (
	"bytes"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
)

const (
	// diceNotationPattern is the RegEx pattern that matches a dice notation
	// string in the format XdY, where X is the number of Y-sided dice to roll.
	// X may be omitted if it is 1, yielding dY instead of 1dY.
	diceNotationPattern = `(?P<count>\d*)d(?P<size>\d{1,})`
)

var (
	diceNotationRegex = regexp.MustCompile(diceNotationPattern)
)

type Die struct {
	Size   int `json:"size"`
	Result int `json:"result"`
}

// NewDie creates and returns a rolled die of size [1, size]. It panics if size < 1.
func NewDie(size int) Die {
	if size < 1 {
		panic("dice: call to create a new Die with less than 1 side")
	}
	d := new(Die)
	d.setSize(size)
	d.roll()
	return *d
}

func (d *Die) setSize(size int) {
	if size < 1 {
		panic("dice: call to setSize with size < 1")
	}
	d.Size = size
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

// func (d Die) Result() int {
// 	return d.result
// }

// A Dice set is a group of like-sided dice from a dice notation string
type Dice struct {
	Size   int    `json:"size"`
	Count  int    `json:"count"`
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

func NewDice(size, count int) Dice {
	s := size
	c := count
	dice := make([]*Die, c)
	for i := range dice {
		die := NewDie(s)
		dice[i] = &die
	}
	total := sumDice(dice)
	return Dice{s, c, total, dice}
}

func (dice Dice) Stats() *Dice {
	return &dice
}

func sumDice(dice []*Die) int {
	sum := 0
	for _, d := range dice {
		sum += d.Result
	}
	return sum
}

// sum returns and sets the total of a rolled dice set
func (d *Dice) sum() int {
	sum := sumDice(d.Dice)
	d.Result = sum
	return sum
}

func (d *Dice) Reroll() int {
	return 0
}

type ParseError struct {
	Notation     string
	NotationElem string
	ValueElem    string
	Message      string
}

func quote(s string) string {
	return "\"" + s + "\""
}

func (e *ParseError) Error() string {
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
func Parse(notation string) (Dice, error) {
	return parse(notation)
}

func parse(notation string) (Dice, error) {
	anotation := notation
	var (
		size  = 1
		count = 1
		err   error
	)
	matches := diceNotationRegex.FindStringSubmatch(anotation)
	if len(matches) < 3 {
		return Dice{}, &ParseError{anotation, anotation, "", ": failed to identify dice components"}
	}
	scount, ssize := matches[1], matches[2]
	count, err = strconv.Atoi(scount)
	if err != nil {
		count = 1
	}
	size, err = strconv.Atoi(ssize)
	if err != nil {
		return Dice{}, &ParseError{anotation, ssize, "size", ": non-int size"}
	}

	return NewDice(size, count), nil
}

type DiceCalculation struct {
	original string
	result   int
	dice     []*Dice
}
