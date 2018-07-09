package main

import (
	"go/ast"
	"go/token"
	"strconv"
)

// The two Eval* functions within this file can simplify basic expressions. If
// the AST contains Lparen/Rparen tokens the contents between them are returned
// as 0. It works exceptionally well for evaluating 1+2+4-8 etc., but was not
// effective enough.

// Eval returns the integer equivalent of the AST of a mathematical expression.
func Eval(exp ast.Expr) int {
	switch exp := exp.(type) {
	case *ast.BinaryExpr:
		return EvalBinaryExpr(exp)
	case *ast.BasicLit:
		switch exp.Kind {
		case token.INT:
			i, _ := strconv.Atoi(exp.Value)
			return i
		}
	}

	return 0
}

// EvalBinaryExpr evaluates basic mathematics operations of an AST. Does not
// handle parenthesis correctly, making it difficult to use in practice.
func EvalBinaryExpr(exp *ast.BinaryExpr) int {
	left := Eval(exp.X)
	right := Eval(exp.Y)

	switch exp.Op {
	case token.ADD:
		return left + right
	case token.SUB:
		return left - right
	case token.MUL:
		return left * right
	case token.QUO:
		return left / right
	}

	return 0
}
