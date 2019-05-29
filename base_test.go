package dice

import (
	"fmt"
	"math/rand"
	"testing"
)

var _ = (rand.Source64)(&source{})

func BenchmarkIntn(b *testing.B) {
	b.ReportAllocs()
	benchmarks := []struct {
		size int
	}{
		{6},
		{20},
		{100},
	}
	for _, bmark := range benchmarks {
		b.Run(fmt.Sprintf("%d", bmark.size), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				Intn(bmark.size)
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
