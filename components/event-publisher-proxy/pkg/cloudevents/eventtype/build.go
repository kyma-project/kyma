package eventtype

import (
	"fmt"
	"strings"
)

func build(prefix, applicationName, event, version string) string {
	if len(strings.TrimSpace(prefix)) == 0 {
		return fmt.Sprintf("%s.%s.%s", applicationName, event, version)
	}
	return fmt.Sprintf("%s.%s.%s.%s", prefix, applicationName, event, version)
}
