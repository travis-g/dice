package dice

import (
	"context"
	"fmt"
	"testing"
)

var _ Roller = (*RollerGroup)(nil)

var groupProperties = []struct {
	name  string
	props RollerProperties
}{{"3d6",
	RollerProperties{
		Type:  TypePolyhedron,
		Size:  6,
		Count: 3,
	},
}, {"3dF",
	RollerProperties{
		Type:  TypeFudge,
		Count: 3,
	},
}, {"3f3",
	RollerProperties{
		Type:  TypeFudge,
		Size:  3,
		Count: 3,
	},
},
}

var ctx = context.Background()

func BenchmarkNewRollerGroup(b *testing.B) {
	for _, bench := range groupProperties {
		b.Run(fmt.Sprintf("%s", bench.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				NewRollerGroup(&bench.props)
			}
		})
	}
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
				&Die{Results: []*Result{{Value: 2}}},
				&Die{Results: []*Result{{Value: 3}}},
				&Die{Results: []*Result{{Value: 4}}},
			},
			want: 9,
		},
		{
			name: "nested",
			g: Group{
				&Die{Results: []*Result{{Value: 2}}},
				&Group{
					&Die{Results: []*Result{{Value: 3}}},
				},
				&Die{Results: []*Result{{Value: 4}}},
			},
			want: 9,
		},
		{
			name: "dropped",
			g: Group{
				&Die{Results: []*Result{{Value: 2, Dropped: true}}},
				&Die{Results: []*Result{{Value: 4}}},
			},
			want: 4,
		},
		{
			name: "mixed",
			g: Group{
				&Die{Results: []*Result{{Value: 2, Dropped: true}}},
				&Die{Results: []*Result{{Value: 1}}},
				&Die{Type: TypeFudge, Results: []*Result{{Value: -1}}},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, err := tt.g.Total(ctx)
			if err != nil {
				t.Errorf("Got error on %v: %v", tt, err)
			}
			if got := total; got != tt.want {
				t.Errorf("Group.Total(%s) = %v, want %v", tt.name, got, tt.want)
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
			name: "empty",
			g:    Group{},
			want: "0",
		},
		{
			name: "basic",
			g: Group{
				&Die{Results: []*Result{{Value: 2}}},
				&Die{Results: []*Result{{Value: 3}}},
				&Die{Results: []*Result{{Value: 4}}},
			},
			want: "2+3+4",
		},
		{
			name: "unrolled",
			g: Group{
				&Die{Size: 3},
				&Die{Results: []*Result{{Value: 3, Dropped: false}}},
				&Die{Results: []*Result{{Value: 4, Dropped: false}}},
			},
			want: "d3+3+4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.Expression(); got != tt.want {
				t.Errorf("Group.Expression(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
