package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/travis-g/dice/cmd/dice/command"
	"github.com/urfave/cli"
)

func main() {
	cmd := cli.NewApp()
	cmd.Name = "dice"
	cmd.Usage = "CLI dice roller"
	cmd.Version = "0.0.1"

	cmd.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "format",
			Value:  "table",
			Usage:  "output format",
			EnvVar: "FORMAT",
		},
	}

	cmd.Commands = []cli.Command{
		cli.Command{
			Name:    "eval",
			Aliases: []string{"e"},
			Usage:   "evaluate a dice expression",
			Action: func(c *cli.Context) error {
				return command.EvalCommand(c)
			},
		},
		cli.Command{
			Name:  "repl",
			Usage: "enter a REPL mode",
			Action: func(c *cli.Context) error {
				return command.REPLCommand(c)
			},
		},
		cli.Command{
			Name:    "roll",
			Aliases: []string{"r"},
			Usage:   "roll plain dice",
			Action: func(c *cli.Context) error {
				return command.RollCommand(c)
			},
		},
	}

	sort.Sort(cli.FlagsByName(cmd.Flags))
	sort.Sort(cli.CommandsByName(cmd.Commands))

	err := cmd.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
