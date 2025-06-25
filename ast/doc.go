// Package ast provides the abstract syntax tree (AST) for dice roll
// expressions.
//
// See https://github.com/alecthomas/participle
//
// # Linting
//
// Reflecting the parsed results of a whitespace-heavy vs. whitespace-less
// expression string should result in the same parsed AST. As examples, the
// following two strings should result in the same ASTs:
//
//	' 3 d6 [foo] //  bar'
//	'3d6[foo]#bar'
//
// Linting standards are in flux, but as a general rule, String methods should
// preserve any whitespace present in labels ("no linting when printing").
package main
