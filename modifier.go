package dice

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
)

type CompareOp string

const (
	CompareEquals  = "="
	CompareLess    = "<"
	CompareGreater = ">"
)

type Modifier interface {
	Apply(context.Context, *PolyhedralDie) error
	fmt.Stringer
}

var _ = Modifier(&RerollModifier{})

type RerollModifier struct {
	Compare string `json:"compare"`
	Point   int    `json:"point"`
}

func (m *RerollModifier) String() string {
	var buf bytes.Buffer
	buf.WriteString("r")
	// concise output if checking equals
	if m.Compare != "=" {
		buf.WriteString(m.Compare)
	}
	buf.WriteString(strconv.Itoa(m.Point))
	return buf.String()
}

func (m *RerollModifier) Apply(ctx context.Context, d *PolyhedralDie) error {
	switch m.Compare {
	case "", CompareEquals:
		for d.Result == float64(m.Point) {
			d.Unrolled = true
			d.Roll(ctx)
		}
	case CompareLess:
		for d.Result <= float64(m.Point) {
			d.Unrolled = true
			d.Roll(ctx)
		}
	case CompareGreater:
		for d.Result > float64(m.Point) {
			d.Unrolled = true
			d.Roll(ctx)
		}
	}
	return nil
}
