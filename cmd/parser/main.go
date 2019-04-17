package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
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
		`(?P<reroll>r[o]?)` + comparePointPattern + `?`)

	sortRegex     = regexp.MustCompile(`(?P<sort>s[ad]?)`)
	dropKeepRegex = regexp.MustCompile(`(?P<op>[dk][lh]?)(?P<num>\d+)`)
)

// A RollFunc is a function called immediately after a die is rolled.
type RollFunc func(context.Context) error

// A CallbackFunc is a function executed on the full roll once all dice settle.
type CallbackFunc func(context.Context) error

func main() {
	// callbacks to roll on each die roll
	rollfuncs := []string(nil)
	// callbacks to execute on the dice group after all dice settle
	callbacks := []string(nil)

	matches := diceExpressionRegex.FindStringSubmatch(os.Args[1])

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
	modifiers := components["modifiers"]
	for modifiers != "" {
		fmt.Println(modifiers)
		switch {
		// rerolls
		case strings.HasPrefix(modifiers, "r"):
			modifierBytes := rerollRegex.ReplaceAllFunc([]byte(modifiers), func(matchBytes []byte) []byte {
				modifierMatches := rerollRegex.FindStringSubmatch(string(matchBytes))

				// extract named capture groups to map
				props := make(map[string]string)
				for i, name := range rerollRegex.SubexpNames() {
					if i != 0 && name != "" {
						props[name] = modifierMatches[i]
					}
				}
				rollfuncs = append(rollfuncs, fmt.Sprintf("%v", props))
				// remove the reroll modifier from the string
				return []byte(nil)
			})
			modifiers = string(modifierBytes)
		// sort
		case strings.HasPrefix(modifiers, "s"):
			modifierBytes := sortRegex.ReplaceAllFunc([]byte(modifiers), func(matchBytes []byte) []byte {
				modifierMatches := sortRegex.FindStringSubmatch(string(matchBytes))

				// extract named capture groups to map
				props := make(map[string]string)
				for i, name := range sortRegex.SubexpNames() {
					if i != 0 && name != "" {
						props[name] = modifierMatches[i]
					}
				}
				callbacks = append(callbacks, fmt.Sprintf("%v", props))
				// remove the reroll modifier from the string
				return []byte(nil)
			})
			modifiers = string(modifierBytes)
		// drop/keep
		case strings.HasPrefix(modifiers, "d"), strings.HasPrefix(modifiers, "k"):
			modifierBytes := dropKeepRegex.ReplaceAllFunc([]byte(modifiers), func(matchBytes []byte) []byte {
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
				callbacks = append(callbacks, fmt.Sprintf("%v", props))
				// remove the reroll modifier from the string
				return []byte(nil)
			})
			if modifiers == string(modifierBytes) {
				fmt.Printf("invalid drop/keep: %s\n", modifiers)
				modifiers = ""
				break
			}
			modifiers = string(modifierBytes)
		default:
			fmt.Printf("invalid modifiers: %s\n", modifiers)
			modifiers = ""
			break
		}
	}
	fmt.Printf("%#v\n%#v\n", callbacks, rollfuncs)
}
