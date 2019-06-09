package dice

import (
	"context"
	"fmt"
)

func ExampleNewRoller() {
	ctx := context.Background()
	roll, _ := NewRoller(&RollerProperties{
		Type: TypePolyhedron,
		Size: 6,
	})
	die := roll.(*Die)
	fmt.Println(die)
	_ = roll.Roll(ctx)
	fmt.Println(die)
}
