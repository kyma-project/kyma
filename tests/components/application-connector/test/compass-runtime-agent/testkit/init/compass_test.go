package init

import (
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/init/types"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/init/types/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCompassConfigurator(t *testing.T) {
	runtimeName := "runtime"
	runtimeID := "runtimeID"
	formationName := "formation"
	connectionToken := "token"
	connectorURL := "connector.com"
	tenant := "tenant"

	t.Run("should register Runtime, Formation and get connection token", func(t *testing.T) {
		// given
		directorClientMock := &mocks.DirectorClient{}
		directorClientMock.On("RegisterRuntime", runtimeName).Return(runtimeID, nil)
		directorClientMock.On("RegisterFormation", formationName).Return(nil)

		directorClientMock.On("UnregisterRuntime", runtimeID).Return(nil)
		directorClientMock.On("UnregisterFormation", formationName).Return(nil)

		directorClientMock.On("GetConnectionToken", runtimeID).Return(connectionToken, connectorURL, nil)
		directorClientMock.On("AssignRuntimeToFormation", runtimeID, formationName).Return(nil)

		// when
		compassConfigurator := NewCompassConfigurator(directorClientMock, tenant)
		require.NotNil(t, compassConfigurator)

		compassRuntimeAgentConfig, rollbackFunc, err := compassConfigurator.Do(runtimeName, formationName)

		// then
		require.NotNil(t, rollbackFunc)
		require.NoError(t, err)
		require.Equal(t, runtimeID, compassRuntimeAgentConfig.RuntimeID)
		require.Equal(t, tenant, compassRuntimeAgentConfig.Tenant)
		require.Equal(t, connectionToken, compassRuntimeAgentConfig.Token)
		require.Equal(t, connectorURL, compassRuntimeAgentConfig.ConnectorUrl)

		// when
		err = rollbackFunc()

		// then
		require.NoError(t, err)
		directorClientMock.AssertExpectations(t)
	})

	t.Run("should fail when failed to register Runtime", func(t *testing.T) {
		// given
		directorClientMock := &mocks.DirectorClient{}
		directorClientMock.On("RegisterRuntime", runtimeName).Return(runtimeID, errors.New("some error"))

		// when
		compassConfigurator := NewCompassConfigurator(directorClientMock, tenant)
		require.NotNil(t, compassConfigurator)

		compassRuntimeAgentConfig, rollbackFunc, err := compassConfigurator.Do(runtimeName, formationName)

		// then
		require.Equal(t, types.CompassRuntimeAgentConfig{}, compassRuntimeAgentConfig)
		require.Nil(t, rollbackFunc)
		require.Error(t, err)
		directorClientMock.AssertExpectations(t)
	})

	t.Run("should fail when failed to register Formation", func(t *testing.T) {
		// given
		directorClientMock := &mocks.DirectorClient{}
		directorClientMock.On("RegisterRuntime", runtimeName).Return(runtimeID, nil)
		directorClientMock.On("RegisterFormation", formationName).Return(errors.New("some error"))
		directorClientMock.On("UnregisterRuntime", runtimeID).Return(nil)

		// when
		compassConfigurator := NewCompassConfigurator(directorClientMock, tenant)
		require.NotNil(t, compassConfigurator)

		compassRuntimeAgentConfig, rollbackFunc, err := compassConfigurator.Do(runtimeName, formationName)

		// then
		require.Equal(t, types.CompassRuntimeAgentConfig{}, compassRuntimeAgentConfig)
		require.NotNil(t, rollbackFunc)
		require.Error(t, err)

		// when
		err = rollbackFunc()

		// then
		require.NoError(t, err)
		directorClientMock.AssertExpectations(t)
	})

	t.Run("should fail when failed to assign Runtime to Formation", func(t *testing.T) {
		// given
		directorClientMock := &mocks.DirectorClient{}
		directorClientMock.On("RegisterRuntime", runtimeName).Return(runtimeID, nil)
		directorClientMock.On("RegisterFormation", formationName).Return(nil)

		directorClientMock.On("UnregisterRuntime", runtimeID).Return(nil)
		directorClientMock.On("UnregisterFormation", formationName).Return(nil)

		directorClientMock.On("AssignRuntimeToFormation", runtimeID, formationName).Return(errors.New("some error"))

		// when
		compassConfigurator := NewCompassConfigurator(directorClientMock, tenant)
		require.NotNil(t, compassConfigurator)

		compassRuntimeAgentConfig, rollbackFunc, err := compassConfigurator.Do(runtimeName, formationName)

		// then
		require.NotNil(t, compassConfigurator)
		require.Equal(t, types.CompassRuntimeAgentConfig{}, compassRuntimeAgentConfig)
		require.NotNil(t, rollbackFunc)
		require.Error(t, err)

		// when
		err = rollbackFunc()

		// then
		require.NoError(t, err)
		directorClientMock.AssertExpectations(t)
	})

	t.Run("should fail when failed to get connection token", func(t *testing.T) {
		// given
		directorClientMock := &mocks.DirectorClient{}
		directorClientMock.On("RegisterRuntime", runtimeName).Return(runtimeID, nil)
		directorClientMock.On("RegisterFormation", formationName).Return(nil)

		directorClientMock.On("UnregisterRuntime", runtimeID).Return(nil)
		directorClientMock.On("UnregisterFormation", formationName).Return(nil)

		directorClientMock.On("AssignRuntimeToFormation", runtimeID, formationName).Return(nil)
		directorClientMock.On("GetConnectionToken", runtimeID).Return("", "", errors.New("some error"))

		// when
		compassConfigurator := NewCompassConfigurator(directorClientMock, tenant)
		require.NotNil(t, compassConfigurator)

		compassRuntimeAgentConfig, rollbackFunc, err := compassConfigurator.Do(runtimeName, formationName)

		// then
		require.NotNil(t, compassConfigurator)
		require.Equal(t, types.CompassRuntimeAgentConfig{}, compassRuntimeAgentConfig)
		require.NotNil(t, rollbackFunc)
		require.Error(t, err)

		// when
		err = rollbackFunc()

		// then
		require.NoError(t, err)
		directorClientMock.AssertExpectations(t)
	})
}
