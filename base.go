package dice

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

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

// Intn is a convenience wrapper for emulating rand.Intn using crypto/rand. Panics if size <= 0.
func Intn(size int) (int, error) {
	bigInt, err := crand.Int(crand.Reader, big.NewInt((int64)(size)))
	return (int)(bigInt.Int64()), err
}

func quote(s string) string {
	return strings.Join([]string{"\"", s, "\""}, "")
}

func expression(i ...interface{}) string {
	raw := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(i...)), "+"), "[]")
	return strings.Replace(raw, "+-", "-", -1)
}

// All returns true if all elements of a slice match a predicate.
func All(vs []string, f func(string) bool) bool {
	for _, v := range vs {
		if !f(v) {
			return false
		}
	}
	return true
}
