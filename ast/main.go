/*
AST parses a dice roll expression into an abstract syntax tree.

See https://github.com/alecthomas/participle
*/
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/travis-g/dice"
)

var (
	ErrorEmptyExpression = errors.New("empty expression")
)

// A Root is the top level AST node of any individual dice roll expression.
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
	Subexpr       *Expr          `( "(" @@ ")"`
	Query         *Query         `| "?{" @@ "}" )`
	Number        *float64       `| @(("-" | "+")? (Int | Float))`
	Dice          *Dice          `| ( @SimpleNotation`
	Modifiers     []*Modifier    `@@*`
	GroupModifier *GroupModifier `@@?`
	Label         string         `("[" @~"]" "]")? )`
	Func          *Func          `| @@`
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
	Count *int `@Uint? ("d"|"D")`
	Size  *int `( @Uint`
	Fudge bool `| @("F"|"f") )`
}

type Quantity struct {
	Number *int   `@Uint`
	Expr   *Expr  `| "(" @@ ")"`
	Query  *Query `| "?{" @@ "}"`
}

type Modifier struct {
	// FIXME: not all types can have a comparison operator and/or value
	Type      string `@(Drop | Keep | Reroll | CriticalSuccess | CriticalFailure | Sort | Explode)`
	CompareOp string `@ComparisonOp?`
	Value     int    `@ComparisonValue?`
}

type GroupModifier struct {
	CompareOp string `@ComparisonOp`
	Value     int    `@ComparisonValue`
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
	return nil
}

// The lexer's state machine rules. Rules are checked in the order that they
// appear.
var rules = lexer.Rules{
	"Root": {
		{Name: "SimpleNotation", Pattern: `\d* *[dD](\d+|F|f)`, Action: lexer.Push("Modifiers")},
		{Name: "InlineWhitespace", Pattern: `[ \t]+`},
		{Name: "EOL", Pattern: `[\n\r]+`},
		{Name: "Whitespace", Pattern: `[ \t\n\r]+`},
		{Name: "Expr", Pattern: `\(`, Action: lexer.Push("Expr")},
		{Name: "Comment", Pattern: `(//|\\|#)`, Action: lexer.Push("Comment")},
		{Name: "Operator", Pattern: `\*\*|[-+*^%/]|<<|>>`},
		lexer.Include("Numbers"),
		{Name: "Ident", Pattern: `[a-zA-Z]{3,}`},
		{Name: "Comma", Pattern: `,`},
		{Name: "Query", Pattern: `\?{`, Action: lexer.Push("Query")},
		{Name: "RollGroup", Pattern: `{`, Action: lexer.Push("RollGroup")},
		{Name: "InlineExpr", Pattern: `\[\[`, Action: lexer.Push("InlineExpr")},
		{Name: "Label", Pattern: `\[`, Action: lexer.Push("Label")},
		{Name: "Char", Pattern: `\$|[^$]+`},
		{Name: "End", Pattern: `$`},
	},
	"Numbers": {
		{Name: "Float", Pattern: `[-+]?\d*\.\d+`},
		{Name: "Int", Pattern: `[-+]?\d+`},
		{Name: "Uint", Pattern: `\d+`},
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
		{Name: "Uint", Pattern: `\d+`},
		{Name: "Fudge", Pattern: `[Ff]`},
	},
	"Modifiers": {
		{Name: "Drop", Pattern: `d[lh]?`, Action: lexer.Push("ModifierValue")},
		{Name: "Keep", Pattern: `k[lh]?`, Action: lexer.Push("ModifierValue")},
		{Name: "Reroll", Pattern: `ro?`, Action: lexer.Push("Comparison")},
		{Name: "Sort", Pattern: `s[ad]?`},
		{Name: "CriticalSuccess", Pattern: `cs`, Action: lexer.Push("Comparison")},
		{Name: "CriticalFailure", Pattern: `cf`, Action: lexer.Push("Comparison")},
		{Name: "Explode", Pattern: `![!p]?`, Action: lexer.Push("Comparison")},
		lexer.Include("Comparison"), // GroupModifier
		lexer.Return(),
	},
	"Comparison": {
		{Name: "ComparisonOp", Pattern: `[<>]`, Action: lexer.Push("ModifierValue")},
		{Name: "ComparisonValue", Pattern: `-?\d+`, Action: lexer.Pop()},
	},
	"ModifierValue": {
		{Name: "ComparisonValue", Pattern: `-?\d+`, Action: lexer.Pop()},
		lexer.Return(),
	},
	"GroupComparison": {
		{Name: "GroupComparisonOp", Pattern: `[<>=]`},
		{Name: "GroupComparisonValue", Pattern: `-?\d+`, Action: lexer.Pop()}, // Float?
	},
	"RollGroup": {
		// TODO
		{Name: "RollGroupEnd", Pattern: `}`, Action: lexer.Pop()},
	},
	"Query": {
		{Name: "QueryEnd", Pattern: `}`, Action: lexer.Pop()},
		{Name: "InlineWhitespace", Pattern: `[ \t]+`},
		{Name: "QueryText", Pattern: `[^,|}]+`},
		{Name: "QueryPunctuation", Pattern: `[,|]`},
	},
	"Label": {
		{Name: "LabelEnd", Pattern: `]`, Action: lexer.Pop()},
		{Name: "InlineWhitespace", Pattern: `[ \t]+`},
		{Name: "LabelText", Pattern: `[^\]]+`},
	},
	"Comment": {
		{Name: "InlineWhitespace", Pattern: `[ \t]+`},
		{Name: "CommentText", Pattern: `.+`},
		lexer.Return(),
	},
}

var Lexer = lexer.MustStateful(rules)

var Parser = participle.MustBuild[Root](
	participle.Lexer(Lexer),
	participle.Elide("Whitespace", "InlineWhitespace", "Comment"),
	participle.UseLookahead(2),
)

// Sanitize is a helper to remove odd formatting from an expression, such as
// leading and trailing whitespace.
func Sanitize(expression string) (sanitized string) {
	sanitized = strings.Trim(expression, " \n\t\r")
	return
}

// ParseString is a wrapper to parse an expression. Expression strings are
// assumed to be sanitized or linted before being parsed.
func ParseString(expression string, trace bool) (*Root, error) {
	if expression == "" {
		return nil, ErrorEmptyExpression
	}
	if trace {
		return Parser.ParseString(expression, expression,
			participle.Trace(os.Stderr))
	} else {
		return Parser.ParseString(expression, expression)
	}
}

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

// ptr returns the pointer to the passed value.
func ptr[T any](v T) *T {
	return &v
}
