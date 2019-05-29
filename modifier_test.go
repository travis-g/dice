package dice

import "testing"

var _ = Modifier(&RerollModifier{})

func TestRerollModifier_String(t *testing.T) {
	type fields struct {
		Compare string
		Point   int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"CompareEquals2",
			fields{"=", 2},
			"r2"},
		{"CompareEqualsImplicit2",
			fields{"", 2},
			"r2"},
		{"CompareLess3",
			fields{"<", 3},
			"r<3"},
		{"CompareGreater3",
			fields{">", 3},
			"r>3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &RerollModifier{
				Compare: tt.fields.Compare,
				Point:   tt.fields.Point,
			}
			if got := m.String(); got != tt.want {
				t.Errorf("RerollModifier.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
