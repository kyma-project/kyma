package sync

import (
	"context"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/sync/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

var (
	sectionCm = types.NamespacedName{Name: "section-cm", Namespace: "cm-ns"}
	parsersCm = types.NamespacedName{Name: "parsers-cm", Namespace: "cm-ns"}
	filesCm   = types.NamespacedName{Name: "files-cm", Namespace: "cm-ns"}
	daemonSet = types.NamespacedName{Name: "fb-ds", Namespace: "cm-ns"}
	secret    = types.NamespacedName{Name: "env-secret", Namespace: "cm-ns"}
)

func TestGetOrCreateConfigMapIsNotFoundCreatesNewConfigMapAndReturnsNoError(t *testing.T) {
	mockClient := &mocks.Client{}
	notFoundErr := errors.NewNotFound(schema.GroupResource{}, "")
	someCm := types.NamespacedName{Name: "some-cm", Namespace: "cm-ns"}
	mockClient.On("Get", mock.Anything, someCm, mock.Anything).Return(notFoundErr)
	mockClient.On("Create", mock.Anything, mock.Anything).Return(nil)
	sut := NewLogPipelineSyncer(mockClient, sectionCm, parsersCm, filesCm, daemonSet, secret)

	result, err := sut.getOrCreateConfigMap(context.Background(), someCm)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, someCm.Name, result.Name)
	require.Equal(t, someCm.Namespace, result.Namespace)
}

func TestGetOrCreateConfigMapAnyOtherErrorReturnsEmptyConfigMap(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, sectionCm, parsersCm, filesCm, daemonSet, secret)

	someCm := types.NamespacedName{Name: "some-cm", Namespace: "cm-ns"}
	result, err := sut.getOrCreateConfigMap(context.Background(), someCm)

	require.Error(t, err)
	require.NotNil(t, result)
	require.Empty(t, result.Name)
	require.Empty(t, result.Namespace)
}

func TestGetOrCreateSecretIsNotFoundCreatesNewSecretAndReturnsNoError(t *testing.T) {
	mockClient := &mocks.Client{}
	notFoundErr := errors.NewNotFound(schema.GroupResource{}, "")
	someSecret := types.NamespacedName{Name: "some-secret", Namespace: "secret-ns"}
	mockClient.On("Get", mock.Anything, someSecret, mock.Anything).Return(notFoundErr)
	mockClient.On("Create", mock.Anything, mock.Anything).Return(nil)
	sut := NewLogPipelineSyncer(mockClient, sectionCm, parsersCm, filesCm, daemonSet, secret)

	result, err := sut.getOrCreateSecret(context.Background(), someSecret)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, someSecret.Name, result.Name)
	require.Equal(t, someSecret.Namespace, result.Namespace)
}

func TestGetOrCreateSecretAnyOtherErrorReturnsEmptySecret(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, sectionCm, parsersCm, filesCm, daemonSet, secret)

	someSecret := types.NamespacedName{Name: "some-secret", Namespace: "secret-ns"}
	result, err := sut.getOrCreateSecret(context.Background(), someSecret)

	require.Error(t, err)
	require.NotNil(t, result)
	require.Empty(t, result.Name)
	require.Empty(t, result.Namespace)
}

func TestSyncSectionsConfigMapClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, sectionCm, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, sectionCm, parsersCm, filesCm, daemonSet, secret)

	lp := telemetryv1alpha1.LogPipeline{}
	result, err := sut.syncSectionsConfigMap(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}

func TestSyncParsersConfigMapErrorClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, parsersCm, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, sectionCm, parsersCm, filesCm, daemonSet, secret)

	lp := telemetryv1alpha1.LogPipeline{}
	result, err := sut.syncParsersConfigMap(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}

func TestSyncFilesConfigMapErrorClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, filesCm, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, sectionCm, parsersCm, filesCm, daemonSet, secret)

	lp := telemetryv1alpha1.LogPipeline{}
	result, err := sut.syncFilesConfigMap(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}

func TestSyncSecretRefsConfigMapErrorClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, secret, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, sectionCm, parsersCm, filesCm, daemonSet, secret)

	lp := telemetryv1alpha1.LogPipeline{}
	result, err := sut.syncSecretRefs(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}
