package dice

import (
	"testing"
)

var groupSets = []Group{
	Group{
		&Die{
			Size:     6,
			Unrolled: true,
		},
	},
}

func TestGroup_Total(t *testing.T) {
	tests := []struct {
		name string
		g    Group
		want float64
	}{
		{
			name: "basic",
			g: Group{
				&Die{Result: 2},
				&Die{Result: 3},
				&Die{Result: 4},
			},
			want: 9,
		},
		{
			name: "nested",
			g: Group{
				&Die{Result: 2},
				&Group{
					&Die{Result: 3},
				},
				&Die{Result: 4},
			},
			want: 9,
		},
		{
			name: "dropped",
			g: Group{
				&Die{Result: 2, Dropped: true},
				&Die{Result: 4},
			},
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.Total(); got != tt.want {
				t.Errorf("Group.Total() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGroup_Expression(t *testing.T) {
	tests := []struct {
		name string
		g    Group
		want string
	}{
		{
			name: "basic",
			g: Group{
				&Die{Result: 2},
				&Die{Result: 3},
				&Die{Result: 4},
			},
			want: "2+3+4",
		},
		{
			name: "unrolled",
			g: Group{
				&Die{Type: "d3", Unrolled: true},
				&Die{Result: 3},
				&Die{Result: 4},
			},
			want: "d3+3+4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.Expression(); got != tt.want {
				t.Errorf("Group.Expression() = %v, want %v", got, tt.want)
			}
		})
	}
}
