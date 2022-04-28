package sync

import (
	"context"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/sync/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

var (
	sectionCm = types.NamespacedName{Name: "section-cm", Namespace: "cm-ns"}
	parsersCm = types.NamespacedName{Name: "parsers-cm", Namespace: "cm-ns"}
	filesCm   = types.NamespacedName{Name: "files-cm", Namespace: "cm-ns"}
	envSecret = types.NamespacedName{Name: "env-secret", Namespace: "cm-ns"}
)

func TestGetOrCreateWithConfigMapIsNotFoundCreatesNewWithGivenNamespacedNameAndNoError(t *testing.T) {
	mockClient := &mocks.Client{}
	notFoundErr := errors.NewNotFound(schema.GroupResource{}, "")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(notFoundErr)
	mockClient.On("Create", mock.Anything, mock.Anything).Return(nil)
	sut := NewLogPipelineSyncer(mockClient, sectionCm, parsersCm, filesCm, envSecret)

	cm := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "some-cm", Namespace: "cm-ns"}}
	err := sut.getOrCreate(context.Background(), &cm)

	require.NoError(t, err)
	require.Equal(t, "some-cm", cm.Name)
	require.Equal(t, "cm-ns", cm.Namespace)
}

func TestGetOrCreateWithConfigMapAnyOtherErrorPropagates(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, sectionCm, parsersCm, filesCm, envSecret)

	cm := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "some-cm", Namespace: "cm-ns"}}
	err := sut.getOrCreate(context.Background(), &cm)

	require.Error(t, err)
}

func TestGetOrCreateWithSecretIsNotFoundCreatesNewWithGivenNamespacedNameAndNoError(t *testing.T) {
	mockClient := &mocks.Client{}
	notFoundErr := errors.NewNotFound(schema.GroupResource{}, "")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(notFoundErr)
	mockClient.On("Create", mock.Anything, mock.Anything).Return(nil)
	sut := NewLogPipelineSyncer(mockClient, sectionCm, parsersCm, filesCm, envSecret)

	secret := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "some-secret", Namespace: "secret-ns"}}
	err := sut.getOrCreate(context.Background(), &secret)

	require.NoError(t, err)
	require.Equal(t, "some-secret", secret.Name)
	require.Equal(t, "secret-ns", secret.Namespace)
}

func TestGetOrCreateWithSecretAnyOtherErrorPropagates(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, sectionCm, parsersCm, filesCm, envSecret)

	secret := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "some-secret", Namespace: "secret-ns"}}
	err := sut.getOrCreate(context.Background(), &secret)

	require.Error(t, err)
}

func TestSyncSectionsConfigMapClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, sectionCm, parsersCm, filesCm, envSecret)

	lp := telemetryv1alpha1.LogPipeline{}
	result, err := sut.syncSectionsConfigMap(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}

func TestSyncParsersConfigMapErrorClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, sectionCm, parsersCm, filesCm, envSecret)

	lp := telemetryv1alpha1.LogPipeline{}
	result, err := sut.syncParsersConfigMap(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}

func TestSyncFilesConfigMapErrorClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, sectionCm, parsersCm, filesCm, envSecret)

	lp := telemetryv1alpha1.LogPipeline{}
	result, err := sut.syncFilesConfigMap(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}

func TestSyncSecretRefsConfigMapErrorClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, sectionCm, parsersCm, filesCm, envSecret)

	lp := telemetryv1alpha1.LogPipeline{}
	result, err := sut.syncSecretRefs(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}
