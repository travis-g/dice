package dice

// package-level variable to prevent optimizations
var i interface{}

// Benchmarks

var diceNotationStrings = []struct {
	notation string
}{
	{"d20"},
	{"1d20"},
}
