package runtime

import (
	"strings"
)

type nodejs struct {
	Config
}

func (n nodejs) SanitizeDependencies(dependencies string) string {
	result := "{}"
	if strings.Trim(dependencies, " ") != "" {
		result = dependencies
	}

	return result
}

var _ Runtime = nodejs{}
