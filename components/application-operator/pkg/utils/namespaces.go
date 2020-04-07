package utils

const (
	kymaSystemNamespace      = "kyma-system"
	kymaIntegrationNamespace = "kyma-integration"
)

func IsSystemNamespace(namespace string) bool {
	return namespace == kymaIntegrationNamespace || namespace == kymaSystemNamespace
}
