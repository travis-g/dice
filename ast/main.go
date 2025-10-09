package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// trace is a flag to enable tracing of the parser to stderr.
var trace *bool

// diagram is a flag to output the extended Backus–Naur form (EBNF) of the
// parser which can be used to visualize the grammar.
var diagram *bool

func init() {
	// define CLI flags only once
	trace = flag.Bool("trace", false, "trace the parser to stderr")
	diagram = flag.Bool("diagram", false, "output parser diagram code")
}

// main is the entry point for the dice parser CLI tool. It's useful for testing
// and debugging the parser.
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
