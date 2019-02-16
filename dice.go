package dice

import (
	"regexp"
	"strconv"
)

var (
	// DiceNotationRegex is the compiled RegEx for parsing supported dice
	// notations.
	DiceNotationRegex = regexp.MustCompile(`(?P<count>\d*)d(?P<size>(?:\d{1,}|F))`)

	// DropKeepNotationRegex is the compiled RegEx for parsing drop/keep dice
	// notations (unimplemented).
	DropKeepNotationRegex = regexp.MustCompile(`(?P<count>\d+)?d(?P<size>\d{1,})(?P<dropkeep>(?P<op>[dk][lh]?)(?P<num>\d{1,}))?`)
)

// Parse parses a dice notation string and returns a Dice set representation.
func Parse(notation string) (RollableSet, error) {
	return parse(notation)
}

// parse is the real-deal notation parsing method.
func parse(notation string) (RollableSet, error) {
	matches := DiceNotationRegex.FindStringSubmatch(notation)
	if len(matches) < 3 {
		return nil, &ErrParseError{notation, notation, "", ": failed to identify dice components"}
	}

	// Parse and cast dice properties from regex capture values
	count, err := strconv.ParseUint(matches[1], 10, 0)
	if err != nil {
		// either there was an implied count, ex 'd20', or count was invalid
		count = 1
	}
	size, err := strconv.ParseUint(matches[2], 10, 0)
	if err == nil {
		// valid size, so build the dice set
		set := NewDieSet(int(size), uint(count))
		return &set, nil
	}

	// Check for special dice types
	if matches[2] == "F" {
		set := NewFateDieSet(uint(count))
		return &set, nil
	}

	// Couldn't parse the "size" as a uint and it's not a special die type
	return &DieSet{}, &ErrParseError{notation, matches[2], "size", ": invalid size"}
}
