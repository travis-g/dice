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
		d.X = float64(c)
	} else {
		d.X = 1
	}
	d.Y = components["size"]
	d.Modifiers = components["modifiers"]
	return nil
}
