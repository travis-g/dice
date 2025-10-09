package main

import (
	"context"
	"testing"
)

func TestWalk_String(t *testing.T) {
	stringFunc := func(ctx context.Context, n *Node, err error) error {
		if n == nil {
			return nil
		}
		if s, ok := any(n).(interface{ String() string }); ok {
			_ = s.String()
			return nil
		}
		return nil
	}

	type args struct {
		ctx  context.Context
		root *Node
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"nil root", args{context.TODO(), nil}, "", false},
		// TODO: Add more test cases for Walk
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Walk(tt.args.ctx, tt.args.root, stringFunc); (err != nil) != tt.wantErr {
				t.Errorf("Walk() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
