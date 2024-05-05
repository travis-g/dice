package main

import (
	"testing"
)

type ExpressionError struct {
	expression string
	wantErr    bool
}

var expressionParseCases = []ExpressionError{
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
			_ = got
		})
	}
}

func BenchmarkParseReader(b *testing.B) {
	for _, tt := range expressionParseCases {
		b.Run("", func(b *testing.B) {
			got, err := Parse(tt.expression, []byte(tt.expression))
			if (err != nil) != tt.wantErr {
				b.Errorf("ParseReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			_ = got
		})
	}
}
