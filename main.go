package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/travis-g/draas/dice"
)

func toJson(i interface{}) (string, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func main() {
	roll, err := dice.Parse(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	json, _ := toJson(roll)
	fmt.Printf("%s", json)
}
