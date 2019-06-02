package dice

import (
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
		{"improper", &c, args{
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
