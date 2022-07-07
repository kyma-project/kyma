package sync

import (
	"context"
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/sync/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var (
	daemonSetConfig = FluentBitDaemonSetConfig{
		FluentBitDaemonSetName:     types.NamespacedName{Name: "telemetry-fluent-bit", Namespace: "cm-ns"},
		FluentBitSectionsConfigMap: types.NamespacedName{Name: "section-cm", Namespace: "cm-ns"},
		FluentBitParsersConfigMap:  types.NamespacedName{Name: "parsers-cm", Namespace: "cm-ns"},
		FluentBitFilesConfigMap:    types.NamespacedName{Name: "files-cm", Namespace: "cm-ns"},
		FluentBitEnvSecret:         types.NamespacedName{Name: "env-secret", Namespace: "cm-ns"},
	}
	pipelineConfig = fluentbit.PipelineConfig{
		InputTag:          "kube",
		MemoryBufferLimit: "10M",
		StorageType:       "filesystem",
		FsBufferLimit:     "1G",
	}
)

func TestGetOrCreateWithConfigMapIsNotFoundCreatesNewWithGivenNamespacedNameAndNoError(t *testing.T) {
	mockClient := &mocks.Client{}
	notFoundErr := errors.NewNotFound(schema.GroupResource{}, "")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(notFoundErr)
	mockClient.On("Create", mock.Anything, mock.Anything).Return(nil)
	sut := NewLogPipelineSyncer(mockClient, daemonSetConfig, pipelineConfig)

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
	sut := NewLogPipelineSyncer(mockClient, daemonSetConfig, pipelineConfig)

	cm := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "some-cm", Namespace: "cm-ns"}}
	err := sut.getOrCreate(context.Background(), &cm)

	require.Error(t, err)
}

func TestGetOrCreateWithSecretIsNotFoundCreatesNewWithGivenNamespacedNameAndNoError(t *testing.T) {
	mockClient := &mocks.Client{}
	notFoundErr := errors.NewNotFound(schema.GroupResource{}, "")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(notFoundErr)
	mockClient.On("Create", mock.Anything, mock.Anything).Return(nil)
	sut := NewLogPipelineSyncer(mockClient, daemonSetConfig, pipelineConfig)

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
	sut := NewLogPipelineSyncer(mockClient, daemonSetConfig, pipelineConfig)

	secret := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "some-secret", Namespace: "secret-ns"}}
	err := sut.getOrCreate(context.Background(), &secret)

	require.Error(t, err)
}

func TestSyncSectionsConfigMapClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, daemonSetConfig, pipelineConfig)

	lp := telemetryv1alpha1.LogPipeline{}
	result, err := sut.syncSectionsConfigMap(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}

func TestSyncParsersConfigMapErrorClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, daemonSetConfig, pipelineConfig)

	lp := telemetryv1alpha1.LogPipeline{}
	result, err := sut.syncParsersConfigMap(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}

func TestSyncFilesConfigMapErrorClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, daemonSetConfig, pipelineConfig)

	lp := telemetryv1alpha1.LogPipeline{}
	result, err := sut.syncFilesConfigMap(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}

func TestUnsupportedTotal(t *testing.T) {
	l1OpCustom := `Name  foo
Alias  bar`
	l2FilterCustom1 := telemetryv1alpha1.Filter{
		Custom: `Name  filter1`,
	}
	l2FilterCustom2 := telemetryv1alpha1.Filter{
		Custom: `Name  filter2`,
	}
	l1 := telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "l1"},
		Spec:       telemetryv1alpha1.LogPipelineSpec{Output: telemetryv1alpha1.Output{Custom: l1OpCustom}},
	}
	l2 := telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "l2"},
		Spec:       telemetryv1alpha1.LogPipelineSpec{Filters: []telemetryv1alpha1.Filter{l2FilterCustom1, l2FilterCustom2}},
	}
	logPipelines := &telemetryv1alpha1.LogPipelineList{Items: []telemetryv1alpha1.LogPipeline{l1, l2}}
	res := updateUnsupportedPluginsTotal(logPipelines)
	require.Equal(t, 2, res)
}
