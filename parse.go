package dice

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// DiceNotationPattern is the base XdY notation pattern for matching dice
	// strings.
	DiceNotationPattern = `(?P<count>\d+)?d(?P<size>\d{1,}|F)`
	// DiceNotationRegex is the compiled RegEx for parsing supported dice
	// notations.
	DiceNotationRegex = regexp.MustCompile(DiceNotationPattern)
	// DiceExpressionRegex is the compiled RegEx for parsing drop/keep dice
	// notations and other expressions that modify the dice group.
	DiceExpressionRegex = regexp.MustCompile(
		`((?P<group>\{.*\})|` +
			DiceNotationPattern +
			`)(?P<dropkeep>(?P<op>[dk][lh]?)(?P<num>\d{1,}))?`)
)

// ParseNotation parses a notation into a group of dice. It returns the group
// unrolled.
func ParseNotation(notation string) (GroupProperties, error) {
	matches := DiceNotationRegex.FindStringSubmatch(notation)
	if len(matches) < 3 {
		return GroupProperties{}, &ErrParseError{notation, notation, "", ": failed to identify dice components"}
	}

	// extract named capture groups to map
	components := make(map[string]string)
	for i, name := range DiceNotationRegex.SubexpNames() {
		if i != 0 && name != "" {
			components[name] = matches[i]
		}
	}

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
	if components["size"] == "F" {
		props := GroupProperties{
			Type:     TypeFate,
			Count:    int(count),
			Unrolled: true,
		}
		return props, nil
	}

	return GroupProperties{}, &ErrParseError{notation, components["size"], "size", ": invalid size"}
}

// ParseExpression parses a notation based on the DiceExpressionRegex, allowing
// for drop/keep sets, reroll expressions, exploding dice, etc.
func ParseExpression(notation string) (GroupProperties, error) {
	matches := DiceExpressionRegex.FindStringSubmatch(notation)
	if len(matches) < 3 {
		return GroupProperties{}, &ErrParseError{notation, notation, "", ": failed to identify dice components"}
	}

	// extract named capture groups to map
	components := make(map[string]string)
	for i, name := range DiceExpressionRegex.SubexpNames() {
		if i != 0 && name != "" {
			components[name] = matches[i]
		}
	}

	// if group is found the core notation was not specified.
	group := components["group"]
	if group != "" {
		return GroupProperties{}, &ErrNotImplemented{"arbitrary group rolls not implemented"}
	}

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
		dropkeep = -props.Count - int(num)
	}

	if dropkeep != 0 {
		props.DropKeep = dropkeep
	}

	return props, nil
}
