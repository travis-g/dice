package dice

import (
	"context"
	"testing"
)

// ensure Die implements Roller
var _ Roller = (*Die)(nil)

func TestDie_Roll(t *testing.T) {
	tests := []struct {
		name    string
		d       *Die
		result  *int
		wantErr bool
	}{
		{
			name: "0-sided",
			d: Must(NewDie(&RollerProperties{
				Type:  TypePolyhedron,
				Size:  0,
				Count: 1,
			})).(*Die),
			result:  ptr(0),
			wantErr: false,
		},
		{
			name: "1-sided",
			d: Must(NewDie(&RollerProperties{
				Type:  TypePolyhedron,
				Size:  1,
				Count: 1,
			})).(*Die),
			result:  ptr(1),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Roll(context.TODO()); (err != nil) != tt.wantErr {
				t.Errorf("Die.Roll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
