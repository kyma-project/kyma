package helpers

import "fmt"

const (
	LambdaPort               = 80
	LambdaPayload            = "payload"
	KymaIntegrationNamespace = "kyma-integration"
	DefaultBrokerName        = "default"
)

func LambdaInClusterEndpoint(name, namespace string, port int) string {
	return fmt.Sprintf("http://%s.%s:%v", name, namespace, port)
}
