package dice

import "fmt"

func ExampleNewRoller() {
	die, _ := NewRoller(&DieProperties{
		Type: TypePolyhedron,
		Size: 6,
	})
	fmt.Print(die)
	// Output: d6
}
