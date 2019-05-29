package dice

import "fmt"

func ExampleDie() {
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
