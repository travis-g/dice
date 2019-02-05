package dice

import (
	crand "crypto/rand"
	"math/big"
	rand "math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	// seed PRNG with crypto-secure integer
	// rand.Seed(cryptoSeed())
}

func cryptoSeed() int64 {
	seed, err := crand.Int(crand.Reader, big.NewInt(int64(^uint64(0)>>1)))
	if err != nil {
		panic(err)
	}
	return seed.Int64()
}

func quote(s string) string {
	return "\"" + s + "\""
}
