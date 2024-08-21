package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/travis-g/dice"
)

var (
	ErrorEmptyExpression = errors.New("empty expression")
)

// See https://github.com/alecthomas/participle

// The root of any individual expression
type Root struct {
	Expr    *Expr   `@@?`
	Comment *string `(("//" | "\\" | "#") @CommentText?)?`
}

type Expr struct {
	L *Factor     `@@`
	R []*OpFactor `@@*`
}

type OpFactor struct {
	Op     string  `@Operator`
	Factor *Factor `@@`
}

type Factor struct {
	Subexpr *Expr    `( "(" @@ ")"`
	Number  *float64 `| @(("-" | "+")? (Int | Float))`
	Query   *Query   `| "?{" @@ "}" )`
	Dice    *Dice    `| @SimpleNotation`
	Func    *Func    `| @@`
}

type Func struct {
	Name string  `@Ident`
	Args []*Args `"(" (@@ ( "," @@ )*)? ")"`
}

type Args struct {
	Arg []*Factor `@@ ("," @@)*`
}

type Query struct {
	Name    string         `@QueryText`
	Options []*QueryOption `("|" @@)*`
}

type QueryOption struct {
	OptionLabel string `(@QueryText ",")?`
	OptionValue string `@QueryText`
}

type Dice struct {
	Count     *int   `@Uint?`
	Size      *int   `("d"|"D") ( @Uint`
	Fudge     bool   `| @("F"|"f") )`
	Modifiers string `@Char?`
	Label     string `("[" @~"]" "]")?`
}

type Quantity struct {
	Number *int   `@Uint`
	Expr   *Expr  `| "(" @@ ")"`
	Query  *Query `| "?{" @@ "}"`
}

type Modifier struct {
	Type      string `@(Drop | Keep | Reroll | CriticalSuccess | CriticalFailure | Sort | Explode)`
	CompareOp string `@ComparePointOp?`
	Value     int    `@Uint?`
}

func (d *Dice) Capture(values []string) error {
	// TODO: parse byte by byte rather than regex
	components := dice.FindNamedCaptureGroups(dice.DiceWithModifiersExpressionRegex, values[0])

	if components["count"] != "" {
		c, err := strconv.Atoi(components["count"])
		if err != nil {
			return err
		}
		d.Count = ptr(c)
	} else {
		// implied count of 1
		d.Count = ptr(1)
	}
	if s, err := strconv.Atoi(components["size"]); err == nil {
		d.Size = ptr(s)
	} else if components["size"] == "F" || components["size"] == "f" {
		d.Fudge = true
		d.Size = ptr(1)
	}
	d.Modifiers = components["modifiers"]
	// FIXME: parse Label
	d.Label = ""
	return nil
}

// The lexer's state machine rules. Rules are checked in the order that they
// appear.
var rules = lexer.Rules{
	"Root": {
		{Name: "SimpleNotation", Pattern: `\d* *[dD](\d+|F|f)([a-zA-Z!=<>]+\d*)*(\[[^\]]+])?`}, // TODO: stateful
		{Name: "InlineWhitespace", Pattern: `[ \t]+`},
		{Name: "EOL", Pattern: `[\n\r]+`},
		{Name: "Whitespace", Pattern: `[ \t\n\r]+`},
		{Name: "Ident", Pattern: `[a-zA-Z]{3,}`},
		{Name: "Expr", Pattern: `\(`, Action: lexer.Push("Expr")},
		{Name: "CommentStart", Pattern: `(//|\\|#)`, Action: lexer.Push("Comment")},
		{Name: "Operator", Pattern: `\*\*|[-+*^%/]|<<|>>`},
		{Name: "Float", Pattern: `[-+]?\d*\.\d+`},
		{Name: "Int", Pattern: `[-+]?\d+`},
		{Name: "Uint", Pattern: `\d+`},
		{Name: "Comma", Pattern: `,`},
		{Name: "QueryStart", Pattern: `\?{`, Action: lexer.Push("Query")},
		{Name: "RollGroupStart", Pattern: `{`, Action: lexer.Push("RollGroup")},
		{Name: "InlineExprStart", Pattern: `\[\[`, Action: lexer.Push("InlineExpr")},
		{Name: "LabelStart", Pattern: `\[`, Action: lexer.Push("Label")},
		{Name: "Char", Pattern: `\$|[^$]+`},
		{Name: "End", Pattern: `$`},
	},
	"Expr": {
		{Name: "ExprEnd", Pattern: `\)`, Action: lexer.Pop()},
		lexer.Include("Root"),
		lexer.Return(),
	},
	"InlineExpr": { // TODO
		{Name: "InlineExprEnd", Pattern: `]]`, Action: lexer.Pop()},
		lexer.Include("Root"),
		lexer.Return(),
	},
	"StatefulNotation": {
		// TODO
	},
	"Modifiers": {
		{Name: "Drop", Pattern: `d[lh]?`},
		{Name: "Keep", Pattern: `k[lh]?`},
		{Name: "Reroll", Pattern: `ro?`},
		{Name: "CriticalSuccess", Pattern: `cs`},
		{Name: "CriticalFailure", Pattern: `cf`},
		{Name: "Sort", Pattern: `s[ad]?`},
		{Name: "Explode", Pattern: `![!p]?`},
		{Name: "ComparePointOp", Pattern: `[<>]`},
		{Name: "ComparePointValue", Pattern: `\d+`},
		lexer.Return(),
	},
	"GroupComparison": {
		{Name: "GroupComparisonOperator", Pattern: `[<>=]`},
		{Name: "GroupComparisonPoint", Pattern: `\d+`, Action: lexer.Pop()}, // Float?
		lexer.Return(),
	},
	"RollGroup": {
		// TODO
		{Name: "RollGroupEnd", Pattern: `}`, Action: lexer.Pop()},
		lexer.Return(),
	},
	"Query": {
		{Name: "InlineWhitespace", Pattern: `[ \t]+`},
		{Name: "QueryText", Pattern: `[^,|}]+`},
		{Name: "QueryPunctuation", Pattern: `[,|]`},
		{Name: "QueryEnd", Pattern: `}`, Action: lexer.Pop()},
		lexer.Return(),
	},
	"Label": {
		{Name: "InlineWhitespace", Pattern: `[ \t]+`},
		{Name: "LabelText", Pattern: `[^\]]+`},
		{Name: "LabelEnd", Pattern: `]`, Action: lexer.Pop()},
		lexer.Return(),
	},
	"Comment": {
		{Name: "InlineWhitespace", Pattern: `[ \t]+`},
		{Name: "CommentText", Pattern: `.+`},
		lexer.Return(),
	},
}

var Lexer = lexer.MustStateful(rules)

var Parser = participle.MustBuild[Root](
	participle.Elide("Whitespace", "InlineWhitespace", "CommentStart"),
	participle.Lexer(Lexer),
)

func ParseString(expression string, trace bool) (*Root, error) {
	if expression == "" {
		return nil, ErrorEmptyExpression
	}
	if trace {
		return Parser.ParseString(expression, expression,
			participle.Trace(os.Stdout))
	} else {
		return Parser.ParseString(expression, expression)
	}
}

func main() {
	debug := flag.Bool("debug", false, "run parser in debug mode")
	diagram := flag.Bool("diagram", false, "print parser diagram")
	flag.Parse()
	if *diagram {
		fmt.Fprintf(os.Stdout, Parser.String())
		return
	}
	expr, err := ParseString(flag.Arg(0), *debug)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(expr); err != nil {
		panic(err)
	}
}

// ptr returns the pointer to the passed value.
func ptr[T any](v T) *T {
	return &v
}
