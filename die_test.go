package dice

import (
	"fmt"
	"testing"
)

var dieSets = []struct {
	size  int
	count uint
}{
	{6, 2},
	{20, 2},
	{100, 2},
}

func BenchmarkNewDie(b *testing.B) {
	b.ReportAllocs()
	for _, tc := range dieSets {
		b.Run(fmt.Sprintf("d%d", tc.size), func(b *testing.B) {
			size := tc.size
			for n := 0; n < b.N; n++ {
				NewDie(size)
			}
		})
	}
}

func BenchmarkNewDieSet(b *testing.B) {
	b.ReportAllocs()
	for _, tc := range dieSets {
		b.Run(fmt.Sprintf("%dd%d", tc.count, tc.size), func(b *testing.B) {
			size := tc.size
			count := tc.count
			for n := 0; n < b.N; n++ {
				NewDieSet(size, count)
			}
		})
	}
}
