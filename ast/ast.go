// AST parses a dice roll expression into an abstract syntax tree.
//
// See https://github.com/alecthomas/participle
//
// # Linting
//
// Reflecting the parsed results of a whitespace-heavy vs. whitespace-less
// expression string should result in the same parsed AST. As examples, the
// following two strings should result in the same ASTs:
//
//  ' 3 d6 [foo] //  bar'
//  '3d6[foo]#bar'
//
// Linting standards are in flux, but as a general rule, String methods should
// preserve any existing whitespace: "no linting when printing".

package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/travis-g/dice"
)

var (
	ErrEmptyExpression    = errors.New("empty expression")
	ErrNotImplemented     = errors.New("not implemented")
	ErrUnreachable        = errors.New("unreachable code")
	ErrInvalidStructField = errors.New("invalid struct field")
)

type Stringer interface {
	String() string
	// TODO: add a "compact" string method
	// CompactString() string
}

// A Root is the top level AST node of any individual dice roll expression.
type Root struct {
	Expr    *Expr   `parser:"@@?"`
	Comment *string `(("//" | "\\" | "#") @CommentText?)?`
}

func (r *Root) String() string {
	buf := new(bytes.Buffer)
	if r.Expr != nil {
		buf.WriteString(r.Expr.String())
	}
	if r.Expr != nil && r.Comment != nil {
		buf.WriteRune(' ')
	}
	if r.Comment != nil && *r.Comment != "" {
		buf.Write([]byte{'#', ' '})
		buf.WriteString(*r.Comment)
	}
	return buf.String()
}

// An Expr is a full expression, which may consist of a single Term or a Term
// followed by one or more operator-Term pairs.
type Expr struct {
	L *Term     `parser:"@@"`
	R []*OpTerm `parser:"@@*"`
}

func (e *Expr) String() string {
	buf := new(bytes.Buffer)
	if e.L != nil {
		buf.WriteString(e.L.String())
	}
	if len(e.R) > 0 {
		buf.WriteRune(' ')
		for _, r := range e.R {
			if r != nil {
				buf.WriteString(r.String())
			}
		}
	}
	return buf.String()
}

// An OpTerm is an operator followed by a Term.
type OpTerm struct {
	Op   string `parser:"@Operator"`
	Term *Term  `parser:"@@"`
}

func (ot *OpTerm) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString(ot.Op)
	buf.WriteRune(' ')
	buf.WriteString(ot.Term.String())
	return buf.String()
}

// A Term is an individual part of an expression. Terms should resolve to a
// value.
type Term struct {
	Subexpr *Expr    `parser:"( '(' @@ ')'"`
	Query   *Query   `parser:"| '?{' @@ '}' )"`
	Number  *float64 `parser:"| @(('-' | '+')? (Int | Float))"`
	Dice    *Dice    `parser:"| ( @SimpleNotation"`
	// TODO: RollGroup
	Modifiers     []*Modifier      `parser:"@@*"`
	GroupModifier []*GroupModifier `parser:"@@*"`
	Func          *Func            `parser:"| @@)"`
	Label         *string          `parser:"('[' @~']' ']')?"`
}

func (t *Term) String() string {
	buf := new(bytes.Buffer)
	if t.Subexpr != nil {
		buf.WriteRune('(')
		buf.WriteString(t.Subexpr.String())
		buf.WriteRune(')')
		return buf.String()
	}
	if t.Query != nil {
		buf.WriteString(t.Query.String())
		return buf.String()
	}
	if t.Number != nil {
		buf.WriteString(strconv.FormatFloat(*t.Number, 'f', -1, 64))
	}
	if t.Dice != nil {
		buf.WriteString(t.Dice.String())
	}
	if len(t.Modifiers) > 0 {
		for _, m := range t.Modifiers {
			buf.WriteString(m.String())
		}
	}
	if len(t.GroupModifier) > 0 {
		// TODO
		// buf.WriteString(t.GroupModifier.String())
	}
	if t.Func != nil {
		buf.WriteString(t.Func.String())
	}
	if t.Label != nil {
		buf.WriteRune('[')
		buf.WriteString(*t.Label)
		buf.WriteRune(']')
	}
	return buf.String()
}

type Func struct {
	Name string `parser:"@Ident"`
	Args *Args  `parser:"'(' @@? ')'"`
}

func (f *Func) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString(f.Name)
	buf.WriteRune('(')
	if f.Args != nil {
		buf.WriteString(f.Args.String())
	}
	buf.WriteRune(')')
	return buf.String()
}

// Args is a list of arguments to a function call, separated by commas.
type Args struct {
	Arg []*Term `parser:"@@ (',' @@)*"`
}

func (a *Args) String() string {
	var args []string
	for _, arg := range a.Arg {
		args = append(args, arg.String())
	}
	return strings.Join(args, ", ")
}

type Query struct {
	Name    string         `parser:"@QueryText"`
	Options []*QueryOption `parser:"('|' @@)*"`
}

