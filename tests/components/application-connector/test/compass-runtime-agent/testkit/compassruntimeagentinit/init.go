package compassruntimeagentinit

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
)

type CompassRuntimeAgentConfigurator interface {
	Do(runtimeName string) (RollbackFunc, error)
}

type DirectorClient interface {
	RegisterRuntime(appName, scenarioName string) (string, error)
	UnregisterRuntime(id string) error
	GetConnectionToken(runtimeID string) error
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

func (crc compassRuntimeAgentConfigurator) Do(runtimeName string) (RollbackFunc, error) {
	runtimeID, err := crc.registerRuntime(runtimeName)
	if err != nil {
		return nil, err
	}

	token, compassConnectorUrl, err := crc.getTokenUrl()
	if err != nil {
		{
			err := newRollbackFunc(runtimeID, crc.directorClient, nil, nil)()
			if err != nil {
				return nil, errors.Wrap(err, "failed to get token URL and unregister runtime")
			}
		}

		return nil, errors.Wrap(err, "failed to get token URL")
	}

	compassRuntimeAgenConfig := CompassRuntimeAgentConfig{
		ConnectorUrl: compassConnectorUrl,
		RuntimeID:    runtimeID,
		Token:        token,
		Tenant:       crc.tenant,
	}

	secretRollbackFunc, err := crc.createCompassRuntimeAgentSecret(compassRuntimeAgenConfig)
	if err != nil {
		{
			err := newRollbackFunc(runtimeID, crc.directorClient, secretRollbackFunc, nil)()
			if err != nil {
				return nil, errors.Wrap(err, "failed to create Compass Runtime Configuration secret and unregister runtime")
			}
		}

		return nil, errors.Wrap(err, "failed to create Compass Runtime Configuration secret")
	}

	deploymentRollbackFunc, err := crc.modifyDeployment()
	if err != nil {
		err := newRollbackFunc(runtimeID, crc.directorClient, secretRollbackFunc, nil)()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create Compass Runtime Configuration secret and unregister runtime")
		}
		return nil, err
	}

	return newRollbackFunc(runtimeID, crc.directorClient, secretRollbackFunc, deploymentRollbackFunc), nil
}

func (crc compassRuntimeAgentConfigurator) registerRuntime(runtimeName string) (string, error) {
	return "", nil
}

func (crc compassRuntimeAgentConfigurator) getTokenUrl() (string, string, error) {
	return "", "", nil
}

func (crc compassRuntimeAgentConfigurator) createCompassRuntimeAgentSecret(config CompassRuntimeAgentConfig) (RollbackSecretFunc, error) {
	return newSecretCreator(crc.coreClientset).Do("", "", config)
}

func (crc compassRuntimeAgentConfigurator) modifyDeployment() (RollbackDeploymentFunc, error) {
	return newDeploymentConfiguration(crc.coreClientset).Do("", "")
}
