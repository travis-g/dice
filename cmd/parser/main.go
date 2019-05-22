/*
Package parser is an experimental expression parser to enable flexible callbacks
on dice Groups.
*/
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/travis-g/dice"
)

var (
	diceNotationPattern = `(?P<count>\d+)?d(?P<size>\d{1,}|F)`
	diceExpressionRegex = regexp.MustCompile(
		diceNotationPattern + `(?P<modifiers>[^-+ \(\)]*)`)
	comparePointPattern = `(?P<compare>[=<>])?(?P<point>\d+)`
	comparePointRegex   = regexp.MustCompile(comparePointPattern)
	compoundingRegex    = regexp.MustCompile(`!!` + comparePointPattern + `?`)
	penetratingRegex    = regexp.MustCompile(`!p` + comparePointPattern + `?`)
	explodingRegex      = regexp.MustCompile(`!` + comparePointPattern + `?`)
	rerollRegex         = regexp.MustCompile(
		`r(?P<once>o?)` + comparePointPattern + `?`)

	sortRegex     = regexp.MustCompile(`(?P<sort>s[ad]?)`)
	dropKeepRegex = regexp.MustCompile(`(?P<op>[dk][lh]?)(?P<num>\d+)`)
)

// A DiceFunc is a function called immediately after a die is rolled.
type DiceFunc func(ctx context.Context, die dice.Interface) error

func SortFunc(ctx context.Context, args map[string]string) DiceFunc {
	return func(ctx context.Context, die dice.Interface) error {
		return errors.New("not implemented")
	}
}

// Prefixes that indicate a modifier start in a string
const (
	rerollPrefix = "r"
	sortPrefix   = "s"
	dropPrefix   = "d"
	keepPrefix   = "k"
)

func main() {
	var ctx = context.Background()

	// callbacks to roll on each die roll
	rollfuncs := []string(nil)
	// callbacks to execute on the dice group after all dice settle
	postfuncs := []string(nil)

	matches := diceExpressionRegex.FindStringSubmatch(os.Args[1])

	test, _ := dice.ParseExpression(os.Args[1])
	fmt.Println(test)

	// extract named capture groups to map
	components := make(map[string]string)
	for i, name := range diceExpressionRegex.SubexpNames() {
		if i != 0 && name != "" {
			components[name] = matches[i]
		}
	}
	fmt.Printf("%#v\n", components)

	// continuously loop through modifier string until we can't discern any more
	// types. Once a modifier type is seen all modifiers of that type are
	// searched out and added to the function sets in the order they appear in
	// the string.
	//
	// There are circumstances where we have to discern potentially ambiguous
	// modifier sets, like "2d6sdh" (sort, drop highest or sort descending?), so
	// the parsing should be left-to-right, like order of operations, and greedy
	modifiers := components["modifiers"]
	for modifiers != "" {
		fmt.Println(modifiers)
		switch {
		// rerolls
		case strings.HasPrefix(modifiers, rerollPrefix):
			remainingBytes := rerollRegex.ReplaceAllFunc([]byte(modifiers), func(matchBytes []byte) []byte {
				modifierMatches := rerollRegex.FindStringSubmatch(string(matchBytes))

				// extract named capture groups to map
				props := make(map[string]string)
				for i, name := range rerollRegex.SubexpNames() {
					if i != 0 && name != "" {
						props[name] = modifierMatches[i]
					}
				}
				rollfuncs = append(rollfuncs, fmt.Sprintf("%v", props))
				point, _ := strconv.Atoi(props["point"])
				test.Modifiers = append(test.Modifiers, &dice.RerollModifier{
					Compare: props["compare"],
					Point:   point,
				})
				// remove the reroll modifier from the string
				return []byte(nil)
			})
			modifiers = string(remainingBytes)

		// sort
		case strings.HasPrefix(modifiers, sortPrefix):
			remainingBytes := sortRegex.ReplaceAllFunc([]byte(modifiers), func(matchBytes []byte) []byte {
				modifierMatches := sortRegex.FindStringSubmatch(string(matchBytes))

				// extract named capture groups to map
				props := make(map[string]string)
				for i, name := range sortRegex.SubexpNames() {
					if i != 0 && name != "" {
						props[name] = modifierMatches[i]
					}
				}
				postfuncs = append(postfuncs, fmt.Sprintf("%v", props))
				// remove the reroll modifier from the string
				return []byte(nil)
			})
			modifiers = string(remainingBytes)

		// drop/keep
		case strings.HasPrefix(modifiers, dropPrefix), strings.HasPrefix(modifiers, keepPrefix):
			remainingBytes := dropKeepRegex.ReplaceAllFunc([]byte(modifiers), func(matchBytes []byte) []byte {
				modifierMatches := dropKeepRegex.FindStringSubmatch(string(matchBytes))

				fmt.Println(modifierMatches)
				// extract named capture groups to map
				props := make(map[string]string)
				for i, name := range dropKeepRegex.SubexpNames() {
					if i != 0 && name != "" {
						props[name] = modifierMatches[i]
					}
				}
				fmt.Println(props)
				if props["num"] == "" {
					return []byte(nil)
				}
				postfuncs = append(postfuncs, fmt.Sprintf("%v", props))
				// remove the reroll modifier from the string
				return []byte(nil)
			})
			if modifiers == string(remainingBytes) {
				fmt.Printf("invalid drop/keep: %s\n", modifiers)
				modifiers = ""
				break
			}
			modifiers = string(remainingBytes)
		default:
			fmt.Printf("invalid modifiers: %s\n", modifiers)
			modifiers = ""
			break
		}
	}
	fmt.Printf("roll funcs: %#v\n", rollfuncs)
	fmt.Printf("post funcs: %#v\n", postfuncs)
	fmt.Printf("props: %#v\n", test)

	die := dice.NewDie(test.Size)
	die.Modifiers = test.Modifiers
	fmt.Println(die)
	die.Roll(ctx)
	fmt.Println(die)
}
