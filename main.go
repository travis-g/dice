package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/travis-g/draas/dice"
	"github.com/travis-g/draas/server"
)

func toJson(i interface{}) (string, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func main() {
	if os.Args[1] == "serve" {
		exit, _ := server.Run()
		os.Exit(exit)
	}

	roll, err := dice.Parse(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	json, _ := toJson(roll)
	fmt.Printf("%s", json)
}
