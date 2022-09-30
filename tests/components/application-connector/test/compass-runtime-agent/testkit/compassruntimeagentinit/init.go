package compassruntimeagentinit

import (
	"fmt"
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
	Do(runtimeName, formationName string) (types.RollbackFunc, error)
}

type compassRuntimeAgentConfigurator struct {
	directorClient                  types.DirectorClient
	certificateSecretConfigurator   types.CertificateSecretConfigurator
	configurationSecretConfigurator types.ConfigurationSecretConfigurator
	compassConnectionConfigurator   types.CompassConnectionConfigurator
	deploymentConfigurator          types.DeploymentConfigurator
	tenant                          string
	testNamespace                   string
}

func NewCompassRuntimeAgentConfigurator(directorClient types.DirectorClient,
	certificateSecretConfigurator types.CertificateSecretConfigurator,
	configurationSecretConfigurator types.ConfigurationSecretConfigurator,
	compassConnectionConfigurator types.CompassConnectionConfigurator,
	deploymentConfigurator types.DeploymentConfigurator,
	tenant string,
	testNamespace string) CompassRuntimeAgentConfigurator {
	return compassRuntimeAgentConfigurator{
		directorClient:                  directorClient,
		certificateSecretConfigurator:   certificateSecretConfigurator,
		configurationSecretConfigurator: configurationSecretConfigurator,
		compassConnectionConfigurator:   compassConnectionConfigurator,
		deploymentConfigurator:          deploymentConfigurator,
		tenant:                          tenant,
		testNamespace:                   testNamespace,
	}
}

func (crc compassRuntimeAgentConfigurator) Do(runtimeName, formationName string) (types.RollbackFunc, error) {
	runtimeID, err := crc.directorClient.RegisterRuntime(runtimeName)
	if err != nil {
		return nil, err
	}

	err = crc.directorClient.RegisterFormation(formationName)
	if err != nil {
		return nil, err
	}

	err = crc.directorClient.AssignRuntimeToFormation(runtimeID, formationName)
	if err != nil {
		return nil, err
	}

	token, compassConnectorUrl, err := crc.directorClient.GetConnectionToken(runtimeID)
	if err != nil {
		return nil, crc.rollbackOnError(err, "failed to get token URL", runtimeID, formationName)
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
			formationName,
			certificateSecretsRollbackFunc)
	}

	configurationSecretRollbackFunc, err := crc.configurationSecretConfigurator.Do(NewCompassRuntimeConfigName, config)
	if err != nil {
		return nil, crc.rollbackOnError(err, "failed to create Compass Runtime Configuration secret",
			runtimeID,
			formationName,
			certificateSecretsRollbackFunc,
			configurationSecretRollbackFunc)
	}

	newCACertNamespacedSecretName := fmt.Sprintf("%s/%s", "istio-system", NewCACertSecretName)
	newClientCertNamespacedSecretName := fmt.Sprintf("%s/%s", "compass-system", NewClientCertSecretName)
	newCompassRuntimeNamespacedSecretConfigName := fmt.Sprintf("%s/%s", "compass-system", NewCompassRuntimeConfigName)

	deploymentRollbackFunc, err := crc.deploymentConfigurator.Do(newCACertNamespacedSecretName, newClientCertNamespacedSecretName, newCompassRuntimeNamespacedSecretConfigName)
	if err != nil {
		return nil, crc.rollbackOnError(err, "failed to modify deployment",
			runtimeID,
			formationName,
			certificateSecretsRollbackFunc,
			configurationSecretRollbackFunc,
			deploymentRollbackFunc)
	}

	compassConnectionRollbackFunc, err := crc.compassConnectionConfigurator.Do()
	if err != nil {
		return nil, crc.rollbackOnError(err, "failed to configure Compass Connection CR",
			runtimeID,
			formationName,
			certificateSecretsRollbackFunc,
			configurationSecretRollbackFunc,
			deploymentRollbackFunc,
			compassConnectionRollbackFunc)
	}

	return newRollbackFunc(runtimeID,
		formationName,
		crc.directorClient,
		compassConnectionRollbackFunc,
		certificateSecretsRollbackFunc,
		configurationSecretRollbackFunc,
		deploymentRollbackFunc), nil
}

func (crc compassRuntimeAgentConfigurator) rollbackOnError(initialError error, wrapMsgString, runtimeID, formationName string, rollbackFunctions ...types.RollbackFunc) error {
	err := newRollbackFunc(runtimeID, formationName, crc.directorClient, rollbackFunctions...)()
	if err != nil {
		initialWrapped := errors.Wrap(initialError, wrapMsgString)
		errorWrapped := errors.Wrap(err, "failed to rollback changes after configuring Compass Runtime Agent error")

		return errors.Wrap(errorWrapped, initialWrapped.Error())
	}

	return initialError
}

func newRollbackFunc(runtimeID, formationName string, directorClient types.DirectorClient, rollbackFunctions ...types.RollbackFunc) types.RollbackFunc {
	var result *multierror.Error

	return func() error {
		if err := directorClient.UnregisterRuntime(runtimeID); err != nil {
			multierror.Append(result, err)
		}

		if err := directorClient.UnregisterFormation(formationName); err != nil {
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
