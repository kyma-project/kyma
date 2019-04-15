package file

import (
	"github.com/pkg/errors"
	"regexp"
)

func Filter(paths []string, filter string) ([]string, error) {
	if filter == "" {
		return paths, nil
	}

	filtered := []string{}
	regex, err := regexp.Compile(filter)
	if err != nil {
		return nil, errors.Wrapf(err, "while compiling path filter regex")
	}
	for _, value := range paths {
		if regex.MatchString(value) {
			filtered = append(filtered, value)
		}
	}
	return filtered, nil
}
