package command

import (
	"fmt"

	"github.com/travis-g/draas/dice/math"
	"github.com/urfave/cli"
)

func EvalCommand(c *cli.Context) error {
	eval := c.Args().Get(0)
	exp, err := math.Evaluate(eval)
	if err != nil {
		return err
	}
	out, err := toTable(toStruct(exp))
	fmt.Println(out)
	return nil
}
