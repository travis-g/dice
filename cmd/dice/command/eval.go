package command

import (
	"context"
	"fmt"

	"github.com/travis-g/dice"
	"github.com/travis-g/dice/math"
	"github.com/urfave/cli"
)

// EvalCommand will evaluate the first argument it is provided as a
// math.DiceExpression and print the result or return any errors during
// evaluation.
func EvalCommand(c *cli.Context) error {
	ctx := dice.NewContextFromContext(context.Background())

	eval := c.Args().Get(0)
	exp, err := math.EvaluateExpression(ctx, eval)
	if err != nil {
		return err
	}
	out, err := Output(c, exp)
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}
