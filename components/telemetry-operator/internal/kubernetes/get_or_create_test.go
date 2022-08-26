package kubernetes

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func TestGetOrCreateWithConfigMapIsNotFoundCreatesNewWithGivenNamespacedNameAndNoError(t *testing.T) {
	mockClient := &mocks.Client{}
	notFoundErr := errors.NewNotFound(schema.GroupResource{}, "")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(notFoundErr)
	mockClient.On("Create", mock.Anything, mock.Anything).Return(nil)

	mockGetterOrCreator := NewGetterOrCreator(mockClient)

	cm := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "some-cm", Namespace: "cm-ns"}}
	err := mockGetterOrCreator.Object(context.Background(), &cm)

	require.NoError(t, err)
	require.Equal(t, "some-cm", cm.Name)
	require.Equal(t, "cm-ns", cm.Namespace)
}

func TestGetOrCreateWithConfigMapAnyOtherErrorPropagates(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)

	mockGetterOrCreator := NewGetterOrCreator(mockClient)

	configMapName := types.NamespacedName{Name: "some-cm", Namespace: "cm-ns"}
	_, err := mockGetterOrCreator.ConfigMap(context.Background(), configMapName)

	require.Error(t, err)
	require.Equal(t, badReqErr, err)
}

func TestGetOrCreateWithSecretIsNotFoundCreatesNewWithGivenNamespacedNameAndNoError(t *testing.T) {
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

func TestGetOrCreateWithSecretAnyOtherErrorPropagates(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)

	mockGetterOrCreator := NewGetterOrCreator(mockClient)

	secret := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "some-secret", Namespace: "secret-ns"}}
	err := mockGetterOrCreator.Object(context.Background(), &secret)

	require.Error(t, err)
}
