package compassruntimeagentinit

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
)

const (
	CompassSystemNamespace        = "compass-system"
	CompassRuntimeAgentDeployment = "compass-runtime-agent"
	NewCompassRuntimeConfigName   = "test-compass-runtime-agent-config"
	RetryAttempts                 = 6
	RetrySeconds                  = 5
)

type CompassRuntimeAgentConfigurator interface {
	Do(runtimeName string) (RollbackFunc, error)
}

//go:generate mockery --name=DirectorClient
type DirectorClient interface {
	RegisterRuntime(runtimeName string) (string, error)
	UnregisterRuntime(id string) error
	GetConnectionToken(runtimeID string) (string, string, error)
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
	runtimeID, err := crc.directorClient.RegisterRuntime(runtimeName)
	if err != nil {
		return nil, err
	}

	token, compassConnectorUrl, err := crc.directorClient.GetConnectionToken(runtimeID)
	if err != nil {
		return nil, crc.rollbackOnError(err, "failed to get token URL", runtimeID, nil, nil)
	}

	config := CompassRuntimeAgentConfig{
		ConnectorUrl: compassConnectorUrl,
		RuntimeID:    runtimeID,
		Token:        token,
		Tenant:       crc.tenant,
	}

	secretRollbackFunc, err := newSecretCreator(crc.kubernetesInterface).Do(NewCompassRuntimeConfigName, CompassSystemNamespace, config)
	if err != nil {
		return nil, crc.rollbackOnError(err, "failed to create Compass Runtime Configuration secret", runtimeID, secretRollbackFunc, nil)
	}

	deploymentRollbackFunc, err := newDeploymentConfiguration(crc.kubernetesInterface).Do(CompassRuntimeAgentDeployment, NewCompassRuntimeConfigName, CompassSystemNamespace)
	if err != nil {
		return nil, crc.rollbackOnError(err, "failed to modify deployment", runtimeID, secretRollbackFunc, deploymentRollbackFunc)
	}

	return newRollbackFunc(runtimeID, crc.directorClient, secretRollbackFunc, deploymentRollbackFunc), nil
}

func (crc compassRuntimeAgentConfigurator) rollbackOnError(initialError error, wrapMsgString, runtimeID string, secretRollbackFunc, deploymentRollbackFunc RollbackFunc) error {
	err := newRollbackFunc(runtimeID, crc.directorClient, secretRollbackFunc, deploymentRollbackFunc)()
	if err != nil {
		initialWrapped := errors.Wrap(initialError, wrapMsgString)
		errorWrapped := errors.Wrap(err, "failed to rollback changes after configuring Compass Runtime Agent error")

		return errors.Wrap(errorWrapped, initialWrapped.Error())
	}

	return initialError
}
