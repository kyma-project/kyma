package initialization

import "k8s.io/client-go/kubernetes"

type CompassRuntimeAgentConfigurator interface {
	Configure(runtimeName string) error
}

type DirectorClient interface {
	RegisterRuntime(appName, scenario string) (string, error)
	UnregisterApplication(id string) error
}

func NewCompassRuntimeAgentConfigurator(directorClient DirectorClient) CompassRuntimeAgentConfigurator {
	return compassRuntimeAgentConfigurator{
		directorClient: directorClient,
	}
}

type compassRuntimeAgentConfigurator struct {
	directorClient DirectorClient
	coreClientset  *kubernetes.Clientset
}

func (crc compassRuntimeAgentConfigurator) Configure(runtimeName string) error {
	return nil
}
