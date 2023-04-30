package dice

import (
	"context"
	"fmt"
	"testing"
)

var _ Roller = (*RollerGroup)(nil)

func newInt(i int) *int {
	return &i
}

func newFloat(f float64) *float64 {
	return &f
}

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
				&Die{Result: &Result{Value: 2.0}},
				&Die{Result: &Result{Value: 3.0}},
				&Die{Result: &Result{Value: 4.0}},
			},
			want: 9,
		},
		{
			name: "nested",
			g: Group{
				&Die{Result: &Result{Value: 2.0}},
				&Group{
					&Die{Result: &Result{Value: 3.0}},
				},
				&Die{Result: &Result{Value: 4.0}},
			},
			want: 9,
		},
		{
			name: "dropped",
			g: Group{
				&Die{Result: &Result{Value: 2.0, Dropped: true}},
				&Die{Result: &Result{Value: 4.0}},
			},
			want: 4,
		},
		{
			name: "mixed",
			g: Group{
				&Die{Result: &Result{Value: 2.0, Dropped: true}},
				&Die{Result: &Result{Value: 1}},
				&Die{Type: TypeFudge, Result: &Result{Value: -1}},
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
				&Die{Result: &Result{Value: 2}},
				&Die{Result: &Result{Value: 3}},
				&Die{Result: &Result{Value: 4}},
			},
			want: "2+3+4",
		},
		{
			name: "unrolled",
			g: Group{
				&Die{Size: 3},
				&Die{Result: &Result{Value: 3, Dropped: false}},
				&Die{Result: &Result{Value: 4, Dropped: false}},
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

func TestRollerGroup_FullRoll(t *testing.T) {
	tests := []struct {
		name       string
		d          *RollerGroup
		wantResult float64
		wantErr    bool
	}{
		{
			name: "basic",
			d: MustNewRollerGroup(&RollerProperties{
				Count: 4,
				Size:  6,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.FullRoll(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("RollerGroup.FullRoll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
