package init

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/init/types"
	log "github.com/sirupsen/logrus"
)

const (
	CompassSystemNamespace        = "kyma-system"
	IstioSystemNamespace          = "istio-system"
	CompassRuntimeAgentDeployment = "compass-runtime-agent"
	NewCompassRuntimeConfigName   = "test-compass-runtime-agent-config"
	NewCACertSecretName           = "ca-cert-test"
	NewClientCertSecretName       = "client-cert-test"
	NewControllerSyncPeriodTime   = "15s"
	RetryAttempts                 = 6
	RetrySeconds                  = 5
)

type CompassRuntimeAgentConfigurator interface {
	Do(runtimeName, formationName string) (types.RollbackFunc, error)
}

type Configurator interface {
	Configure(runtimeName, formationName string) (types.RollbackFunc, error)
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
	log.Info("Configuring Compass")
	compassRuntimeAgentConfig, compassConfiguratorRollbackFunc, err := crc.compassConfigurator.Do(runtimeName, formationName)
	if err != nil {
		return nil, crc.rollbackOnError(err,
			compassConfiguratorRollbackFunc)
	}

	log.Info("Configuring certificate secrets")
	certificateSecretsRollbackFunc, err := crc.certificateSecretConfigurator.Do(NewCACertSecretName, NewClientCertSecretName)
	if err != nil {
		return nil, crc.rollbackOnError(err,
			compassConfiguratorRollbackFunc,
			certificateSecretsRollbackFunc)
	}

	log.Info("Preparing Compass Runtime Agent configuration secret")
	configurationSecretRollbackFunc, err := crc.configurationSecretConfigurator.Do(NewCompassRuntimeConfigName, compassRuntimeAgentConfig)
	if err != nil {
		return nil, crc.rollbackOnError(err,
			compassConfiguratorRollbackFunc,
			certificateSecretsRollbackFunc,
			configurationSecretRollbackFunc)
	}

	newCACertNamespacedSecretName := fmt.Sprintf("%s/%s", IstioSystemNamespace, NewCACertSecretName)
	newClientCertNamespacedSecretName := fmt.Sprintf("%s/%s", CompassSystemNamespace, NewClientCertSecretName)
	newCompassRuntimeNamespacedSecretConfigName := fmt.Sprintf("%s/%s", CompassSystemNamespace, NewCompassRuntimeConfigName)
	newControllerSyncPeriodTime := fmt.Sprintf("%s", NewControllerSyncPeriodTime)

	log.Info("Preparing Compass Runtime Agent configuration secret")
	deploymentRollbackFunc, err := crc.deploymentConfigurator.Do(newCACertNamespacedSecretName,
		newClientCertNamespacedSecretName,
		newCompassRuntimeNamespacedSecretConfigName, newControllerSyncPeriodTime)
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

	return newRollbackFunc(compassConfiguratorRollbackFunc,
		certificateSecretsRollbackFunc,
		configurationSecretRollbackFunc,
		deploymentRollbackFunc,
		compassConnectionRollbackFunc), nil
}

func (crc compassRuntimeAgentConfigurator) rollbackOnError(initialErr error, rollbackFunctions ...types.RollbackFunc) error {
	var result *multierror.Error
	result = multierror.Append(result, initialErr)

	err := newRollbackFunc(rollbackFunctions...)()
	if err != nil {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

func newRollbackFunc(rollbackFunctions ...types.RollbackFunc) types.RollbackFunc {
	var result *multierror.Error

	return func() error {
		for _, f := range rollbackFunctions {
			if f != nil {
				if err := f(); err != nil {
					result = multierror.Append(result, err)
				}
			}
		}

		return result.ErrorOrNil()
	}
}
