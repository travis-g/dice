package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"

	"github.com/Knetic/govaluate"
	"github.com/rs/zerolog/log"
)

type RollExpression struct {
	Expression govaluate.EvaluableExpression `json:"expression"`
	Result     float64                       `json:"result"`
	Dice       []*Dice                       `json:"dice"`
}

type Result struct {
	roll     string      `json:"roll"`
	rolls    interface{} `json:"rolls"`
	Expanded string      `json:"expanded"`
	Result   interface{} `json:"result"`
	total    int         `json:"total"`
}

// Dice is a group of several Die objects that should be added together.
type Dice struct {
	Size  int    `json:"size"`
	Dice  []*Die `json:"rolls"`
	Count int    `json:"count"`
	// Result *Result `json:"result"`
}

type Die struct {
	Size   int `json:"size"`
	result int
}

// roll rolls the die and returns a die face value in the range [1, r.Size].
// Any changes to RNG, ex. to roll using a CSPRNG or pseudo-CSPRNG, should
// probably be made in this function.
func (r *Die) roll() int {
	return 1 + rand.Intn(r.Size)
}

// NewDie creates a new Die.
func NewDie(size int) *Die {
	d := Die{
		Size: size,
	}
	return &d
}

// Roll rolls a given Die.
func Roll(d *Die) {
	d.Roll()
}

// Type returns the type of die in dice notation (ex. "d20")
func (r *Die) Type() string {
	return "d" + strconv.Itoa(r.Size)
}

// Result returns the result of rolling a die.
func (r *Die) Result() int {
	return r.Roll()
}

// Roll returns the roll of a Die. If the die has already been rolled it will
// not be rerolled.
func (r *Die) Roll() int {
	// return Die's roll if previously rolled
	if r.result != 0 {
		return r.result
	}
	// Invalid die size returns 0
	if r.Size < 1 {
		r.result = 0
		return r.result
	}
	// Roll the die
	return r.roll()
}

func (r *Dice) String() string {
	return strconv.Itoa(r.Count) + "d" + strconv.Itoa(r.Size)
}

// Roll makes a single roll of a die in a set of Dice
// HACK replace this with something that rolls the full set of dice
func (r *Dice) roll() int {
	// Invalid die size returns 0
	if r.Size < 1 {
		return 0
	}
	return 1 + rand.Intn(r.Size)
}

func (d *Dice) RollAll() int {
	total := 0
	for _, die := range d.Dice {
		die.Roll()
		total += die.Result()
	}
	return total
}

func NewDiceSet(count, size int) *Dice {
	d := Dice{
		Size:  size,
		Count: count,
		Dice:  []*Die{},
	}
	for i := 0; i < d.Count; i++ {
		fmt.Println(i)
		d.Dice = append(d.Dice, NewDie(d.Size))
	}
	return &d
}

func (d *Dice) Cast() *Result {
	total := 0
	rolls := []int{}
	if d.Size < 1 {
		// Bad die => bad roll
		return &Result{
			Result: total, // 0
		}
	}
	// Roll & record:
	for i := 0; i < d.Count; i++ {
		result := d.roll()
		rolls = append(rolls, result)
		total += result
	}
	expanded := arrayToString(rolls, "+")
	log.Debug().
		Str("roll", d.String()).
		Int("result", total).
		Str("expanded", expanded).
		Msg("rolled")
	return &Result{
		Expanded: expanded,
		Result:   total,
	}
}

func ReplaceAllStringSubmatchFunc(re *regexp.Regexp, str string, repl func([]string) string) string {
	result := ""
	lastIndex := 0

	for _, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			groups = append(groups, str[v[i]:v[i+1]])
		}

		result += str[lastIndex:v[0]] + repl(groups)
		lastIndex = v[1]
	}

	return result + str[lastIndex:]
}

func NewDice(s string) *Dice {
	res := DiceRegex.FindStringSubmatch(s)
	if len(res) < 3 {
		// Return bunk 0 for non-rolls
		return &Dice{0, []*Die{}, 0}
	}
	count, err := strconv.Atoi(res[1])
	if err != nil {
		count = 1
	}
	size, err := strconv.Atoi(res[2])
	if err != nil {
		size = 1
	}
	return &Dice{
		Count: count,
		Dice:  []*Die{},
		Size:  size,
	}
}

func EvalDiceNotationString(s string) (*Result, error) {
	rolled := ReplaceAllStringSubmatchFunc(DiceRegex, s, func(match []string) string {
		die := NewDice(match[0])
		cast := die.Cast()
		return fmt.Sprintf("(%s)", cast.Expanded)
	})

	// Parse the roll result expression into an AST
	// exp, err := parser.ParseExpr(rolled)
	// if err != nil {
	// 	fmt.Printf("parsing failed: %s\n", err)
	// }

	expression, err := govaluate.NewEvaluableExpression(rolled)
	if err != nil {
		log.Error().
			Err(err).
			Str("expression", s).
			Str("rolled", rolled).
			Msg("error creating expression")
		return &Result{
			Expanded: rolled,
			Result:   0,
		}, err
	}
	result, err := expression.Evaluate(nil)
	if err != nil {
		log.Error().
			Err(err).
			Str("expression", rolled).
			Msg("error evaluating expression")
		return &Result{
			Result:   0,
			Expanded: rolled,
		}, err
	}

	return &Result{
		roll:     s,
		Expanded: rolled,
		Result:   result,
	}, nil
}
