package dice

import (
	crypto "crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

type source struct{}

// CryptoInt64 is a convenience function that returns a cryptographically random
// int64. If there is a problem generating enough entropy it will return a
// non-nil error.
//
// This function was designed to seed math/rand, if ever necessary.
func CryptoInt64() (int64, error) {
	i, err := crypto.Int(crypto.Reader, big.NewInt(int64(^uint64(0)>>1)))
	if err != nil {
		return i.Int64(), err
	}
	return i.Int64(), nil
}

// Intn is a convenience wrapper for emulating rand.Intn using crypto/rand.
// Rather than panicking if max <= 0 as crypt.Int would, if max <= 0 an
// ErrSizeZero error is returned and n will be 0. Any other errors encountered
// when generating the integer are passed through by err.
func Intn(max int) (n int, err error) {
	if max <= 0 {
		n = 0
		err = ErrSizeZero
		return
	}
	bigInt, err := crypto.Int(crypto.Reader, big.NewInt(int64(max)))
	n = int(bigInt.Int64())
	return
}

// quote returns the input string wrapped within quotation marks.
func quote(s string) string {
	return strings.Join([]string{"\"", s, "\""}, "")
}

// expression creates a math expression from an arbitrary set of interfaces,
// simplifying the result using the commutative property of addition.
func expression(i ...interface{}) string {
	raw := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(i...)), "+"), "[]")
	return strings.Replace(raw, "+-", "-", -1)
}
