package dice

import (
	"fmt"
	"math/rand"
	"testing"
)

// Ensure package sources implement the necessary interfaces.
var _ = (rand.Source64)(&csprngSource{})

// Set of basic range sizes.
var benchmarks = []struct {
	size int
}{
	{6},
	{20},
	{100},
}

func BenchmarkSource_Intn(b *testing.B) {
	b.ReportAllocs()
	for _, bmark := range benchmarks {
		b.Run(fmt.Sprintf("%d", bmark.size), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				Source.Intn(bmark.size)
			}
		})
	}
}

func BenchmarkCryptoIntn(b *testing.B) {
	b.ReportAllocs()
	for _, bmark := range benchmarks {
		b.Run(fmt.Sprintf("%d", bmark.size), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				CryptoIntn(bmark.size)
			}
		})
	}
}

func BenchmarkCryptoInt64(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		CryptoInt64()
	}
}

func Test_quote(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"empty", "", `""`},
		{"simple", "foo", `"foo"`},
		{"with spaces", "foo bar", `"foo bar"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := quote(tt.s); got != tt.want {
				t.Errorf("quote() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_csprngSource_Seed(t *testing.T) {
	tests := []struct {
		name string
		s    *csprngSource
		i    int64
	}{
		{"coverage", &csprngSource{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Seed(tt.i)
		})
	}
}
