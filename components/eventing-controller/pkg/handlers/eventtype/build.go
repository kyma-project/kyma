package eventtype

import "fmt"

func build(prefix, applicationName, event, version string) string {
	return fmt.Sprintf("%s.%s.%s.%s", prefix, applicationName, event, version)
}
