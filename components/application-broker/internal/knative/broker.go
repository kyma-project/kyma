package knative

import (
	"fmt"

	"github.com/kyma-project/kyma/components/application-broker/internal"
)

// GetDefaultBrokerURI returns the default broker URI for a given namespace.
func GetDefaultBrokerURI(ns internal.Namespace) string {
	return fmt.Sprintf("http://default-broker.%s", ns)
}
