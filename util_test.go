package dice

import (
	"fmt"
	"testing"
)

func BenchmarkIntn(b *testing.B) {
	b.ReportAllocs()
	benchmarks := []struct {
		size int
	}{
		{6},
		{8},
		{10},
		{12},
		{20},
		{50},
		{100},
	}
	var rand int
	for _, bmark := range benchmarks {
		b.Run(fmt.Sprintf("%d", bmark.size), func(b *testing.B) {
			var err error

			for n := 0; n < b.N; n++ {
				rand, err = Intn(bmark.size)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
	i = rand
}

func BenchmarkCryptoInt64(b *testing.B) {
	b.ReportAllocs()
	var (
		rand int64
		err  error
	)
	for n := 0; n < b.N; n++ {
		rand, err = CryptoInt64()
		if err != nil {
			b.Fatal(err)
		}
	}
	i = rand
}
