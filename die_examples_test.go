package dice

import "fmt"

func ExampleDie_output() {
	die := &Die{
		Size: 20,
	}
	fateDie := &Die{
		Type: TypeFudge,
		Size: 1,
	}
	fmt.Println(die, fateDie)
	// Output: d20 dF
}

func ExampleNewDie_output() {
	die, _ := NewDie(&DieProperties{
		Type: TypePolyhedron,
		Size: 6,
	})
	fmt.Print(die)
	// Output: d6
}
