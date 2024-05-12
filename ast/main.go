package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type Root struct {
	Expr    *Expr  `@@`
	Comment string `(("//" | "\\" | "#") @CommentText?)?`
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
	Number  *float64 `@(("-" | "+")? (Int | Float))`
	Dice    *Dice    `| @SimpleNotation`
	Func    *Func    `| @@`
	Subexpr *Expr    `| "(" @@ ")"`
	Query   *Query   `| "?{" @@ "}"`
}

type Func struct {
	Name string  `@Ident`
	Args []*Args `"(" (@@ ( "," @@ )*)? ")"`
}

type Args struct {
	A []*Factor `@@ ("," @@)*`
}

type Query struct {
	Name    string         `@QueryText` // FIXME
	Options []*QueryOption `("|" @@)*`
}

type QueryOption struct {
	OptionLabel string `(@QueryText ",")?`
	OptionValue string `@QueryText`
}

type Dice struct {
	// FIXME
	X         float64 `@@`
	Y         string  `"d" @@`
	Modifiers string  `@Char?`
	Label     string  `("[" @~"]" "]")?`
}

func (d *Dice) Capture(values []string) error {
	return d.ParseNotation(values[0])
}

type RegexPattern string

const (
	RegexComparePoint    = `(=|<|>)`
	RegexModifierExplode = `![!p]?` + RegexComparePoint + `?`
)

// The lexer's state machine rules
var rules = lexer.Rules{
	"Root": {
		{Name: "SimpleNotation", Pattern: `(?i)\d* *d(\d+|F)([a-z!=<>]+\d*)*(\[[^\]]+])?`, Action: nil}, // TODO: stateful
		{Name: "InlineWhitespace", Pattern: `[ \t]+`, Action: nil},
		{Name: "Whitespace", Pattern: `[ \t\n\r]+`, Action: nil},
		{Name: "Ident", Pattern: `[a-zA-Z]{3,}`, Action: nil},
		{Name: "Expr", Pattern: `\(`, Action: lexer.Push("Expr")},
		{Name: "CommentStart", Pattern: `(//|\\|#)`, Action: lexer.Push("Comment")},
		{Name: "Operator", Pattern: `\*\*|[-+\*^%/]|<<|>>`, Action: nil},
		{Name: "Float", Pattern: `[-+]?\d*\.\d+`, Action: nil},
		{Name: "Int", Pattern: `[-+]?\d+`, Action: nil},
		{Name: "QueryStart", Pattern: `\?{`, Action: lexer.Push("Query")},
		{Name: "EOL", Pattern: `[\n\r]+`, Action: nil},
		{Name: "Comma", Pattern: `,`, Action: nil},
		{Name: "Char", Pattern: `\$|[^$]+`, Action: nil},
	},
	"Expr": {
		{Name: "ExprEnd", Pattern: `\)`, Action: lexer.Pop()},
		lexer.Include("Root"),
		lexer.Return(),
	},
	"InlineExpr": { // TODO
		// {Name: "InlineExpr", Pattern: `\[\[`, Action: lexer.Push("InlineExpr")},
		{Name: "InlineExprEnd", Pattern: `]]`, Action: lexer.Pop()},
		lexer.Include("Expr"),
	},
	"StatefulNotation": {
		// TODO
	},
	"Modifiers": {
		{"Drop", `d[lh]?`, nil},
		{"Keep", `k[lh]?`, nil},
		{"Reroll", `ro?`, nil},
		{"CriticalSuccess", `cs`, nil},
		{"CriticalFailure", `cf`, nil},
		{"Sort", `s[ad]?`, nil},
		{"Explode", `![!p]?`, nil},
		{"ComparePointOp", `[=<>]`, nil},
		{"ComparePointValue", `\d+`, nil},
		lexer.Return(),
	},
	"GroupComparison": {
		{"GroupComparisonOperator", `(<|>|=)`, nil},
		{"GroupComparisonPoint", `\d+`, lexer.Pop()}, // Float?
	},
	"RollGroup": { // TODO
		{Name: "RollGroupEnd", Pattern: `}`, Action: lexer.Pop()},
	},
	"Query": { // FIXME
		{Name: "QueryText", Pattern: `[^,|}]+`, Action: nil},
		{Name: "QueryPunctuation", Pattern: `[,|]`, Action: nil},
		{Name: "QueryEnd", Pattern: `}`, Action: lexer.Pop()},
	},
	"Label": {
		{Name: "LabelEnd", Pattern: `]`, Action: lexer.Pop()},
		{Name: "LabelText", Pattern: `[^\]]+`, Action: nil},
	},
	"Comment": {
		{Name: "CommentText", Pattern: `.+`},
	},
}

var Lexer = lexer.MustStateful(rules)

var parser = participle.MustBuild[Root](
	participle.UseLookahead(3),
	participle.Elide("InlineWhitespace", "CommentStart", "CommentText"),
	participle.Lexer(Lexer),
)

func ParseString(expression string, trace bool) (*Root, error) {
	if trace {
		return parser.ParseString(expression, expression,
			participle.Trace(os.Stdout))
	} else {
		return parser.ParseString(expression, expression)
	}
}

func main() {
	debug := flag.Bool("debug", false, "run parser in debug mode")
	flag.Parse()
	expr, err := ParseString(flag.Arg(0), *debug)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(expr); err != nil {
		panic(err)
	}
}
