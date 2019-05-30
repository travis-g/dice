package dice

import (
	"context"
	"fmt"
)

func ExampleNewRoller() {
	ctx := context.Background()
	roll, _ := NewRoller(&DieProperties{
		Type: TypePolyhedron,
		Size: 6,
	})
	die := roll.(*Die)
	fmt.Println(die)
	_ = roll.Roll(ctx)
	fmt.Println(die)
}
