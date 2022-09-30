package compassruntimeagentinit

import (
	"context"
	ccv1 "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/compassruntimeagentinit/types"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/compassruntimeagentinit/types/mocks"
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
		directorMock := &mocks.DirectorClient{}
		certificateSecretConfiguratorMock := &mocks.CertificateSecretConfigurator{}
		configurationSecretConfiguratorMock := &mocks.ConfigurationSecretConfigurator{}
		compassConnectionConfiguratorMock := &mocks.CompassConnectionConfigurator{}
		deploymentConfiguratorMock := &mocks.DeploymentConfigurator{}

		certificateSecretsRollbackFunc := RollbackFuncTest{}
		configurationSecretRollbackFunc := RollbackFuncTest{}
		compassConnectionRollbackFunc := RollbackFuncTest{}
		deploymentRollbackFunc := RollbackFuncTest{}

		runtimeName := "newRuntime"
		runtimeID := "runtimeID"
		token := "token"
		connectorURL := "www.someurl.com"
		tenant := "tenant"
		formationName := "newFormation"

		config := types.CompassRuntimeAgentConfig{
			ConnectorUrl: connectorURL,
			RuntimeID:    runtimeID,
			Token:        token,
			Tenant:       tenant,
		}

		directorMock.On("RegisterRuntime", runtimeName).Return(runtimeID, nil)
		directorMock.On("RegisterFormation", formationName).Return(nil)
		directorMock.On("AssignRuntimeToFormation", runtimeID, formationName).Return(nil)

		directorMock.On("GetConnectionToken", runtimeID).Return(token, connectorURL, nil)

		directorMock.On("UnregisterRuntime", runtimeID).Return(nil)
		directorMock.On("UnregisterFormation", formationName).Return(nil)

		certificateSecretConfiguratorMock.On("Do", NewCACertSecretName, NewClientCertSecretName).Return(certificateSecretsRollbackFunc.Func(), nil)
		configurationSecretConfiguratorMock.On("Do", NewCompassRuntimeConfigName, config).Return(configurationSecretRollbackFunc.Func(), nil)
		compassConnectionConfiguratorMock.On("Do").Return(compassConnectionRollbackFunc.Func(), nil)
		deploymentConfiguratorMock.On("Do",
			"test/"+NewCACertSecretName,
			"test/"+NewClientCertSecretName,
			"test/"+NewCompassRuntimeConfigName).Return(deploymentRollbackFunc.Func(), nil)

		configurator := NewCompassRuntimeAgentConfigurator(directorMock, certificateSecretConfiguratorMock, configurationSecretConfiguratorMock, compassConnectionConfiguratorMock, deploymentConfiguratorMock, "tenant", "test")

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
		directorMock.AssertExpectations(t)
		require.True(t, certificateSecretsRollbackFunc.invoked)
		require.True(t, configurationSecretRollbackFunc.invoked)
		require.True(t, compassConnectionRollbackFunc.invoked)
		require.True(t, deploymentRollbackFunc.invoked)
	})

	//t.Run("should fail if failed to register runtime", func(t *testing.T) {
	//	// given
	//	runtimeName := "newRuntime"
	//
	//	directorMock := &mocks.DirectorClient{}
	//
	//	directorMock.On("RegisterRuntime", runtimeName).Return("", errors.New("some error"))
	//
	//	configurator := NewCompassRuntimeAgentConfigurator(directorMock, nil, nil, "tenant")
	//
	//	// when
	//	rollbackFunc, err := configurator.Do(runtimeName)
	//
	//	// then
	//	require.Nil(t, rollbackFunc)
	//	require.Error(t, err)
	//})

	//t.Run("should fail if failed to get token URL", func(t *testing.T) {
	//	// given
	//	runtimeName := "newRuntime"
	//	runtimeID := "runtimeID"
	//
	//	directorMock := &mocks.DirectorClient{}
	//	fakeClientSet := fake.NewSimpleClientset()
	//
	//	directorMock.On("RegisterRuntime", runtimeName).Return(runtimeID, nil)
	//	directorMock.On("GetConnectionToken", runtimeID).Return("", "", errors.New("some error"))
	//	directorMock.On("UnregisterRuntime", runtimeID).Return(nil)
	//
	//	configurator := NewCompassRuntimeAgentConfigurator(directorMock, fakeClientSet, nil, "tenant")
	//
	//	// when
	//	rollbackFunc, err := configurator.Do(runtimeName)
	//
	//	// then
	//	require.Nil(t, rollbackFunc)
	//	require.Error(t, err)
	//})
	//
	//t.Run("should fail if failed to create secret", func(t *testing.T) {
	//	// given
	//	runtimeName := "newRuntime"
	//	runtimeID := "runtimeID"
	//	token := "token"
	//	connectorURL := "www.someurl.com"
	//
	//	directorMock := &mocks.DirectorClient{}
	//	fakeClientSet := fake.NewSimpleClientset()
	//
	//	directorMock.On("RegisterRuntime", runtimeName).Return(runtimeID, nil)
	//	directorMock.On("GetConnectionToken", runtimeID).Return(token, connectorURL, nil)
	//	directorMock.On("UnregisterRuntime", runtimeID).Return(nil)
	//
	//	secret := v1.Secret{
	//		ObjectMeta: metav1.ObjectMeta{
	//			Name:      NewCompassRuntimeConfigName,
	//			Namespace: CompassSystemNamespace,
	//		},
	//		Data: map[string][]byte{},
	//	}
	//	_, err := fakeClientSet.CoreV1().Secrets(CompassSystemNamespace).Create(context.Background(), &secret, metav1.CreateOptions{})
	//	require.NoError(t, err)
	//
	//	configurator := NewCompassRuntimeAgentConfigurator(directorMock, fakeClientSet, nil, "tenant")
	//
	//	// when
	//	rollbackFunc, err := configurator.Do(runtimeName)
	//
	//	// then
	//	require.Nil(t, rollbackFunc)
	//	require.Error(t, err)
	//})

	//t.Run("should fail if failed to modify deployment", func(t *testing.T) {
	//	// given
	//	runtimeName := "newRuntime"
	//	runtimeID := "runtimeID"
	//	token := "token"
	//	connectorURL := "www.someurl.com"
	//
	//	directorMock := &mocks.DirectorClient{}
	//	fakeClientSet := fake.NewSimpleClientset()
	//
	//	directorMock.On("RegisterRuntime", runtimeName).Return(runtimeID, nil)
	//	directorMock.On("GetConnectionToken", runtimeID).Return(token, connectorURL, nil)
	//	directorMock.On("UnregisterRuntime", runtimeID).Return(nil)
	//
	//	deployment := v12.Deployment{
	//		ObjectMeta: metav1.ObjectMeta{
	//			Name:      CompassRuntimeAgentDeployment,
	//			Namespace: CompassSystemNamespace,
	//			Annotations: map[string]string{
	//				"note": "no envs in the container",
	//			},
	//		},
	//		Spec: v12.DeploymentSpec{
	//			Template: v1.PodTemplateSpec{
	//				Spec: v1.PodSpec{
	//					Containers: []v1.Container{{}},
	//				},
	//			},
	//		},
	//		Status: v12.DeploymentStatus{
	//			AvailableReplicas: 1,
	//		},
	//	}
	//	_, err := fakeClientSet.AppsV1().Deployments(CompassSystemNamespace).Create(context.Background(), &deployment, metav1.CreateOptions{})
	//	require.NoError(t, err)
	//
	//	configurator := NewCompassRuntimeAgentConfigurator(directorMock, fakeClientSet, nil, "tenant")
	//
	//	// when
	//	rollbackFunc, err := configurator.Do(runtimeName)
	//
	//	// then
	//	require.Nil(t, rollbackFunc)
	//	require.Error(t, err)
	//})
	//
	//t.Run("rollback function should fail if failed to unregister Runtime", func(t *testing.T) {
	//	// given
	//	runtimeName := "newRuntime"
	//	runtimeID := "runtimeID"
	//	token := "token"
	//	connectorURL := "www.someurl.com"
	//
	//	directorMock := &mocks.DirectorClient{}
	//	fakeClientSet := fake.NewSimpleClientset()
	//
	//	directorMock.On("RegisterRuntime", runtimeName).Return(runtimeID, nil)
	//	directorMock.On("GetConnectionToken", runtimeID).Return(token, connectorURL, nil)
	//	directorMock.On("UnregisterRuntime", runtimeID).Return(errors.New("some error"))
	//
	//	err := createCRADeployment(fakeClientSet)
	//	require.NoError(t, err)
	//
	//	configurator := NewCompassRuntimeAgentConfigurator(directorMock, fakeClientSet, nil, "tenant")
	//
	//	// when
	//	rollbackFunc, err := configurator.Do(runtimeName)
	//
	//	// then
	//	require.NoError(t, err)
	//
	//	// when
	//	err = rollbackFunc()
	//
	//	// then
	//	require.Error(t, err)
	//})
	//
	//t.Run("rollback function should fail if failed to delete configuration secret", func(t *testing.T) {
	//	// given
	//	runtimeName := "newRuntime"
	//	runtimeID := "runtimeID"
	//	token := "token"
	//	connectorURL := "www.someurl.com"
	//
	//	directorMock := &mocks.DirectorClient{}
	//	fakeClientSet := fake.NewSimpleClientset()
	//
	//	directorMock.On("RegisterRuntime", runtimeName).Return(runtimeID, nil)
	//	directorMock.On("GetConnectionToken", runtimeID).Return(token, connectorURL, nil)
	//	directorMock.On("UnregisterRuntime", runtimeID).Return(nil)
	//
	//	err := createCRADeployment(fakeClientSet)
	//	require.NoError(t, err)
	//
	//	configurator := NewCompassRuntimeAgentConfigurator(directorMock, fakeClientSet, nil, "tenant")
	//
	//	// when
	//	rollbackFunc, err := configurator.Do(runtimeName)
	//
	//	// then
	//	require.NoError(t, err)
	//
	//	// when
	//	err = fakeClientSet.CoreV1().Secrets(CompassSystemNamespace).Delete(context.Background(), NewCompassRuntimeConfigName, metav1.DeleteOptions{})
	//
	//	// then
	//	require.NoError(t, err)
	//
	//	// when
	//	err = rollbackFunc()
	//
	//	// then
	//	require.Error(t, err)
	//})
}

func createCRADeployment(fakeClientSet *fake.Clientset) error {
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
	return err
}

func createCompassConnection() *ccv1.CompassConnection {
	return &ccv1.CompassConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name: "compass-connection",
		},
	}
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
