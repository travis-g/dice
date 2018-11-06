package command

import "encoding/json"

// generic interface to JSON output function
func toJson(i interface{}) (string, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