func (q *Query) String() string {
	buf := new(bytes.Buffer)
	buf.Write([]byte{'?', '{'})
	buf.WriteString(q.Name)
	for _, opt := range q.Options {
		buf.WriteByte('|')
		if opt.OptionLabel != "" {
			buf.WriteString(opt.OptionLabel)
			buf.Write([]byte{',', ' '})
		}
		buf.WriteString(opt.OptionValue)
	}
	buf.WriteRune('}')
	return buf.String()
}

// QueryOptions are suggested/available values for a query answer.
type QueryOption struct {
	OptionLabel string `parser:"(@QueryText ',')?"`
	OptionValue string `parser:"@QueryText"`
}

func (qo *QueryOption) String() string {
	buf := new(bytes.Buffer)
	if qo.OptionLabel != "" {
		buf.WriteString(qo.OptionLabel)
		buf.WriteString(", ")
	}
	buf.WriteString(qo.OptionValue)
	return buf.String()
}

type Dice struct {
	// TODO: Sub-expressions should be supported as possible Dice counts and
	// sizes
	_     *Expr `parser:"( '(' @@ ')'"`
	Count *int  `parser:"| @Uint? ) ('d'|'D')"`
	Size  *int  `parser:"( @Uint"`
	Fudge bool  `parser:"| @('F'|'f')"`
	_     *Expr `parser:"| '(' @@ ')' )"`
}

// Capture loads a Dice structure from a string slice of values. The values
// slice is expected to be a single element.
func (d *Dice) Capture(values []string) error {
	// HACK: parse byte by byte rather than regex
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

func (d *Dice) String() string {
	buf := new(bytes.Buffer)
	if d.Count != nil {
		if *d.Count != 1 {
			fmt.Fprintf(buf, "%d", *d.Count)
		}
	}
	buf.WriteRune('d')
	if d.Fudge {
		buf.WriteRune('F')
	} else {
		fmt.Fprintf(buf, "%d", *d.Size)
	}
	return buf.String()
}

type Quantity struct {
	Number  *int   `parser:"@Uint"`
	Subexpr *Expr  `parser:"| '(' @@ ')'"`
	Query   *Query `parser:"| '?{' @@ '}' )"`
}

func (q *Quantity) String() string {
	buf := new(bytes.Buffer)
	switch {
	case q.Number != nil:
		fmt.Fprintf(buf, "%d", *q.Number)
	case q.Subexpr != nil:
		buf.WriteRune('(')
		buf.WriteString(q.Subexpr.String())
		buf.WriteRune(')')
	case q.Query != nil:
		buf.WriteString(q.Query.String())
	default:
		panic(ErrUnreachable)
	}
	return buf.String()
}

// A modifier changes how a roll/set of dice has its result(s) calculated, or
// how it is displayed/tracked internally.
// FIXME: not all types need or can have a comparison operator and/or value.
type Modifier struct {
	Type      string  `parser:"@(Drop | Keep | Reroll | CriticalSuccess | CriticalFailure | Sort | Explode)"`
	CompareOp *string `parser:"@ComparisonOp?"`
	Value     *int    `parser:"@ComparisonValue?"`
}

func (m *Modifier) String() string {
	// TODO(travis-g): implement all modifier types
	buf := new(bytes.Buffer)
	switch m.Type {
	case "Drop":
	case "Keep":
	case "Reroll":
	case "CriticalSuccess":
	case "CriticalFailure":
	case "Sort":
	case "Explode":
	default:
		panic(ErrNotImplemented)
	}
	return buf.String()
}

type GroupModifier struct {
	Failure   bool   `parser:"@('F'|'f')?"`
	CompareOp string `parser:"@ComparisonOp"`
	Value     int    `parser:"@ComparisonValue"`
}

func (gm *GroupModifier) String() string {
	buf := new(bytes.Buffer)
	if gm.Failure {
		buf.WriteString("f")
	}
	buf.WriteString(gm.CompareOp)
	fmt.Fprintf(buf, "%d", gm.Value)
	return buf.String()
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
		{Name: "Comma", Pattern: `,`},
		{Name: "Query", Pattern: `\?{`, Action: lexer.Push("Query")},
		{Name: "RollGroup", Pattern: `{`, Action: lexer.Push("RollGroup")},
		// {Name: "InlineExpr", Pattern: `\[\[`, Action: lexer.Push("InlineExpr")},
		{Name: "Label", Pattern: `\[`, Action: lexer.Push("Label")},
		{Name: "Ident", Pattern: `(?i)[a-z][a-z_0-9]{2,}`},
		{Name: "Char", Pattern: `\$|[^$]+`},
	},
	"Expr": {
		{Name: "ExprEnd", Pattern: `\)`, Action: lexer.Pop()},
		lexer.Include("Root"),
		lexer.Return(),
	},
	"InlineExpr": {
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

// ParseString is a wrapper to parse an expression from an input string.
// Expression strings are assumed to be sanitized or linted before being parsed.
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
