package command

import (
	"fmt"
	"strings"

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
	// TODO(travis-g): output a specific field only if c.String("field")
	data, err := toMapStringInterface(i)
	if err != nil {
		return "", err
	}
	switch format := strings.ToLower(c.String("format")); format {
	case "":
		return fmt.Sprintf("%s", i), nil
	case "table":
		return toTable(data)
	case "json":
		return toJSON(data)
	case "yaml", "yml":
		return toYaml(data)
	case "gostring":
		return toGoString(i)
	case "gv", "graphviz", "dot":
		return toGraphviz(i)
	default:
		return "", fmt.Errorf("requested format %v unhandled", format)
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
