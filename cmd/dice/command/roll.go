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
	dice, err := dice.Parse(roll)
	if err != nil {
		return err
	}
	fmt.Println(dice.Result)
	return nil
}
