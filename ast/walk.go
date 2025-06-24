package main

import "context"

func Walk(ctx context.Context, n *Node, fn WalkFunc) error {
	if err := fn(ctx, n); err != nil {
		return err
	}
	return ErrNotImplemented
}

type WalkFunc func(ctx context.Context, n *Node) error
