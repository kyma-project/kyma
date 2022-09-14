package compassruntimeagentinit

import (
	"context"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/compassruntimeagentinit/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestCompassRuntimeAgentInit(t *testing.T) {
	t.Run("should succeed and return rollback function", func(t *testing.T) {
		// given
		runtimeName := "newRuntime"
		runtimeID := "runtimeID"
		token := "token"
		connectorURL := "www.someurl.com"

		directorMock := &mocks.DirectorClient{}
		fakeClientSet := fake.NewSimpleClientset()

		directorMock.On("RegisterRuntime", runtimeName).Return(runtimeID, nil)
		directorMock.On("GetConnectionToken", runtimeID).Return(token, connectorURL, nil)
		directorMock.On("UnregisterRuntime", runtimeID).Return(nil)

		deployment := v12.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      CompassRuntimeAgentDeployment,
				Namespace: CompassSystemNamespace,
			},
			Spec: v12.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{{
							Env: []v1.EnvVar{{
								Name:  ConfigurationSecretEnvName,
								Value: "defaultnamespace/defaultsecretname",
							}},
						}},
					},
				},
			},
			Status: v12.DeploymentStatus{
				AvailableReplicas: 1,
			},
		}
		_, err := fakeClientSet.AppsV1().Deployments(CompassSystemNamespace).Create(context.Background(), &deployment, metav1.CreateOptions{})
		require.NoError(t, err)

		configurator := NewCompassRuntimeAgentConfigurator(directorMock, fakeClientSet, "tenant")

		// when
		rollbackFunc, err := configurator.Do(runtimeName)

		// then
		require.NoError(t, err)

		// when
		err = rollbackFunc()

		// then
		require.NoError(t, err)
	})

	t.Run("should fail if failed to register runtime", func(t *testing.T) {
		// given
		runtimeName := "newRuntime"

		directorMock := &mocks.DirectorClient{}

		directorMock.On("RegisterRuntime", runtimeName).Return("", errors.New("some error"))

		configurator := NewCompassRuntimeAgentConfigurator(directorMock, nil, "tenant")

		// when
		rollbackFunc, err := configurator.Do(runtimeName)

		// then
		require.Nil(t, rollbackFunc)
		require.Error(t, err)
	})

	t.Run("should fail if failed to get token URL", func(t *testing.T) {
		// given
		runtimeName := "newRuntime"
		runtimeID := "runtimeID"

		directorMock := &mocks.DirectorClient{}
		fakeClientSet := fake.NewSimpleClientset()

		directorMock.On("RegisterRuntime", runtimeName).Return(runtimeID, nil)
		directorMock.On("GetConnectionToken", runtimeID).Return("", "", errors.New("some error"))
		directorMock.On("UnregisterRuntime", runtimeID).Return(nil)

		configurator := NewCompassRuntimeAgentConfigurator(directorMock, fakeClientSet, "tenant")

		// when
		rollbackFunc, err := configurator.Do(runtimeName)

		// then
		require.Nil(t, rollbackFunc)
		require.Error(t, err)
	})

	t.Run("should fail if failed to create secret", func(t *testing.T) {
		// given
		runtimeName := "newRuntime"
		runtimeID := "runtimeID"
		token := "token"
		connectorURL := "www.someurl.com"

		directorMock := &mocks.DirectorClient{}
		fakeClientSet := fake.NewSimpleClientset()

		directorMock.On("RegisterRuntime", runtimeName).Return(runtimeID, nil)
		directorMock.On("GetConnectionToken", runtimeID).Return(token, connectorURL, nil)
		directorMock.On("UnregisterRuntime", runtimeID).Return(nil)

		secret := v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      NewCompassRuntimeConfigName,
				Namespace: CompassSystemNamespace,
			},
			Data: map[string][]byte{},
		}
		_, err := fakeClientSet.CoreV1().Secrets(CompassSystemNamespace).Create(context.Background(), &secret, metav1.CreateOptions{})
		require.NoError(t, err)

		configurator := NewCompassRuntimeAgentConfigurator(directorMock, fakeClientSet, "tenant")

		// when
		rollbackFunc, err := configurator.Do(runtimeName)

		// then
		require.Nil(t, rollbackFunc)
		require.Error(t, err)
	})

	t.Run("should fail if failed to modify deployment", func(t *testing.T) {
		// TODO
		// given
		runtimeName := "newRuntime"
		runtimeID := "runtimeID"
		token := "token"
		connectorURL := "www.someurl.com"

		directorMock := &mocks.DirectorClient{}
		fakeClientSet := fake.NewSimpleClientset()

		directorMock.On("RegisterRuntime", runtimeName).Return(runtimeID, nil)
		directorMock.On("GetConnectionToken", runtimeID).Return(token, connectorURL, nil)
		directorMock.On("UnregisterRuntime", runtimeID).Return(nil)

		configurator := NewCompassRuntimeAgentConfigurator(directorMock, fakeClientSet, "tenant")

		// when
		rollbackFunc, err := configurator.Do(runtimeName)

		// then
		require.Nil(t, rollbackFunc)
		require.Error(t, err)
		//
		//// given
		//runtimeName := "newRuntime"
		//runtimeID := "runtimeID"
		//token := "token"
		//connectorURL := "www.someurl.com"
		//
		//directorMock := &mocks.DirectorClient{}
		//fakeClientSet := fake.NewSimpleClientset()
		//
		//directorMock.On("RegisterRuntime", runtimeName).Return(runtimeID, nil)
		//directorMock.On("GetConnectionToken", runtimeID).Return(token, connectorURL, nil)
		//directorMock.On("UnregisterRuntime", runtimeID).Return(nil)
		//
		//configurator := NewCompassRuntimeAgentConfigurator(directorMock, fakeClientSet, "tenant")
		//
		//// when
		//rollbackFunc, err := configurator.Do(runtimeName)
		//
		//// then
		//require.NoError(t, err)
		//
		//// when
		//err = rollbackFunc()
		//
		//// then
		//require.NoError(t, err)
	})

	t.Run("rollback function should fail if failed to unregister Runtime", func(t *testing.T) {
		// given
		runtimeName := "newRuntime"
		runtimeID := "runtimeID"
		token := "token"
		connectorURL := "www.someurl.com"

		directorMock := &mocks.DirectorClient{}
		fakeClientSet := fake.NewSimpleClientset()

		directorMock.On("RegisterRuntime", runtimeName).Return(runtimeID, nil)
		directorMock.On("GetConnectionToken", runtimeID).Return(token, connectorURL, nil)
		directorMock.On("UnregisterRuntime", runtimeID).Return(errors.New("some error"))

		configurator := NewCompassRuntimeAgentConfigurator(directorMock, fakeClientSet, "tenant")

		// when
		rollbackFunc, err := configurator.Do(runtimeName)

		// then
		require.NoError(t, err)

		// when
		err = rollbackFunc()

		// then
		require.Error(t, err)
	})

	t.Run("rollback function should fail if failed to delete configuration secret", func(t *testing.T) {
		// given
		runtimeName := "newRuntime"
		runtimeID := "runtimeID"
		token := "token"
		connectorURL := "www.someurl.com"

		directorMock := &mocks.DirectorClient{}
		fakeClientSet := fake.NewSimpleClientset()

		directorMock.On("RegisterRuntime", runtimeName).Return(runtimeID, nil)
		directorMock.On("GetConnectionToken", runtimeID).Return(token, connectorURL, nil)
		directorMock.On("UnregisterRuntime", runtimeID).Return(nil)

		configurator := NewCompassRuntimeAgentConfigurator(directorMock, fakeClientSet, "tenant")

		// when
		rollbackFunc, err := configurator.Do(runtimeName)

		// then
		require.NoError(t, err)

		// when
		err = fakeClientSet.CoreV1().Secrets(CompassSystemNamespace).Delete(context.Background(), NewCompassRuntimeConfigName, metav1.DeleteOptions{})

		// then
		require.NoError(t, err)

		// when
		err = rollbackFunc()

		// then
		require.Error(t, err)
	})
}
