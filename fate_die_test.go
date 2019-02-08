package dice

import "testing"

func BenchmarkNewFateDie(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		NewFateDie()
	}
}
