/*
Package math implements dice expression mathematics and functions.

When evaluating an expression with Expression, the package follows order of operations:

	All dice notations/expressions are rolled and expanded,
	Parenthesis (deepest first),
	Functions,
	Exponentiation,
	Multiplication, division, and modulus from left to right,
	Addition and subtraction from left to right

The math package currently relies heavily on github.com/Knetic/govaluate, but
only uses a subset of the package's features.
*/
package math
