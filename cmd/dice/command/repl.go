package command

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/travis-g/dice"
	"github.com/travis-g/dice/math"
	"github.com/urfave/cli"
)

const replPrompt = ">>> "

// REPLCommand is a command that will initiate a dice REPL.
func REPLCommand(c *cli.Context) error {
	scanner := bufio.NewScanner(os.Stdin)

	// Check if data was piped through Stdin, or if the REPL is interactive
	in, _ := os.Stdin.Stat()
	interactive := ((in.Mode() & os.ModeCharDevice) != 0)

	// Begin the REPL:
	for {
		// context for each interation
		ctx := dice.NewContextFromContext(context.Background())
		if interactive {
			fmt.Fprint(os.Stderr, replPrompt)
		}
		scanned := scanner.Scan()
		if !scanned {
			return nil
		}

		line := scanner.Text()
		if line != "quit" {
			ctx, cancel := context.WithTimeout(ctx, time.Second*5)
			exp, err := math.EvaluateExpression(ctx, line)
			cancel()
			if exp == nil {
				err = math.ErrNilExpression
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			out, err := Output(c, exp)
			if err != nil {
				return err
			}
			fmt.Println(out)
		} else {
			return nil
		}
	}
}
