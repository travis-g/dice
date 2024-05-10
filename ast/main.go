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
	Expr *Expr `@@`
}

type Expr struct {
	L *Factor     `@@ " "?`
	R []*OpFactor `@@*`
}

type OpFactor struct {
	Op     string  `@Operator`
	Factor *Factor `@@`
}

type Factor struct {
	Number float64 `@SignedInt | @Float | @Uint |`
	Dice   *Dice   `@Notation |`
	Expr   *Expr   `"(" @@ ")" |`
	Query  *Query  `"?{" @@ "}"`
}

type Query struct {
	Name string `@QueryName`
}

type Dice struct {
	X         string `@@`
	Y         string `"d" @@`
	Modifiers string `@Char?`
	Label     string `("[" @~"]" "]")?`
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
		{Name: "Whitespace", Pattern: `[ \t]+`, Action: nil},
		{Name: "Notation", Pattern: `(?i)\d* *d(\d+|F)([a-z]\d)*(\[[^\]]+])?`, Action: nil}, // TODO: stateful
		{Name: "Expr", Pattern: `\(`, Action: lexer.Push("Expr")},
		{Name: "CommentStart", Pattern: `\/\/|\|#[^\n]*`, Action: nil}, // TODO: stateful
		{Name: "Operator", Pattern: `\*\*|[-\*^%/+]|<<|>>`, Action: nil},
		{Name: "Float", Pattern: `-?\d*\.\d`, Action: nil},
		lexer.Include("SignedInt"),
		lexer.Include("Uint"),
		{Name: "Query", Pattern: `\?{`, Action: lexer.Push("Query")},
		{Name: "Ident", Pattern: `[a-zA-Z][a-zA-Z\d]*`, Action: nil},
		{Name: "EOL", Pattern: `[\n\r]+`, Action: nil},
		{Name: "Char", Pattern: `\$|[^$]+`, Action: nil},
	},
	"Expr": {
		{Name: "ExprEnd", Pattern: `\)`, Action: lexer.Pop()},
		lexer.Include("Root"),
	},
	"InlineExpr": { // TODO
		{Name: "InlineExpr", Pattern: `\[\[`, Action: lexer.Push("InlineExpr")},
		{Name: "InlineExprEnd", Pattern: `]]`, Action: lexer.Pop()},
		// lexer.Include("Expr"),
	},
	"Notation": { // TODO
		{Name: "NotationEnd", Pattern: ``, Action: lexer.Pop()},
		{Name: "Size", Pattern: `\d+|[fF]`, Action: nil},
	},
	"Modifiers": {
		{"Drop", `d[lh]?`, nil},
		{"Keep", `k[lh]?`, nil},
		{"Reroll", `ro?`, nil},
		{"CriticalSuccess", `cs`, nil},
		{"CriticalFailure", `cf`, nil},
		{"Sort", `s[ad]?`, nil},
	},
	"Comparison": {
		{"ComparisonOperator", `(<|>|=)`, nil},
	},
	"Uint": {
		{"Uint", `\d+`, nil},
	},
	"SignedInt": {
		{Name: "SignedInt", Pattern: `[-+]\d+`, Action: nil},
	},
	"RollGroup": { // TODO
		{Name: "RollGroupEnd", Pattern: `}`, Action: lexer.Pop()},
	},
	"Query": {
		{Name: "QueryEnd", Pattern: `}`, Action: lexer.Pop()},
		{Name: "QueryName", Pattern: `[^}]+`, Action: nil},
		// {Name: "QueryChoice", Pattern: `[^|}]+`, Action: nil},
	},
	"Label": {
		{Name: "LabelEnd", Pattern: `]`, Action: lexer.Pop()},
		{Name: "LabelText", Pattern: `[^\]]+`, Action: nil},
	},
	"Comment": {
		// {EOL},
		{Name: "Comment", Pattern: `.+`, Action: nil},
	},
}

var Lexer = lexer.MustStateful(rules)

var parser = participle.MustBuild[Root](
	participle.UseLookahead(3),
	participle.Elide("Whitespace"),
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
