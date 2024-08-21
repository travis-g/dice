package dice

import (
	"context"
	"fmt"
)

func ExampleNewRoller() {
	roll, _ := NewRoller(&RollerProperties{
		Type: TypePolyhedron,
		Size: 6,
	})
	die := roll.(*Die)
	fmt.Println(die)
	_ = roll.FullRoll(context.TODO())
	fmt.Println(die)
}
