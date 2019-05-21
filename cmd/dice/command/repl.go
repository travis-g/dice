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

	fmt.Println("Use quit() or Ctrl-C to exit")

	// Begin the REPL:
	for {
		fmt.Fprintf(os.Stderr, replPrompt)
		scanned := scanner.Scan()
		if !scanned {
			return nil
		}

		line := scanner.Text()
		if line != "quit()" {
			exp, err := math.Evaluate(ctx, line)
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
