package main

import (
	"os"
	"testing"
)

func Test_main(t *testing.T) {
	tmp := os.Args // initial args
	tests := []struct {
		name string
		args []string
	}{
		{"blank", []string{}},
		{"base", []string{"2d20 + 1 # test"}},
		{"trace", []string{"-trace", "1d20"}},
		{"diagram", []string{"-diagram"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tmp
			for _, arg := range tt.args {
				os.Args = append(os.Args, arg)
			}
			main()
		})
	}
}
