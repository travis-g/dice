package command

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/travis-g/dice/math"
	"github.com/urfave/cli"
)

const replPrompt = ">>> "

// REPLCommand is a command that will initiate a dice REPL.
func REPLCommand(c *cli.Context) error {
	ctx := context.Background()
	scanner := bufio.NewScanner(os.Stdin)

	// Check if data was piped through Stdin, or if the REPL is interactive
	in, _ := os.Stdin.Stat()
	interactive := ((in.Mode() & os.ModeCharDevice) != 0)

	// Begin the REPL:
	for {
		if interactive {
			fmt.Fprintf(os.Stderr, replPrompt)
		}
		scanned := scanner.Scan()
		if !scanned {
			return nil
		}

		line := scanner.Text()
		if line != "quit" {
			exp, err := math.EvaluateExpression(ctx, line)
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
