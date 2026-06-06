package main

import (
	"context"
	"errors"
	"fmt"
)

// SkipNode is used as a return value from [WalkFunc] to indicate that the AST
// node is to be skipped.
var SkipNode = errors.New("skip this node")

// SkipAll is used as a return value from [WalkFunc] to indicate that all
// remaining AST nodes are to be skipped.
var SkipAll = errors.New("skip everything and stop the walk")

// A Node represents a leaf in the AST of a dice roll expression.
type Node interface {
	// Resolve traverses the Node and any children to ensures all required data
	// is fetched, such as [Query] parameters.
	Resolve(ctx context.Context) (Node, error)

	// Nodes should implement Stringer to provide a string.
	fmt.Stringer
}

// WalkFunc is a function that is called against [Node] during a walk.
type WalkFunc func(ctx context.Context, n Node, err error) error

// Walk traverses the AST starting at a given root [Node], calling the provided
// [WalkFunc] for each Node encountered.
func Walk(ctx context.Context, root Node, fn WalkFunc) error {
	// TODO: track depth of the walk/number of nodes walked to counter a
	// possible overflow
	if root == nil {
		return nil
	}
	err := fn(ctx, root, nil)
	if err != nil {
		if errors.Is(err, SkipNode) {
			return nil
		}
		return err
	}
	return walkChildren(ctx, root, fn)
}

func walkChildren(ctx context.Context, node Node, fn WalkFunc) error {
	switch n := node.(type) {
	case *Root:
		if n.Expr != nil {
			if err := Walk(ctx, n.Expr, fn); err != nil {
				return err
			}
		}
	case *Expr:
		if n.L != nil {
			if err := Walk(ctx, n.L, fn); err != nil {
				return err
			}
		}
		for _, opTerm := range n.R {
			if opTerm == nil {
				continue
			}
			if err := Walk(ctx, opTerm, fn); err != nil {
				return err
			}
		}
	case *OpTerm:
		if n.Term != nil {
			if err := Walk(ctx, n.Term, fn); err != nil {
				return err
			}
		}
	case *Term:
		if n.Subexpr != nil {
			if err := Walk(ctx, n.Subexpr, fn); err != nil {
				return err
			}
		}
		if n.Query != nil {
			if err := Walk(ctx, n.Query, fn); err != nil {
				return err
			}
		}
		if n.Dice != nil {
			if err := Walk(ctx, n.Dice, fn); err != nil {
				return err
			}
		}
		for _, mod := range n.Modifiers {
			if mod == nil {
				continue
			}
			if err := Walk(ctx, mod, fn); err != nil {
				return err
			}
		}
		for _, gm := range n.GroupModifier {
			if gm == nil {
				continue
			}
			if err := Walk(ctx, gm, fn); err != nil {
				return err
			}
		}
		if n.Func != nil {
			if err := Walk(ctx, n.Func, fn); err != nil {
				return err
			}
		}
	case *Func:
		if n.Args != nil {
			if err := Walk(ctx, n.Args, fn); err != nil {
				return err
			}
		}
	case *Args:
		for _, arg := range n.Arg {
			if arg == nil {
				continue
			}
			if err := Walk(ctx, arg, fn); err != nil {
				return err
			}
		}
	case *Query:
		for _, opt := range n.Options {
			if opt == nil {
				continue
			}
			if err := Walk(ctx, opt, fn); err != nil {
				return err
			}
		}
	case *QueryOption, *Dice, *Modifier, *GroupModifier:
		// no children to walk
	default:
		// unknown types
	}
	return nil
}
