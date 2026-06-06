package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/travis-g/dice"
)

// trace is a flag to enable tracing of the parser to stderr.
var trace *bool

// diagram is a flag to output the extended Backus–Naur form (EBNF) of the
// parser which can be used to visualize the grammar.
var diagram *bool

// query option parameters to supply to the evaluation context.
var params keyValueFlag

// keyValueFlag is a key=value format flag used for supplying query option
// values.
type keyValueFlag map[string]string

var _ flag.Value = (*keyValueFlag)(nil)

func (kv *keyValueFlag) String() string {
	return fmt.Sprintf("%v", map[string]string(*kv))
}

func (kv *keyValueFlag) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return errors.New("expected key=value")
	}

	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])

	if key == "" {
		return errors.New("empty key")
	}

	if *kv == nil {
		*kv = make(keyValueFlag)
	}

	(*kv)[key] = val
	return nil
}

func init() {
	// define CLI flags only once
	trace = flag.Bool("trace", false, "trace the parser to stderr")
	diagram = flag.Bool("diagram", false, "output parser diagram code")
	flag.Var(&params, "param", "key-value query parameter pairs (repeatable)")
}

// main is the entry point for the dice parser as a CLI tool. It's useful for
// testing and debugging the parser.
func main() {
	ctx := context.Background()

	flag.Parse()
	if *diagram {
		fmt.Fprintf(os.Stdout, Parser.String())
		return
	}

	ctx = context.WithValue(ctx, dice.CtxKeyParameters, map[string]string(params))
	if err := json.NewEncoder(os.Stdout).Encode(params); err != nil {
		panic(err)
	}

	expr, err := ParseString(flag.Arg(0), *trace)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(expr); err != nil {
		panic(err)
	}

	result, err := expr.Evaluate(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Fprintln(os.Stdout, result)
	}
}
