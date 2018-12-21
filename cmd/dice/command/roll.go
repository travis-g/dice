package command

import (
	"fmt"

	"github.com/travis-g/dice"
	"github.com/urfave/cli"
)

func RollCommand(c *cli.Context) error {
	roll := c.Args().Get(0)
	dice, err := dice.Parse(roll)
	if err != nil {
		return err
	}
	fmt.Println(dice.Result)
	return nil
}
