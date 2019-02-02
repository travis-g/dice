package dice

import (
	mathrand "math/rand"
	"time"
)

func init() {
	// seed PRNG
	mathrand.Seed(time.Now().UTC().UnixNano())
}

func quote(s string) string {
	return "\"" + s + "\""
}
