package dice

import (
	"fmt"
	"math/rand"
	"testing"
)

// Ensure csprngSource statisfies the rand.Source64 interface; rand.Source64
// sources use half the entropy of a regular rand.Source.
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

func BenchmarkIntn(b *testing.B) {
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
