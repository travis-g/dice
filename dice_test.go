package dice_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	"github.com/travis-g/dice"
)

// CustomDie is a die with a custom set of faces.
type CustomDie struct {
	Faces  []float64
	Result float64

	// atomic; track whether CustomDie was rolled
	rolled int32
}

// Roll will return one of the indices of Faces. If the die has been rolled
// before an error is returned.
func (c *CustomDie) Roll(ctx context.Context) error {
	if ok := atomic.CompareAndSwapInt32(&c.rolled, 0, 1); !ok {
		return dice.ErrRolled
	}
	c.Result = c.Faces[dice.Source.Intn(len(c.Faces))]
	return nil
}

// Total returns the result of rolling the die, if it's been rolled. Otherwise,
// an error is returned.
func (c *CustomDie) Total(ctx context.Context) (float64, error) {
	if atomic.LoadInt32(&c.rolled) == 0 {
		return 0, dice.ErrUnrolled
	}
	return c.Result, nil
}

// String implements fmt.Stringer so that the die can be printed.
func (c *CustomDie) String() string {
	if atomic.LoadInt32(&c.rolled) == 0 {
		var sides []string
		for _, s := range c.Faces {
			sides = append(sides, fmt.Sprintf("%v", s))
		}
		return fmt.Sprintf("{%v}", strings.Join(sides, ","))
	}
	return fmt.Sprintf("%v", c.Result)
}

// Ensure that CustomDie implements dice.Roller
var _ dice.Roller = (*CustomDie)(nil)

func Example() {
	ctx := context.Background()

	// customDie can only roll 6s.
	customDie := &CustomDie{
		Faces: []float64{6, 6, 6, 6, 6, 6},
	}

	fmt.Fprintln(os.Stdout, customDie)
	err := customDie.Roll(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	result, err := customDie.Total(ctx)
	fmt.Fprintln(os.Stdout, result)
	// Output:
	// {6,6,6,6,6,6}
	// 6
}
