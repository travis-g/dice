package dice

import (
	crypto "crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strings"
)

// Source is the dice package's global PRNG source.
var Source *rand.Rand

// source is a wrapper for crypto.Reader that implements rand.Source
type source struct{}

func (s *source) Seed(int64) {
	// noop
}

func (s *source) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}

func (s *source) Uint64() (u uint64) {
	err := binary.Read(crypto.Reader, binary.BigEndian, &u)
	if err != nil {
		panic(err)
	}
	return
}

func init() {
	// seed global math/rand instance
	seed, err := CryptoInt64()
	if err != nil {
		panic(err)
	}
	rand.Seed(seed)
	// set the global package Source to a rand.Rand that sources crypto.Reader
	Source = rand.New(&source{})
}

// CryptoInt64 is a convenience function that returns a cryptographically random
// int64. If there is a problem generating enough entropy it will return a
// non-nil error.
//
// This function was designed to seed math/rand with a uniform random value.
func CryptoInt64() (int64, error) {
	i, err := crypto.Int(crypto.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return i.Int64(), err
	}
	return i.Int64(), nil
}

// Intn is a convenience wrapper for emulating rand.Intn using crypto/rand.
// Rather than panicking if max <= 0, if max <= 0 an ErrSizeZero error is
// returned and n will be 0. Any other errors encountered when generating the
// integer are passed through by err.
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
