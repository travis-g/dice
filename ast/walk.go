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
	return fn(ctx, root, nil)
}
