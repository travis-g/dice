package command

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/ryanuber/columnize"
	yaml "gopkg.in/yaml.v2"
)

var (
	// thematic separator
	delim = `🎲`
)

// generic interface to JSON output function
func toJSON(i interface{}) (string, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func toTable(data map[string]interface{}) (string, error) {
	props := make([]string, 0, len(data)+1)
	if len(data) > 0 {
		keys := make([]string, 0, len(data))
		for k := range data {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := data[k]

			props = append(props, fmt.Sprintf("%s %s %v", k, delim, v))
		}
	}
	str := columnOutput(props, &columnize.Config{
		Delim: delim,
	})
	return str, nil
}

func columnOutput(list []string, c *columnize.Config) string {
	if len(list) == 0 {
		return ""
	}

	if c == nil {
		c = &columnize.Config{}
	}
	if c.Glue == "" {
		c.Glue = "    "
	}
	if c.Empty == "" {
		c.Empty = "n/a"
	}

	return columnize.Format(list, c)
}

func toYaml(data map[string]interface{}) (string, error) {
	tmp, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(tmp), nil
}

func toStruct(iface interface{}) map[string]interface{} {
	var out map[string]interface{}
	tmp, err := json.Marshal(iface)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(tmp, &out)
	return out
}
