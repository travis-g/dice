/*
AST parses a dice roll expression into an abstract syntax tree.

See https://github.com/alecthomas/participle
*/
package main

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/travis-g/dice"
)

var (
	ErrEmptyExpression = errors.New("empty expression")
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
	Subexpr       *Expr            `( "(" @@ ")"`
	Query         *Query           `| "?{" @@ "}" )`
	Number        *float64         `| @(("-" | "+")? (Int | Float))`
	Dice          *Dice            `| ( @SimpleNotation`
	Modifiers     []*Modifier      `@@*`
	GroupModifier []*GroupModifier `@@*`
	Label         string           `("[" @~"]" "]")? )`
	Func          *Func            `| @@`
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
	// FIXME: not all types need or can have a comparison operator and/or value!
	Type      string `@(Drop | Keep | Reroll | CriticalSuccess | CriticalFailure | Sort | Explode)`
	CompareOp string `@ComparisonOp?`
	Value     int    `@ComparisonValue?`
}

type GroupModifier struct {
	Failure   bool   `@("f"|"F")?`
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

// The lexer state machine rules. Rules are checked in the order that they
// appear within the named rule set.
var rules = lexer.Rules{
	"_": {
		{Name: "InlineWhitespace", Pattern: `[ \t]+`},
		{Name: "EOL", Pattern: `[\n\r]+`},
		{Name: "Whitespace", Pattern: `[ \t\n\r]+`},
		{Name: "End", Pattern: `$`},
	},
	"Numbers": {
		{Name: "Float", Pattern: `[-+]?\d*\.\d+`},
		{Name: "Int", Pattern: `[-+]?\d+`},
		{Name: "Uint", Pattern: `\d+`},
	},
	"Root": {
		{Name: "SimpleNotation", Pattern: `(?i)\d* *d(\d+|F)`, Action: lexer.Push("Modifiers")},
		lexer.Include("_"),
		{Name: "Comment", Pattern: `(//|\\|#)`, Action: lexer.Push("Comment")},
		{Name: "Operator", Pattern: `\*\*|[-+*^%/]|<<|>>`},
		lexer.Include("Numbers"),
		{Name: "Expr", Pattern: `\(`, Action: lexer.Push("Expr")},
		{Name: "Ident", Pattern: `(?i)[a-z][a-z_]{2,}`},
		{Name: "Comma", Pattern: `,`},
		{Name: "Query", Pattern: `\?{`, Action: lexer.Push("Query")},
		{Name: "RollGroup", Pattern: `{`, Action: lexer.Push("RollGroup")},
		// {Name: "InlineExpr", Pattern: `\[\[`, Action: lexer.Push("InlineExpr")},
		{Name: "Label", Pattern: `\[`, Action: lexer.Push("Label")},
		{Name: "Char", Pattern: `\$|[^$]+`},
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
	"Modifiers": {
		{Name: "Drop", Pattern: `(?i)d[lh]?`, Action: lexer.Push("ModifierValue")},
		{Name: "Keep", Pattern: `(?i)k[lh]?`, Action: lexer.Push("ModifierValue")},
		{Name: "Reroll", Pattern: `(?i)ro?`, Action: lexer.Push("Comparison")},
		{Name: "Sort", Pattern: `(?i)s[ad]?`},
		{Name: "CriticalSuccess", Pattern: `(?i)cs`, Action: lexer.Push("Comparison")},
		{Name: "CriticalFailure", Pattern: `(?i)cf`, Action: lexer.Push("Comparison")},
		{Name: "Explode", Pattern: `(?i)![!p]?`, Action: lexer.Push("Comparison")},
		{Name: "Failure", Pattern: `(?i)f`, Action: lexer.Push("Comparison")}, // TODO: GroupComparison?
		lexer.Include("Comparison"), // GroupModifier
		lexer.Return(),
	},
	"Comparison": {
		{Name: "ComparisonOp", Pattern: `[<>=]`, Action: lexer.Push("ModifierValue")},
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
	// participle.UseLookahead(2),
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
		return nil, ErrEmptyExpression
	}
	if trace {
		return Parser.ParseString(expression, expression,
			participle.Trace(os.Stderr))
	} else {
		return Parser.ParseString(expression, expression)
	}
}

// ptr returns the pointer to the passed value.
func ptr[T any](v T) *T {
	return &v
}
