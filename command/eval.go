package command

import (
	"fmt"
	"os"

	"github.com/travis-g/draas/dice/math"
)

func EvalCommand(eval string) error {
	exp, err := math.Eval(eval)
	if err != nil {
		return err
	}
	fmt.Println(exp.Result)
	json, err := toJson(exp)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, string(json))
	return nil
}
