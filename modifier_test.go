package dice

import (
	"context"
	"testing"
)

var _ = Modifier(&RerollModifier{})
var _ = Modifier(&DropKeepModifier{})

func TestCompareOp_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	var c CompareOp
	tests := []struct {
		name    string
		c       *CompareOp
		args    args
		wantErr bool
	}{
		{"encoded", &c, args{
			[]byte(`"\u003c"`),
		}, false},
		{"unencoded", &c, args{
			[]byte(`"<"`),
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("CompareOp.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDropKeepModifier_Apply(t *testing.T) {
	newTestRoller := func() Roller {
		testRoller, err := NewRollerGroup(&RollerProperties{
			Size:  6,
			Count: 3,
		})
		if err != nil {
			t.Errorf("error creating roller, %v", err)
		}
		return testRoller
	}
	type fields struct {
		Method DropKeepMethod
		Num    int
	}
	type args struct {
		ctx context.Context
		r   Roller
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{{
		name: "drop_0",
		fields: fields{
			Method: DropKeepMethodDropLowest,
			Num:    0,
		},
		args: args{
			ctx: ctx,
			r:   newTestRoller(),
		},
		wantErr: false,
	}, {
		name: "drop_1",
		fields: fields{
			Method: DropKeepMethodDropLowest,
			Num:    1,
		},
		args: args{
			ctx: ctx,
			r:   newTestRoller(),
		},
		wantErr: false,
	}, {
		name: "drop_3",
		fields: fields{
			Method: DropKeepMethodDropLowest,
			Num:    3,
		},
		args: args{
			ctx: ctx,
			r:   newTestRoller(),
		},
		wantErr: false,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk := &DropKeepModifier{
				Method: tt.fields.Method,
				Num:    tt.fields.Num,
			}
			if err := dk.Apply(tt.args.ctx, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("DropKeepModifier.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
			dropped := Filter((tt.args.r).(*RollerGroup).Group, func(r Roller) bool {
				return r.IsDropped(ctx)
			})
			if len(dropped) != dk.Num {
				t.Errorf("DropKeepModifier.Apply() drop mismatch, got %d, want %d", len(dropped), tt.fields.Num)
			}
		})
	}
}
