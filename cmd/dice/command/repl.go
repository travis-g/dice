package command

import (
	"bufio"
	"fmt"
	"os"

	"github.com/travis-g/dice/math"
	"github.com/urfave/cli"
)

const replPrompt = ">>> "

// REPLCommand is a command that will initiate a dice REPL.
func REPLCommand(c *cli.Context) error {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Use quit() or Ctrl-C to exit")

	// Begin the REPL:
	for {
		fmt.Printf(replPrompt)
		scanned := scanner.Scan()
		if !scanned {
			return nil
		}

		line := scanner.Text()
		if line != "quit()" {
			exp, err := math.Evaluate(line)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			m, _ := toStruct(exp)
			out, _ := toTable(m)
			fmt.Println(out)
		} else {
			return nil
		}
	}
}
