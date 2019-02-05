package command

import (
	"fmt"

	"github.com/urfave/cli"
)

// Field extracts a field from a given interface.
func Field(i interface{}, field string) interface{} {
	var val interface{}
	if data, ok := i.(map[string]interface{}); ok {
		val = data[field]
	}
	return val
}

// Output prints an interface based on the desired format.
func Output(c *cli.Context, i interface{}) (string, error) {
	data, err := toMapStringInterface(i)
	if err != nil {
		return "", err
	}
	switch format := c.String("format"); format {
	// TODO(travis-g): use i.String() output for unspecified format
	case "table":
		return toTable(data)
	case "json":
		return toJSON(data)
	case "yaml", "yml":
		return toYaml(data)
	default:
		return "", fmt.Errorf("requested format %v unhandled", err)
	}
}

// OutputField prints a given field from a provided interface using a provided
// context's format.
func OutputField(c *cli.Context, i interface{}, field string) (string, error) {
	data, err := toMapStringInterface(i)
	if err != nil {
		return "", err
	}
	return Output(c, data[field])
}
