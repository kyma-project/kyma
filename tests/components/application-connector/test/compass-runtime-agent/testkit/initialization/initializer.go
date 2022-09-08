package initialization

import "k8s.io/client-go/kubernetes"

type CompassRuntimeAgentConfigurator interface {
	Configure(runtimeName string) error
}

type DirectorClient interface {
	RegisterRuntime(appName, scenario string) (string, error)
	UnregisterApplication(id string) error
}

type CompassRuntimeAgentConfig struct {
	ConnectorUrl string
	RuntimeID    string
	Token        string
	Tenant       string
}

func NewCompassRuntimeAgentConfigurator(directorClient DirectorClient, coreClientset *kubernetes.Clientset, tenant string) CompassRuntimeAgentConfigurator {
	return compassRuntimeAgentConfigurator{
		directorClient: directorClient,
		coreClientset:  coreClientset,
		tenant:         tenant,
	}
}

type compassRuntimeAgentConfigurator struct {
	directorClient DirectorClient
	coreClientset  *kubernetes.Clientset
	tenant         string
}

func (crc compassRuntimeAgentConfigurator) Configure(runtimeName string) error {
	runtimeID, err := crc.registerRuntime(runtimeName)
	if err != nil {
		return err
	}

	token, compassConnectorUrl, err := crc.getTokenUrl()
	if err != nil {
		return err
	}

	compassRuntimeAgenConfig := CompassRuntimeAgentConfig{
		ConnectorUrl: compassConnectorUrl,
		RuntimeID:    runtimeID,
		Token:        token,
		Tenant:       crc.tenant,
	}

	err = crc.createCompassRuntimeAgentConfig(compassRuntimeAgenConfig)
	if err != nil {
		return err
	}

	err = crc.modifyDeployment()
	if err != nil {
		return err
	}

	return crc.waitForDeploymentRollout()
}

func (crc compassRuntimeAgentConfigurator) registerRuntime(runtimeName string) (string, error) {
	return "", nil
}

func (crc compassRuntimeAgentConfigurator) getTokenUrl() (string, string, error) {
	return "", "", nil
}

func (crc compassRuntimeAgentConfigurator) createCompassRuntimeAgentConfig(config CompassRuntimeAgentConfig) error {
	return nil
}

func (crc compassRuntimeAgentConfigurator) modifyDeployment() error {
	return nil
}

func (crc compassRuntimeAgentConfigurator) waitForDeploymentRollout() error {
	return nil
}
