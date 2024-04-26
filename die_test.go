package dice

import (
	"context"
	"testing"
)

// ensure Die implements Roller
var _ Roller = (*Die)(nil)

func TestDie_Roll(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		d       *Die
		args    args
		wantErr bool
	}{
		{
			name: "0-sided",
			d: MustNewDie(&RollerProperties{
				Type:  TypePolyhedron,
				Size:  0,
				Count: 1,
			}).(*Die),
			args: args{context.TODO()},
		},
		{
			name: "1-sided",
			d: MustNewDie(&RollerProperties{
				Type:  TypePolyhedron,
				Size:  1,
				Count: 1,
			}).(*Die),
			args: args{context.TODO()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Roll(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Die.Roll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
