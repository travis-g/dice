package main

import (
	"testing"
)

var unoptimize any

type ExpressionWantError struct {
	expression string
	wantErr    bool
}

var expressionParseCases = []ExpressionWantError{
	{"1d20", false},
	{"d20", false},
	{"1d20+1", false},
	{"1+1", false},
	{"1+1d20", false},
	{"3 + 5 * 2", false},
	{"(3 + 5) * 2", false},
	{"8 / 2 * 4", false},
	{"8 / (2 * 4)", false},
	{"4 + 6 / 2", false},
	{"(4 + 6) / 2", false},
	{"2 * 3**2", false},
	{"(2 * 3)**2", false},
	{"10 - 3 + 2", false},
	{"10 - (3 + 2)", false},
}

func TestParse(t *testing.T) {
	for _, tt := range expressionParseCases {
		t.Run(tt.expression, func(t *testing.T) {
			got, err := Parse(tt.expression, []byte(tt.expression))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			unoptimize = got.(Expression)
		})
	}
}

func BenchmarkParse(b *testing.B) {
	for _, tt := range expressionParseCases {
		b.Run("", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				i, err := Parse(tt.expression, []byte(tt.expression))
				if (err != nil) != tt.wantErr {
					b.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				unoptimize = i
			}
		})
	}
}
