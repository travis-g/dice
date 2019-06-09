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
	rerollRegex   = regexp.MustCompile(`r` + ComparePointPattern + `?`)
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

// ParseNotationWithModifier parses the provided notation with updated regular
// expressions that also extract dice group modifiers.
func ParseNotationWithModifier(ctx context.Context, notation string) (RollerProperties, error) {
	props := RollerProperties{
		DieModifiers:   ModifierList{},
		GroupModifiers: ModifierList{},
	}

	components := getNamedCaptures(DiceWithModifiersExpressionRegex, notation)

	// Parse and cast dice properties from regex capture values
	count64, err := strconv.ParseInt(components["count"], 10, 0)
	count := int(count64)
	if err != nil {
		// either there was an implied count, ex 'd20', or count was invalid
		count = 1
	}
	props.Count = count

	if components["size"] == "F" {
		props.Type = TypeFudge
		props.Size = 1
	} else if size, err := strconv.ParseUint(components["size"], 10, 0); err != nil {
		return props, &ErrParseError{notation, components["size"], "size", ": invalid size"}
	} else {
		props.Size = uint(size)
	}

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
				captures := getNamedCaptures(rerollRegex, string(matchBytes))

				point, _ := strconv.Atoi(captures["point"])
				props.DieModifiers = append(props.DieModifiers, &RerollModifier{
					CompareTarget: CompareTarget{
						Compare: LookupCompareOp(captures["compare"]),
						Target:  point,
					},
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
				captures := getNamedCaptures(dropKeepRegex, string(matchBytes))

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

// ParseNotation parses a notation into a group of dice. It returns the group
// unrolled.
func ParseNotation(notation string) (GroupProperties, error) {
	components := getNamedCaptures(DiceNotationRegex, notation)

	// Parse and cast dice properties from regex capture values
	count, err := strconv.ParseInt(components["count"], 10, 0)
	if err != nil {
		// either there was an implied count, ex 'd20', or count was invalid
		count = 1
	}

	size, err := strconv.ParseUint(components["size"], 10, 0)
	// err is nil, which means a valid uint size
	if err == nil {
		props := GroupProperties{
			Type:     TypePolyhedron,
			Size:     int(size),
			Count:    int(count),
			Unrolled: true,
		}
		return props, nil
	}

	// size was not a uint, check for special dice types
	switch s := components["size"]; s {
	case "F":
		props := GroupProperties{
			Type:     TypeFudge,
			Size:     1,
			Count:    int(count),
			Unrolled: true,
		}
		return props, nil
	default:
		return GroupProperties{}, &ErrParseError{notation, s, "size", ": invalid size"}
	}
}

// ParseExpression parses a notation based on the DiceExpressionRegex, allowing
// for drop/keep sets, reroll expressions, exploding dice, etc.
func ParseExpression(notation string) (GroupProperties, error) {
	components := getNamedCaptures(DiceExpressionRegex, notation)

	// if group is found the core notation was not specified.
	group := components["group"]
	if group != "" {
		return GroupProperties{}, &ErrNotImplemented{"arbitrary group rolls not implemented"}
	}

	// Call ParseNotation with the core dice count and size.
	props, err := ParseNotation(strings.Join([]string{components["count"], components["size"]}, "d"))
	if err != nil {
		return props, err
	}

	var dropkeep int
	op := components["op"]
	num, _ := strconv.ParseInt(components["num"], 10, 0)
	switch op {
	case "d", "dl":
		dropkeep = int(num)
	case "k", "kh":
		dropkeep = props.Count - int(num)
	case "dh":
		dropkeep = -int(num)
	case "kl":
		dropkeep = int(num) - props.Count
	}

	if dropkeep != 0 {
		props.DropKeep = dropkeep
	}

	return props, nil
}

// ParseExpressionWithModifiers parses a given expression into a properties
// object with support for modifiers.
func ParseExpressionWithModifiers(ctx context.Context, expression string) (RollerProperties, error) {
	components := getNamedCaptures(DiceWithModifiersExpressionRegex, expression)

	// if group is found the core notation was not specified.
	group := components["group"]
	if group != "" {
		return RollerProperties{}, &ErrNotImplemented{"arbitrary group rolls not implemented"}
	}

	// Call ParseNotation with the core dice count and size.
	props, err := ParseNotationWithModifier(context.TODO(), expression)
	if err != nil {
		return RollerProperties{}, err
	}

	return props, nil

}
