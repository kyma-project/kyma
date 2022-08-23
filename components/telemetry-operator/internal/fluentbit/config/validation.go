package config

import (
	"fmt"
	"strings"
)

func ParseCustomSection(section string) (ParameterList, error) {
	var params ParameterList
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, " ")
		if !found {
			return nil, fmt.Errorf("invalid line: %s", line)
		}
		params.Add(Parameter{
			Key:   strings.ToLower(strings.TrimSpace(key)),
			Value: strings.TrimSpace(value),
		})
	}
	return params, nil
}
