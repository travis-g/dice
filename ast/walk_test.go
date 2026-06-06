package main

import (
	"context"
	"reflect"
	"testing"
)

func typeName(n Node) string {
	if n == nil {
		return "<nil>"
	}
	return reflect.TypeOf(n).String()
}

func TestWalk_Traversal(t *testing.T) {
	root, err := ParseString("1 + 2 * (3 + 4)", false)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	visited := []string{}
	visit := func(ctx context.Context, n Node, err error) error {
		if err != nil {
			return err
		}
		visited = append(visited, typeName(n))
		return nil
	}

	if err := Walk(context.Background(), root, visit); err != nil {
		t.Fatalf("Walk() returned error: %v", err)
	}

	if len(visited) == 0 {
		t.Fatal("expected nodes to be visited")
	}
}
