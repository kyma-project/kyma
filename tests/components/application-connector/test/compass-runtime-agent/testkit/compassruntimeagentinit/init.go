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

type compassRuntimeAgentConfigurator struct {
	directorClient      DirectorClient
	kubernetesInterface kubernetes.Interface
	tenant              string
}

func NewCompassRuntimeAgentConfigurator(directorClient DirectorClient, kubernetesInterface kubernetes.Interface, tenant string) CompassRuntimeAgentConfigurator {
	return compassRuntimeAgentConfigurator{
		directorClient:      directorClient,
		kubernetesInterface: kubernetesInterface,
		tenant:              tenant,
	}
}

func (crc compassRuntimeAgentConfigurator) Do(runtimeName string) (RollbackFunc, error) {
	runtimeID, err := crc.registerRuntime(runtimeName)
	if err != nil {
		return nil, err
	}

	token, compassConnectorUrl, err := crc.getTokenUrl()
	if err != nil {
		return nil, crc.rollbackOnError(err, "failed to get token URL", runtimeID, nil, nil)
	}

	compassRuntimeAgentConfig := CompassRuntimeAgentConfig{
		ConnectorUrl: compassConnectorUrl,
		RuntimeID:    runtimeID,
		Token:        token,
		Tenant:       crc.tenant,
	}

	secretRollbackFunc, err := crc.createCompassRuntimeAgentSecret(compassRuntimeAgentConfig)
	if err != nil {
		return nil, crc.rollbackOnError(err, "failed to create Compass Runtime Configuration secret", runtimeID, nil, nil)
	}

	deploymentRollbackFunc, err := crc.modifyDeployment()
	if err != nil {
		return nil, crc.rollbackOnError(err, "failed to modify deployment", runtimeID, nil, nil)
	}

	return newRollbackFunc(runtimeID, crc.directorClient, secretRollbackFunc, deploymentRollbackFunc), nil
}

func (crc compassRuntimeAgentConfigurator) rollbackOnError(initialError error, wrapMsgString, runtimeID string, secretRollbackFunc RollbackSecretFunc, deploymentRollbackFunc RollbackDeploymentFunc) error {
	err := newRollbackFunc(runtimeID, crc.directorClient, secretRollbackFunc, deploymentRollbackFunc)()
	if err != nil {
		initialWrapped := errors.Wrap(initialError, wrapMsgString)
		errorWrapped := errors.Wrap(err, "failed to rollback changes after configuring Compass Runtime Agent error")

		return errors.Wrap(errorWrapped, initialWrapped.Error())
	}

	return initialError
}

func (crc compassRuntimeAgentConfigurator) registerRuntime(runtimeName string) (string, error) {
	return "", nil
}

func (crc compassRuntimeAgentConfigurator) getTokenUrl() (string, string, error) {
	return "", "", nil
}

func (crc compassRuntimeAgentConfigurator) createCompassRuntimeAgentSecret(config CompassRuntimeAgentConfig) (RollbackSecretFunc, error) {
	return newSecretCreator(crc.kubernetesInterface).Do("", "", config)
}

func (crc compassRuntimeAgentConfigurator) modifyDeployment() (RollbackDeploymentFunc, error) {
	return newDeploymentConfiguration(crc.kubernetesInterface).Do("", "")
}
