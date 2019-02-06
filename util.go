package dice

import (
	crand "crypto/rand"
	"math/big"
	rand "math/rand"
	"time"
)

func init() {
	// seed math/rand with the time by default
	Seed(time.Now().UTC().UnixNano())
}

// Seed is a convenience function to re-seed math/rand.
func Seed(seed int64) {
	rand.Seed(seed)
}

// CryptoInt64 is a convenience function that returns a cryptographically random
// int64. If there is a problem generating enough entropy it will return a
// non-nil error.
func CryptoInt64() (int64, error) {
	i, err := crand.Int(crand.Reader, big.NewInt(int64(^uint64(0)>>1)))
	if err != nil {
		return 0, err
	}
	return i.Int64(), nil
}

func quote(s string) string {
	return "\"" + s + "\""
}
