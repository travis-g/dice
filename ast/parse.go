package main

import (
	"strconv"

	"github.com/travis-g/dice"
)

func (d *Dice) ParseNotation(notation string) error {
	components := dice.FindNamedCaptureGroups(dice.DiceWithModifiersExpressionRegex, notation)

	if components["count"] != "" {
		c, err := strconv.Atoi(components["count"])
		if err != nil {
			return err
		}
		d.Count = ptr(c)
	} else {
		d.Count = ptr(1)
	}
	if s, err := strconv.Atoi(components["size"]); err != nil {
		d.Size = ptr(s)
	} else if components["size"] == "F" {
		// TODO
		d.Size = ptr(1)
	}
	d.Modifiers = components["modifiers"]
	return nil
}
