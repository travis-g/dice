package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

var (
	trace   *bool
	diagram *bool
)

func init() {
	// define CLI flags only once
	trace = flag.Bool("trace", false, "trace the parser to stderr")
	diagram = flag.Bool("diagram", false, "output parser diagram code")
}

func main() {
	flag.Parse()
	if *diagram {
		fmt.Fprintf(os.Stdout, Parser.String())
		return
	}

	expr, err := ParseString(flag.Arg(0), *trace)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(expr); err != nil {
		panic(err)
	}
}
