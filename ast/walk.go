package main

import "context"

// A Node abstracts any leaf in an expression AST.
type Node interface {
	// Eval traverses any sub-Nodes and returns a result.
	Eval(ctx context.Context) error
	// Resolve traverses any sub-Nodes and ensures all required data is fetched,
	// such as query parameters.
	Resolve(ctx context.Context) (*Node, error)
}

func Walk(ctx context.Context, n *Node, fn WalkFunc) error {
	if err := fn(ctx, n); err != nil {
		return err
	}
	return ErrNotImplemented
}

type WalkFunc func(ctx context.Context, n *Node) error
