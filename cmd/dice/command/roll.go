package command

import (
	"fmt"

	"github.com/travis-g/dice"
	"github.com/urfave/cli"
)

// RollCommand is a command that will create a Dice from the first argument
// passed and roll it, printing the result.
func RollCommand(c *cli.Context) error {
	roll := c.Args().Get(0)
	props, err := dice.ParseNotation(roll)
	if err != nil {
		return err
	}
	group, err := dice.NewGroup(props)
	_, err = group.Roll()
	if err != nil {
		return err
	}
	// get post-roll properties
	props = dice.Properties(&group)
	out, err := Output(c, &props)
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}
