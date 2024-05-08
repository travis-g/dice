package main

import "github.com/travis-g/dice"

func (d *Dice) ParseNotation(notation string) error {
	components := dice.FindNamedCaptureGroups(dice.DiceWithModifiersExpressionRegex, notation)

	d.X = components["count"]
	d.Y = components["size"]
	d.Modifiers = components["modifiers"]
	return nil
}
