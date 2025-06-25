package main

import (
	"context"
	"testing"
)

func TestWalk(t *testing.T) {
	type args struct {
		ctx  context.Context
		root *Node
		fn   WalkFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"nil root", args{context.TODO(), nil, func(ctx context.Context, n *Node, err error) error { return nil }}, false},
		// TODO: Add more test cases for Walk
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Walk(tt.args.ctx, tt.args.root, tt.args.fn); (err != nil) != tt.wantErr {
				t.Errorf("Walk() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
