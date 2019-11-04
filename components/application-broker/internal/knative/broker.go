package knative

import "fmt"

// GetDefaultBrokerURI returns the default broker URI for a given namespace.
func GetDefaultBrokerURI(ns string) string {
	return fmt.Sprintf("http://default-broker.%s", ns)
}
