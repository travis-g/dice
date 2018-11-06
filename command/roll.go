package command

import (
	"fmt"

	"github.com/travis-g/draas/dice"
)

func RollCommand(roll string) error {
	dice, err := dice.Parse(roll)
	if err != nil {
		return err
	}
	fmt.Println(dice.Result)
	return nil
}
