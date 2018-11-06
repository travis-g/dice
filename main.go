package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/travis-g/draas/command"
	"github.com/urfave/cli"
)

func toJson(i interface{}) (string, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func main() {
	cmd := cli.NewApp()
	cmd.Name = "dice"
	cmd.Usage = "CLI dice roller"
	cmd.Version = "0.0.1"

	cmd.Commands = []cli.Command{
		cli.Command{
			Name:    "roll",
			Aliases: []string{"r"},
			Usage:   "roll dice",
			Action: func(c *cli.Context) error {
				return command.RollCommand(c)
			},
		},
		cli.Command{
			Name:    "eval",
			Aliases: []string{"e"},
			Usage:   "evaluate a dice expression",
			Action: func(c *cli.Context) error {
				return command.EvalCommand(c)
			},
		},
	}

	err := cmd.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
