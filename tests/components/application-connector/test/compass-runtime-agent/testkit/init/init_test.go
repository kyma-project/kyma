package init

import (
	"fmt"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/init/types"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/init/types/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCompassRuntimeAgentInit(t *testing.T) {
	runtimeName := "newRuntime"
	runtimeID := "runtimeID"
	token := "token"
	connectorURL := "www.someurl.com"
	tenant := "tenant"
	formationName := "newFormation"

	t.Run("should succeed and return rollback function", func(t *testing.T) {
		// given
		compassConfiguratorMock := &mocks.CompassConfigurator{}
		certificateSecretConfiguratorMock := &mocks.CertificateSecretConfigurator{}
		configurationSecretConfiguratorMock := &mocks.ConfigurationSecretConfigurator{}
		compassConnectionConfiguratorMock := &mocks.CompassConnectionConfigurator{}
		deploymentConfiguratorMock := &mocks.DeploymentConfigurator{}

		compassConfiguratorRollbackFunc := RollbackFuncTest{}
		certificateSecretsRollbackFunc := RollbackFuncTest{}
		configurationSecretRollbackFunc := RollbackFuncTest{}
		compassConnectionRollbackFunc := RollbackFuncTest{}
		deploymentRollbackFunc := RollbackFuncTest{}

		config := types.CompassRuntimeAgentConfig{
			ConnectorUrl: connectorURL,
			RuntimeID:    runtimeID,
			Token:        token,
			Tenant:       tenant,
		}

		compassConfiguratorMock.On("Do", runtimeName, formationName).Return(config, compassConfiguratorRollbackFunc.Func(), nil)
		certificateSecretConfiguratorMock.On("Do", NewCACertSecretName, NewClientCertSecretName).Return(certificateSecretsRollbackFunc.Func(), nil)
		configurationSecretConfiguratorMock.On("Do", NewCompassRuntimeConfigName, config).Return(configurationSecretRollbackFunc.Func(), nil)
		compassConnectionConfiguratorMock.On("Do").Return(compassConnectionRollbackFunc.Func(), nil)
		deploymentConfiguratorMock.On("Do",
			fmt.Sprintf("%s/%s", IstioSystemNamespace, NewCACertSecretName),
			fmt.Sprintf("%s/%s", CompassSystemNamespace, NewClientCertSecretName),
			fmt.Sprintf("%s/%s", CompassSystemNamespace, NewCompassRuntimeConfigName),
			fmt.Sprintf("%s", NewControllerSyncPeriodTime)).
			Return(deploymentRollbackFunc.Func(), nil)

		configurator := NewCompassRuntimeAgentConfigurator(compassConfiguratorMock, certificateSecretConfiguratorMock, configurationSecretConfiguratorMock, compassConnectionConfiguratorMock, deploymentConfiguratorMock, "tenant")

		// when
		rollbackFunc, err := configurator.Do(runtimeName, formationName)

		// then
		require.NoError(t, err)
		certificateSecretConfiguratorMock.AssertExpectations(t)
		compassConnectionConfiguratorMock.AssertExpectations(t)
		deploymentConfiguratorMock.AssertExpectations(t)

		//when
		err = rollbackFunc()

		// then
		require.NoError(t, err)
		require.True(t, compassConfiguratorRollbackFunc.invoked)
		require.True(t, certificateSecretsRollbackFunc.invoked)
		require.True(t, configurationSecretRollbackFunc.invoked)
		require.True(t, compassConnectionRollbackFunc.invoked)
		require.True(t, deploymentRollbackFunc.invoked)
	})

	t.Run("should fail if failed to register runtime", func(t *testing.T) {
		// given
		compassConfiguratorMock := &mocks.CompassConfigurator{}
		compassConfiguratorRollbackFunc := RollbackFuncTest{}

		compassConfiguratorMock.On("Do", runtimeName, formationName).Return(types.CompassRuntimeAgentConfig{}, compassConfiguratorRollbackFunc.Func(), errors.New("some error"))

		configurator := NewCompassRuntimeAgentConfigurator(compassConfiguratorMock, nil, nil, nil, nil, "tenant")

		// when
		rollbackFunc, err := configurator.Do(runtimeName, formationName)

		// then
		require.Error(t, err)
		require.Nil(t, rollbackFunc)
		assert.True(t, compassConfiguratorRollbackFunc.invoked)
	})

	t.Run("should fail if failed to create configuration secret", func(t *testing.T) {
		// given
		compassConfiguratorMock := &mocks.CompassConfigurator{}
		certificateSecretConfiguratorMock := &mocks.CertificateSecretConfigurator{}
		configurationSecretConfiguratorMock := &mocks.ConfigurationSecretConfigurator{}

		compassConfiguratorRollbackFunc := RollbackFuncTest{}
		certificateSecretsRollbackFunc := RollbackFuncTest{}

		config := types.CompassRuntimeAgentConfig{
			ConnectorUrl: connectorURL,
			RuntimeID:    runtimeID,
			Token:        token,
			Tenant:       tenant,
		}

		compassConfiguratorMock.On("Do", runtimeName, formationName).Return(config, compassConfiguratorRollbackFunc.Func(), nil)
		certificateSecretConfiguratorMock.On("Do", NewCACertSecretName, NewClientCertSecretName).Return(certificateSecretsRollbackFunc.Func(), nil)
		configurationSecretConfiguratorMock.On("Do", NewCompassRuntimeConfigName, config).Return(nil, errors.New("some error"))

		configurator := NewCompassRuntimeAgentConfigurator(compassConfiguratorMock, certificateSecretConfiguratorMock, configurationSecretConfiguratorMock, nil, nil, "tenant")

		// when
		rollbackFunc, err := configurator.Do(runtimeName, formationName)

		// then
		require.Error(t, err)
		require.Nil(t, rollbackFunc)
		compassConfiguratorMock.AssertExpectations(t)
		certificateSecretConfiguratorMock.AssertExpectations(t)
		certificateSecretConfiguratorMock.AssertExpectations(t)
		require.True(t, compassConfiguratorRollbackFunc.invoked)
		require.True(t, certificateSecretsRollbackFunc.invoked)
	})

	t.Run("should fail if failed to modify deployment", func(t *testing.T) {
		// given
		compassConfiguratorMock := &mocks.CompassConfigurator{}
		certificateSecretConfiguratorMock := &mocks.CertificateSecretConfigurator{}
		configurationSecretConfiguratorMock := &mocks.ConfigurationSecretConfigurator{}
		deploymentConfiguratorMock := &mocks.DeploymentConfigurator{}

		compassConfiguratorRollbackFunc := RollbackFuncTest{}
		certificateSecretsRollbackFunc := RollbackFuncTest{}
		configurationSecretRollbackFunc := RollbackFuncTest{}

		config := types.CompassRuntimeAgentConfig{
			ConnectorUrl: connectorURL,
			RuntimeID:    runtimeID,
			Token:        token,
			Tenant:       tenant,
		}

		compassConfiguratorMock.On("Do", runtimeName, formationName).Return(config, compassConfiguratorRollbackFunc.Func(), nil)
		certificateSecretConfiguratorMock.On("Do", NewCACertSecretName, NewClientCertSecretName).Return(certificateSecretsRollbackFunc.Func(), nil)
		configurationSecretConfiguratorMock.On("Do", NewCompassRuntimeConfigName, config).Return(configurationSecretRollbackFunc.Func(), nil)
		deploymentConfiguratorMock.On("Do",
			fmt.Sprintf("%s/%s", IstioSystemNamespace, NewCACertSecretName),
			fmt.Sprintf("%s/%s", CompassSystemNamespace, NewClientCertSecretName),
			fmt.Sprintf("%s/%s", CompassSystemNamespace, NewCompassRuntimeConfigName),
			fmt.Sprintf("%s", NewControllerSyncPeriodTime)).
			Return(nil, errors.New("some error"))

		configurator := NewCompassRuntimeAgentConfigurator(compassConfiguratorMock, certificateSecretConfiguratorMock, configurationSecretConfiguratorMock, nil, deploymentConfiguratorMock, "tenant")

		// when
		rollbackFunc, err := configurator.Do(runtimeName, formationName)

		// then
		require.Error(t, err)
		require.Nil(t, rollbackFunc)
		certificateSecretConfiguratorMock.AssertExpectations(t)
		deploymentConfiguratorMock.AssertExpectations(t)
		require.True(t, compassConfiguratorRollbackFunc.invoked)
		require.True(t, certificateSecretsRollbackFunc.invoked)
		require.True(t, configurationSecretRollbackFunc.invoked)
	})

	t.Run("should fail if failed to configure Compass Connection CR", func(t *testing.T) {
		// given
		compassConfiguratorMock := &mocks.CompassConfigurator{}
		certificateSecretConfiguratorMock := &mocks.CertificateSecretConfigurator{}
		configurationSecretConfiguratorMock := &mocks.ConfigurationSecretConfigurator{}
		compassConnectionConfiguratorMock := &mocks.CompassConnectionConfigurator{}
		deploymentConfiguratorMock := &mocks.DeploymentConfigurator{}

		compassConfiguratorRollbackFunc := RollbackFuncTest{}
		certificateSecretsRollbackFunc := RollbackFuncTest{}
		configurationSecretRollbackFunc := RollbackFuncTest{}
		compassConnectionRollbackFunc := RollbackFuncTest{}
		deploymentRollbackFunc := RollbackFuncTest{}

		config := types.CompassRuntimeAgentConfig{
			ConnectorUrl: connectorURL,
			RuntimeID:    runtimeID,
			Token:        token,
			Tenant:       tenant,
		}

		compassConfiguratorMock.On("Do", runtimeName, formationName).Return(config, compassConfiguratorRollbackFunc.Func(), nil)
		certificateSecretConfiguratorMock.On("Do", NewCACertSecretName, NewClientCertSecretName).Return(certificateSecretsRollbackFunc.Func(), nil)
		configurationSecretConfiguratorMock.On("Do", NewCompassRuntimeConfigName, config).Return(configurationSecretRollbackFunc.Func(), nil)
		compassConnectionConfiguratorMock.On("Do").Return(compassConnectionRollbackFunc.Func(), errors.New("some error"))
		deploymentConfiguratorMock.On("Do",
			fmt.Sprintf("%s/%s", IstioSystemNamespace, NewCACertSecretName),
			fmt.Sprintf("%s/%s", CompassSystemNamespace, NewClientCertSecretName),
			fmt.Sprintf("%s/%s", CompassSystemNamespace, NewCompassRuntimeConfigName),
			fmt.Sprintf("%s", NewControllerSyncPeriodTime)).
			Return(deploymentRollbackFunc.Func(), nil)

		configurator := NewCompassRuntimeAgentConfigurator(compassConfiguratorMock, certificateSecretConfiguratorMock, configurationSecretConfiguratorMock, compassConnectionConfiguratorMock, deploymentConfiguratorMock, "tenant")

		// when
		rollbackFunc, err := configurator.Do(runtimeName, formationName)

		// then
		require.Error(t, err)
		require.Nil(t, rollbackFunc)
		certificateSecretConfiguratorMock.AssertExpectations(t)
		compassConnectionConfiguratorMock.AssertExpectations(t)
		deploymentConfiguratorMock.AssertExpectations(t)
		require.True(t, compassConfiguratorRollbackFunc.invoked)
		require.True(t, certificateSecretsRollbackFunc.invoked)
		require.True(t, configurationSecretRollbackFunc.invoked)
		//require.True(t, compassConnectionRollbackFunc.invoked)
		require.True(t, deploymentRollbackFunc.invoked)
	})
}

type RollbackFuncTest struct {
	invoked bool
}

func (rfc *RollbackFuncTest) Func() types.RollbackFunc {
	return func() error {
		rfc.invoked = true
		return nil
	}
}
