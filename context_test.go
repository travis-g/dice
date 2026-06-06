package dice

import "testing"

func Test_contextKey_String(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		want string
	}{
		{"test", "dice context value test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := contextKey(tt.name)
			got := k.String()

			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
