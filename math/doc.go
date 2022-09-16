/*
Package math implements dice expression mathematics and functions.

When evaluating an expression with Expression, the package follows order of
operations:

	All dice notations/expressions are rolled and expanded,
	Parenthesis (deepest first),
	Functions,
	Exponentiation,
	Multiplication, division, and modulus from left to right,
	Addition and subtraction from left to right

The math package currently relies heavily on
https://github.com/Knetic/govaluate.

# Benchmarks

The benchmarks for the math package's functions use math/rand as the random byte
source rather than crypto/rand in order to limit slowness attributed to sourcing
entropy.
*/
package math
