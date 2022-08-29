package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes/mocks"
)

func TestGetOrCreateConfigMapError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)

	mockGetterOrCreator := NewGetterOrCreator(mockClient)

	configMapName := types.NamespacedName{Name: "some-cm", Namespace: "cm-ns"}
	_, err := mockGetterOrCreator.ConfigMap(context.Background(), configMapName)

	require.Error(t, err)
	require.Equal(t, badReqErr, err)
}

func TestGetOrCreateConfigMapGetSuccess(t *testing.T) {
	mockClient := &mocks.Client{}
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockGetterOrCreator := NewGetterOrCreator(mockClient)

	configMapName := types.NamespacedName{Name: "some-cm", Namespace: "cm-ns"}
	cm, err := mockGetterOrCreator.ConfigMap(context.Background(), configMapName)

	require.NoError(t, err)
	require.Equal(t, "some-cm", cm.Name)
	require.Equal(t, "cm-ns", cm.Namespace)
}

func TestGetOrCreateConfigMapCreateSuccess(t *testing.T) {
	mockClient := &mocks.Client{}
	notFoundErr := errors.NewNotFound(schema.GroupResource{}, "")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(notFoundErr)
	mockClient.On("Create", mock.Anything, mock.Anything).Return(nil)

	mockGetterOrCreator := NewGetterOrCreator(mockClient)

	configMapName := types.NamespacedName{Name: "some-cm", Namespace: "cm-ns"}
	cm, err := mockGetterOrCreator.ConfigMap(context.Background(), configMapName)

	require.NoError(t, err)
	require.Equal(t, "some-cm", cm.Name)
	require.Equal(t, "cm-ns", cm.Namespace)
}

func TestGetOrCreateSecretError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)

	mockGetterOrCreator := NewGetterOrCreator(mockClient)

	secretName := types.NamespacedName{Name: "some-secret", Namespace: "secret-ns"}
	_, err := mockGetterOrCreator.Secret(context.Background(), secretName)

	require.Error(t, err)
	require.Equal(t, badReqErr, err)
}

func TestGetOrCreateSecretSuccess(t *testing.T) {
	mockClient := &mocks.Client{}
	notFoundErr := errors.NewNotFound(schema.GroupResource{}, "")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(notFoundErr)
	mockClient.On("Create", mock.Anything, mock.Anything).Return(nil)

	mockGetterOrCreator := NewGetterOrCreator(mockClient)

	secretName := types.NamespacedName{Name: "some-secret", Namespace: "secret-ns"}
	secret, err := mockGetterOrCreator.Secret(context.Background(), secretName)

	require.NoError(t, err)
	require.Equal(t, "some-secret", secret.Name)
	require.Equal(t, "secret-ns", secret.Namespace)
}
