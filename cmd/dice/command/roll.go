package command

import (
	"context"
	"fmt"

	"github.com/travis-g/dice"
	"github.com/urfave/cli"
)

// RollCommand is a command that will create a Dice from the first argument
// passed and roll it, printing the result.
func RollCommand(c *cli.Context) error {
	ctx := context.Background()

	roll := c.Args().Get(0)
	props, err := dice.ParseNotation(ctx, roll)
	if err != nil {
		return err
	}
	group, _ := dice.NewRollerGroup(&props)
	err = group.FullRoll(ctx)
	if err != nil {
		return err
	}
	out, err := Output(c, group)
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}
