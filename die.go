package dice

import (
	"fmt"
)

var _ = Interface(&Die{})

// A Die represents a variable-sided die in memory, including the result of
// rolling it.
type Die struct {
	Interface `json:"self,omitempty"`
	Type      string  `json:"type"`
	Result    float64 `json:"result"`
	Size      int     `json:"size"`
	Dropped   bool    `json:"dropped,omitempty"`
	Unrolled  bool    `json:"unrolled,omitempty"`
}

// String returns an expression-like representation of a rolled Die or the kind
// of die if it has not been rolled.
func (d *Die) String() string {
	if !d.Unrolled {
		return fmt.Sprintf("%v", d.Total())
	}
	return d.Type
}

// GoString prints the Go syntax of a Die.
func (d *Die) GoString() string {
	return fmt.Sprintf("%#v", *d)
}

// Total returns the result of a die. If dropped, 0 is returned.
func (d *Die) Total() float64 {
	if d.Dropped {
		return 0.0
	}
	if d.Unrolled {
		d.Roll()
	}
	return d.Result
}

// Roll will Roll a given Die (if unrolled) and set the die's result. Results
// are in the range [1, size]. If the die already has a result it will not be
// rerolled.
func (d *Die) Roll() (float64, error) {
	if !d.Unrolled {
		return d.Result, nil
	}
	if d.Result == 0 {
		i, err := Intn(d.Size)
		if err != nil {
			return 0, err
		}
		d.Result = (float64)(1 + i)
		d.Unrolled = false
	}
	return d.Result, nil
}
