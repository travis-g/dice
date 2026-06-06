package main

import "github.com/alecthomas/participle/v2/lexer"

// Rules for whitespaces and end of lines. These are included in the Root state
// to allow for whitespace between tokens, but are not included in other states.
var (
	InlineWhitespace = lexer.Rule{Name: "InlineWhitespace", Pattern: `[ \t]+`}
	Whitespace       = lexer.Rule{Name: "Whitespace", Pattern: `[ \t\n\r]+`}
	EOL              = lexer.Rule{Name: "EOL", Pattern: `[\n\r]+`}
	End              = lexer.Rule{Name: "End", Pattern: `$`}
)

// Rules for identifiers, numbers, and other basic tokens.
var (
	Ident = lexer.Rule{Name: "Ident", Pattern: `(?i)[a-z][a-z_0-9]{2,}`}
	Float = lexer.Rule{Name: "Float", Pattern: `[-+]?\d*\.\d+`}
	Int   = lexer.Rule{Name: "Int", Pattern: `[-+]?\d+`}
	Uint  = lexer.Rule{Name: "Uint", Pattern: `\d+`}
	Comma = lexer.Rule{Name: "Comma", Pattern: `,`}
	Char  = lexer.Rule{Name: "Char", Pattern: `\$|[^$]+`}
)

// Rules for parsing the various components between tokens.
var (
	MathOperator = lexer.Rule{Name: "MathOperator", Pattern: `[-+*^%/]|\*\*|<<|>>`}
)
