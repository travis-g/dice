package dice

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Regexes for parsing basic dice notation strings.
var (
	// DiceNotationPattern is the base XdY notation pattern for matching dice
	// strings.
	DiceNotationPattern = `(?P<count>\d+)?d(?P<size>\d{1,}|F)`

	// DiceNotationRegex is the compiled RegEx for parsing supported dice
	// notations.
	DiceNotationRegex = regexp.MustCompile(DiceNotationPattern)

	// DiceExpressionRegex is the compiled RegEx for parsing drop/keep dice
	// notations and other expressions that would modify a dice group's result.
	DiceExpressionRegex = regexp.MustCompile(
		`((?P<group>\{.*\})|` +
			DiceNotationPattern +
			`)(?P<dropkeep>(?P<op>[dk][lh]?)(?P<num>\d{1,}))?`)

	// ComparePointpattern is the base pattern that matches compare points
	// within dice modifiers.
	ComparePointPattern = `(?P<compare>[=<>])?(?P<point>\d+)`

	// ComparePointRegex is the compiled RegEx for parsing supported dice
	// modifiers' core compare points.
	ComparePointRegex = regexp.MustCompile(ComparePointPattern)
)

// DiceWithModifiersExpressionRegex is the compiled RegEx for parsing a dice
// notation with modifier strings appended.
var DiceWithModifiersExpressionRegex = regexp.MustCompile(
	DiceNotationPattern + `(?P<modifiers>[^-+ \(\)]*)`)
var (
	rerollRegex   = regexp.MustCompile(`r(?P<once>o)?` + ComparePointPattern + `?`)
	sortRegex     = regexp.MustCompile(`s(?P<sort>[ad]?)`)
	dropKeepRegex = regexp.MustCompile(`(?P<op>[dk][lh]?)(?P<num>\d+)`)
)

// Prefixes that indicate a modifier start in a string
const (
	rerollPrefix = "r"
	sortPrefix   = "s"
	dropPrefix   = "d"
	keepPrefix   = "k"
)

// ParseNotation parses the provided notation with updated regular
// expressions that also extract dice group modifiers.
func ParseNotation(ctx context.Context, notation string) (RollerProperties, error) {
	props := RollerProperties{
		DieModifiers:   ModifierList{},
		GroupModifiers: ModifierList{},
	}

	components := FindNamedCaptureGroups(DiceWithModifiersExpressionRegex, notation)

	// Parse and cast dice properties from regex capture values
	count64, err := strconv.ParseInt(components["count"], 10, 0)
	count := int(count64)
	if err != nil {
		// either there was an implied count, ex 'd20', or count was invalid
		count = 1
	}
	props.Count = count

	var size64 int64
	if components["size"] == "F" {
		props.Type = TypeFudge
		props.Size = 1
	} else if size64, err = strconv.ParseInt(components["size"], 10, 0); err != nil {
		return props, &ErrParseError{notation, components["size"], "size", ": invalid size"}
	} else if size64 <= 0 {
		return props, ErrSizeZero
	}
	props.Size = int(size64)

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
		switch {
		// rerolls
		case strings.HasPrefix(modifiers, rerollPrefix):
			remainingBytes := rerollRegex.ReplaceAllFunc([]byte(modifiers), func(matchBytes []byte) []byte {
				captures := FindNamedCaptureGroups(rerollRegex, string(matchBytes))

				point, _ := strconv.Atoi(captures["point"])
				// TODO: fix Once after recursion issue is solved
				// once := captures["once"] == "o"
				props.DieModifiers = append(props.DieModifiers, &RerollModifier{
					CompareTarget: &CompareTarget{
						Compare: LookupCompareOp(captures["compare"]),
						Target:  point,
					},
					// Once: once,
				})
				return []byte{}
			})
			modifiers = string(remainingBytes)

		// sort
		case strings.HasPrefix(modifiers, sortPrefix):
			remainingBytes := sortRegex.ReplaceAllFunc([]byte(modifiers), func(matchBytes []byte) []byte {
				// TODO
				// captures := getNamedCaptures(sortRegex, string(matchBytes))

				return []byte(nil)
			})
			modifiers = string(remainingBytes)

		// drop/keep
		case strings.HasPrefix(modifiers, dropPrefix), strings.HasPrefix(modifiers, keepPrefix):
			remainingBytes := dropKeepRegex.ReplaceAllFunc([]byte(modifiers), func(matchBytes []byte) []byte {
				captures := FindNamedCaptureGroups(dropKeepRegex, string(matchBytes))

				if captures["num"] == "" {
					return []byte(nil)
				}
				num, _ := strconv.Atoi(captures["num"])
				props.GroupModifiers = append(props.GroupModifiers, &DropKeepModifier{
					Method: DropKeepMethod(captures["op"]),
					Num:    num,
				})
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
	return props, nil
}

// ParseExpression parses a given expression into a properties
// object with support for modifiers.
func ParseExpression(ctx context.Context, expression string) (RollerProperties, error) {
	components := FindNamedCaptureGroups(DiceWithModifiersExpressionRegex, expression)

	// if group is found the core notation was not specified.
	group := components["group"]
	if group != "" {
		return RollerProperties{}, &ErrNotImplemented{"arbitrary group rolls not implemented"}
	}

	// Call ParseNotation with the core dice count and size.
	props, err := ParseNotation(ctx, expression)
	if err != nil {
		return RollerProperties{}, err
	}

	return props, nil

}
