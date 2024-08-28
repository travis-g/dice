package main

import "context"

func Walk(ctx context.Context, n *Node, fn WalkFunc) error {
	panic(ErrNotImplemented)
}

type WalkFunc func(ctx context.Context, n *Node, err error) error
