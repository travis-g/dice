package dice

import (
	crypto "crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

// CryptoInt64 is a convenience function that returns a cryptographically random
// int64. If there is a problem generating enough entropy it will return a
// non-nil error.
func CryptoInt64() (int64, error) {
	i, err := crypto.Int(crypto.Reader, big.NewInt(int64(^uint64(0)>>1)))
	if err != nil {
		return i.Int64(), err
	}
	return i.Int64(), nil
}

// Intn is a convenience wrapper for emulating rand.Intn using crypto/rand. As
// crypto.Int panics if max <= 0, if size is <= 0 an ErrSizeZero error is
// returned and n will be 0. If there is a problem generating enough entropy it
// will return a non-nil error.
func Intn(size int) (n int, err error) {
	if size <= 0 {
		n = 0
		err = &ErrSizeZero{}
		return
	}
	bigInt, err := crypto.Int(crypto.Reader, big.NewInt(int64(size)))
	n = int(bigInt.Int64())
	return
}

func quote(s string) string {
	return strings.Join([]string{"\"", s, "\""}, "")
}

func expression(i ...interface{}) string {
	raw := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(i...)), "+"), "[]")
	return strings.Replace(raw, "+-", "-", -1)
}

// All returns true if all dice interfaces of a slice match a predicate.
func All(vs []*Interface, f func(*Interface) bool) bool {
	for _, v := range vs {
		if !f(v) {
			return false
		}
	}
	return true
}
