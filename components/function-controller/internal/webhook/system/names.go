package system

import "os"

const (
	DefaultNamespace      = "kyma-system"
	SystemNamespaceEnvVar = "SYSTEM_NAMESPACE"
)

// GetNamespace holds the K8s namespace where our system components run.
func GetNamespace() string {
	systemNamespace := os.Getenv(SystemNamespaceEnvVar)
	if systemNamespace == "" {
		return DefaultNamespace
	}
	return systemNamespace
}
