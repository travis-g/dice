package dice

import (
	"context"
	"reflect"
	"testing"
)

func TestParseNotation(t *testing.T) {
	tests := []struct {
		name     string
		notation string
		want     RollerProperties
		wantErr  bool
	}{
		{
			name:     "basic",
			notation: "2d20",
			want: RollerProperties{
				Type:           TypePolyhedron,
				Count:          2,
				Size:           20,
				DieModifiers:   ModifierList{},
				GroupModifiers: ModifierList{},
			},
		},
		{
			name:     "capitalized",
			notation: "D6",
			want: RollerProperties{
				Type:           TypePolyhedron,
				Count:          1,
				Size:           6,
				DieModifiers:   ModifierList{},
				GroupModifiers: ModifierList{},
			},
		},
		{
			name:     "fudge",
			notation: "2dF",
			want: RollerProperties{
				Type:           TypeFudge,
				Count:          2,
				DieModifiers:   ModifierList{},
				GroupModifiers: ModifierList{},
			},
		},
		{
			name:     "weird-caps",
			notation: "Df",
			want: RollerProperties{
				Type:           TypeFudge,
				Count:          1,
				DieModifiers:   ModifierList{},
				GroupModifiers: ModifierList{},
			},
		},
		{
			name:     "keep-1",
			notation: "2d20k1",
			want: RollerProperties{
				Type:         TypePolyhedron,
				Count:        2,
				Size:         20,
				DieModifiers: ModifierList{},
				GroupModifiers: []Modifier{
					&DropKeepModifier{DropKeepMethodKeep, 1},
				},
			},
		},
		{
			name:     "keep-implied-1",
			notation: "2d20k",
			want: RollerProperties{
				Type:         TypePolyhedron,
				Count:        2,
				Size:         20,
				DieModifiers: ModifierList{},
				GroupModifiers: []Modifier{
					&DropKeepModifier{DropKeepMethodKeep, 1},
				},
			},
		},
		{
			name:     "keep-lowest-implied-1",
			notation: "2d20kl",
			want: RollerProperties{
				Type:         TypePolyhedron,
				Count:        2,
				Size:         20,
				DieModifiers: ModifierList{},
				GroupModifiers: []Modifier{
					&DropKeepModifier{DropKeepMethodKeepLowest, 1},
				},
			},
		},
		{
			name:     "drop-1",
			notation: "2d20d1",
			want: RollerProperties{
				Type:         TypePolyhedron,
				Count:        2,
				Size:         20,
				DieModifiers: ModifierList{},
				GroupModifiers: []Modifier{
					&DropKeepModifier{DropKeepMethodDrop, 1},
				},
			},
		},
		{
			name:     "drop-implied-1",
			notation: "2D20d",
			want: RollerProperties{
				Type:         TypePolyhedron,
				Count:        2,
				Size:         20,
				DieModifiers: ModifierList{},
				GroupModifiers: []Modifier{
					&DropKeepModifier{DropKeepMethodDrop, 1},
				},
			},
		},
		{
			name:     "drop-highest-implied-1",
			notation: "2d20dh",
			want: RollerProperties{
				Type:         TypePolyhedron,
				Count:        2,
				Size:         20,
				DieModifiers: ModifierList{},
				GroupModifiers: []Modifier{
					&DropKeepModifier{DropKeepMethodDropHighest, 1},
				},
			},
		},
		{
			name:     "reroll-once-1",
			notation: "3d6ro1r>3",
			want: RollerProperties{
				Type:  TypePolyhedron,
				Count: 3,
				Size:  6,
				DieModifiers: ModifierList{
					&RerollModifier{Once: true, CompareTarget: &CompareTarget{EMPTY, 1}},
					&RerollModifier{CompareTarget: &CompareTarget{GTR, 3}},
				},
				GroupModifiers: ModifierList{},
			},
		},
		{
			name:     "valid-before-junk",
			notation: "3d6sabcxyz3",
			want: RollerProperties{
				Type:         TypePolyhedron,
				Count:        3,
				Size:         6,
				DieModifiers: ModifierList{},
				GroupModifiers: ModifierList{
					&SortModifier{SortDirectionAscending},
				},
			},
		},
		{
			name:     "junk",
			notation: "3d6abcxyz3sa",
			want: RollerProperties{
				Type:           TypePolyhedron,
				Count:          3,
				Size:           6,
				DieModifiers:   ModifierList{},
				GroupModifiers: ModifierList{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNotation(context.Background(), tt.notation)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNotation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseNotation() = %v, want %v", got, tt.want)
			}
		})
	}
}
