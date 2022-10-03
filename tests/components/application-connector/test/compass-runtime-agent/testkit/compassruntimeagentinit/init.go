package compassruntimeagentinit

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/compassruntimeagentinit/types"
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
	compassConfigurator             types.CompassConfigurator
	certificateSecretConfigurator   types.CertificateSecretConfigurator
	configurationSecretConfigurator types.ConfigurationSecretConfigurator
	compassConnectionConfigurator   types.CompassConnectionConfigurator
	deploymentConfigurator          types.DeploymentConfigurator
	testNamespace                   string
}

func NewCompassRuntimeAgentConfigurator(compassConfigurator types.CompassConfigurator,
	certificateSecretConfigurator types.CertificateSecretConfigurator,
	configurationSecretConfigurator types.ConfigurationSecretConfigurator,
	compassConnectionConfigurator types.CompassConnectionConfigurator,
	deploymentConfigurator types.DeploymentConfigurator,
	testNamespace string) CompassRuntimeAgentConfigurator {
	return compassRuntimeAgentConfigurator{
		compassConfigurator:             compassConfigurator,
		certificateSecretConfigurator:   certificateSecretConfigurator,
		configurationSecretConfigurator: configurationSecretConfigurator,
		compassConnectionConfigurator:   compassConnectionConfigurator,
		deploymentConfigurator:          deploymentConfigurator,
		testNamespace:                   testNamespace,
	}
}

func (crc compassRuntimeAgentConfigurator) Do(runtimeName, formationName string) (types.RollbackFunc, error) {
	compassRuntimeAgentConfig, compassConfiguratorRollbackFunc, err := crc.compassConfigurator.Do(runtimeName, formationName)
	if err != nil {
		return nil, crc.rollbackOnError(err,
			compassConfiguratorRollbackFunc)
	}

	certificateSecretsRollbackFunc, err := crc.certificateSecretConfigurator.Do(NewCACertSecretName, NewClientCertSecretName)
	if err != nil {
		return nil, crc.rollbackOnError(err,
			compassConfiguratorRollbackFunc,
			certificateSecretsRollbackFunc)
	}

	configurationSecretRollbackFunc, err := crc.configurationSecretConfigurator.Do(NewCompassRuntimeConfigName, compassRuntimeAgentConfig)
	if err != nil {
		return nil, crc.rollbackOnError(err,
			compassConfiguratorRollbackFunc,
			certificateSecretsRollbackFunc,
			configurationSecretRollbackFunc)
	}

	newCACertNamespacedSecretName := fmt.Sprintf("%s/%s", "istio-system", NewCACertSecretName)
	newClientCertNamespacedSecretName := fmt.Sprintf("%s/%s", "compass-system", NewClientCertSecretName)
	newCompassRuntimeNamespacedSecretConfigName := fmt.Sprintf("%s/%s", "compass-system", NewCompassRuntimeConfigName)

	deploymentRollbackFunc, err := crc.deploymentConfigurator.Do(newCACertNamespacedSecretName,
		newClientCertNamespacedSecretName,
		newCompassRuntimeNamespacedSecretConfigName)
	if err != nil {
		return nil, crc.rollbackOnError(err,
			compassConfiguratorRollbackFunc,
			certificateSecretsRollbackFunc,
			configurationSecretRollbackFunc,
			deploymentRollbackFunc)
	}

	compassConnectionRollbackFunc, err := crc.compassConnectionConfigurator.Do()
	if err != nil {
		return nil, crc.rollbackOnError(err,
			compassConfiguratorRollbackFunc,
			certificateSecretsRollbackFunc,
			configurationSecretRollbackFunc,
			deploymentRollbackFunc,
			compassConnectionRollbackFunc)
	}

	return newRollbackFunc(compassConnectionRollbackFunc,
		certificateSecretsRollbackFunc,
		configurationSecretRollbackFunc,
		deploymentRollbackFunc), nil
}

func (crc compassRuntimeAgentConfigurator) rollbackOnError(initialErr error, rollbackFunctions ...types.RollbackFunc) error {
	var result *multierror.Error
	multierror.Append(result, initialErr)

	err := newRollbackFunc(rollbackFunctions...)()
	if err != nil {
		multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

func newRollbackFunc(rollbackFunctions ...types.RollbackFunc) types.RollbackFunc {
	var result *multierror.Error

	return func() error {
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
