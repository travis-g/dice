package dice

import (
	"bytes"
	crypto "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"regexp"
	"strings"
)

// Source is the dice package's global PRNG source. Source uses the system's
// native cryptographically secure pseudorandom number generator by default.
//
// Source must be safe for concurrent use: to use something akin to math/rand's
// thread safe global reader try binding a Source64 with a Mutex. See
// math/rand's globalRand variable source code for an example.
var Source = rand.New(&csprngSource{})

// csprngSource is a wrapper for crypto.Reader that implements
// rand.Source64.
type csprngSource struct{}

// Seed is a noop; a csprngSource does not need to be seeded.
func (s *csprngSource) Seed(int64) {
	// noop; system CSPRNG cannot be seeded
}

func (s *csprngSource) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}

// Uint64 satisfies the rand.Source64 interface.
func (s *csprngSource) Uint64() (u uint64) {
	err := binary.Read(crypto.Reader, binary.BigEndian, &u)
	if err != nil {
		panic(err)
	}
	return
}

// CryptoInt64 is a convenience function that returns a cryptographically random
// int64 using the system's CSPRNG. If there is a problem generating enough
// entropy it will return a non-nil error.
//
// This function was designed to seed math/rand Sources with uniform random
// values. It will not use the package's global Source.
func CryptoInt64() (int64, error) {
	i, err := crypto.Int(crypto.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return i.Int64(), err
	}
	return i.Int64(), nil
}

// CryptoIntn is a convenience wrapper for emulating rand.CryptoIntn using crypto/rand.
// Rather than panicking if max <= 0, if max <= 0 an ErrSizeZero error is
// returned and n will be 0. Any other errors encountered when generating the
// integer are passed through by err.
//
// CryptoIntn does not use the package's global Source.
func CryptoIntn(max int) (n int, err error) {
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

// FindNamedCaptureGroups finds string submatches within an input string based
// on a compiled Regexp and returns a map of the named capture groups with their
// captured submatches.
func FindNamedCaptureGroups(exp *regexp.Regexp, in string) map[string]string {
	submatches := exp.FindStringSubmatch(in)

	captures := make(map[string]string)
	for i, name := range exp.SubexpNames() {
		if i != 0 && name != "" {
			captures[name] = submatches[i]
		}
	}
	return captures
}

func jsonEncode(in interface{}) []byte {
	if in == nil {
		return nil
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(in); err != nil {
		return nil
	}
	return buf.Bytes()
}
