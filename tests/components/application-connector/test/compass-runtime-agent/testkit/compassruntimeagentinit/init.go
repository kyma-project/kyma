package compassruntimeagentinit

import (
	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/compassruntimeagentinit/types"
	"github.com/pkg/errors"
)

const (
	CompassSystemNamespace        = "compass-system"
	CompassRuntimeAgentDeployment = "compass-runtime-agent"
	NewCompassRuntimeConfigName   = "test-compass-runtime-agent-config"
	NewCACertSecretName           = "ca-cert-test"
	NewClientCertSecretName       = "client-cert-test"
	RetryAttempts                 = 6
	RetrySeconds                  = 5
)

type CompassRuntimeAgentConfigurator interface {
	Do(runtimeName string) (types.RollbackFunc, error)
}

type compassRuntimeAgentConfigurator struct {
	directorClient                  types.DirectorClient
	certificateSecretConfigurator   types.CertificateSecretConfigurator
	configurationSecretConfigurator types.ConfigurationSecretConfigurator
	compassConnectionConfigurator   types.CompassConnectionConfigurator
	deploymentConfigurator          types.DeploymentConfigurator
	tenant                          string
}

func NewCompassRuntimeAgentConfigurator(directorClient types.DirectorClient,
	certificateSecretConfigurator types.CertificateSecretConfigurator,
	configurationSecretConfigurator types.ConfigurationSecretConfigurator,
	compassConnectionConfigurator types.CompassConnectionConfigurator,
	deploymentConfigurator types.DeploymentConfigurator,
	tenant string) CompassRuntimeAgentConfigurator {
	return compassRuntimeAgentConfigurator{
		directorClient:                  directorClient,
		certificateSecretConfigurator:   certificateSecretConfigurator,
		configurationSecretConfigurator: configurationSecretConfigurator,
		compassConnectionConfigurator:   compassConnectionConfigurator,
		deploymentConfigurator:          deploymentConfigurator,
		tenant:                          tenant,
	}
}

func (crc compassRuntimeAgentConfigurator) Do(runtimeName string) (types.RollbackFunc, error) {
	runtimeID, err := crc.directorClient.RegisterRuntime(runtimeName)
	if err != nil {
		return nil, err
	}

	token, compassConnectorUrl, err := crc.directorClient.GetConnectionToken(runtimeID)
	if err != nil {
		return nil, crc.rollbackOnError(err, "failed to get token URL", runtimeID)
	}

	config := types.CompassRuntimeAgentConfig{
		ConnectorUrl: compassConnectorUrl,
		RuntimeID:    runtimeID,
		Token:        token,
		Tenant:       crc.tenant,
	}

	certificateSecretsRollbackFunc, err := crc.certificateSecretConfigurator.Do(NewCACertSecretName, NewClientCertSecretName)
	if err != nil {
		return nil, crc.rollbackOnError(err, "failed to create Compass Runtime Configuration secret",
			runtimeID,
			certificateSecretsRollbackFunc)
	}

	configurationSecretRollbackFunc, err := crc.configurationSecretConfigurator.Do(NewCompassRuntimeConfigName, config)
	if err != nil {
		return nil, crc.rollbackOnError(err, "failed to create Compass Runtime Configuration secret",
			runtimeID,
			certificateSecretsRollbackFunc,
			configurationSecretRollbackFunc)
	}

	deploymentRollbackFunc, err := crc.deploymentConfigurator.Do(NewCACertSecretName, NewClientCertSecretName, NewCompassRuntimeConfigName)
	if err != nil {
		return nil, crc.rollbackOnError(err, "failed to modify deployment",
			runtimeID,
			certificateSecretsRollbackFunc,
			configurationSecretRollbackFunc,
			deploymentRollbackFunc)
	}

	compassConnectionRollbackFunc, err := crc.compassConnectionConfigurator.Do()
	if err != nil {
		return nil, crc.rollbackOnError(err, "failed to configure Compass Connection CR",
			runtimeID,
			certificateSecretsRollbackFunc,
			configurationSecretRollbackFunc,
			deploymentRollbackFunc,
			compassConnectionRollbackFunc)
	}

	return newRollbackFunc(runtimeID,
		crc.directorClient,
		compassConnectionRollbackFunc,
		certificateSecretsRollbackFunc,
		configurationSecretRollbackFunc,
		deploymentRollbackFunc), nil
}

func (crc compassRuntimeAgentConfigurator) rollbackOnError(initialError error, wrapMsgString, runtimeID string, rollbackFunctions ...types.RollbackFunc) error {
	err := newRollbackFunc(runtimeID, crc.directorClient, rollbackFunctions...)()
	if err != nil {
		initialWrapped := errors.Wrap(initialError, wrapMsgString)
		errorWrapped := errors.Wrap(err, "failed to rollback changes after configuring Compass Runtime Agent error")

		return errors.Wrap(errorWrapped, initialWrapped.Error())
	}

	return initialError
}

func newRollbackFunc(runtimeID string, directorClient types.DirectorClient, rollbackFunctions ...types.RollbackFunc) types.RollbackFunc {
	var result *multierror.Error

	return func() error {
		if err := directorClient.UnregisterRuntime(runtimeID); err != nil {
			multierror.Append(result, err)
		}

		for _, f := range rollbackFunctions {
			if f != nil {
				if err := f(); err != nil {
					multierror.Append(result, err)
				}
			}
		}

		return result.ErrorOrNil()
	}
}
