package sync

import (
	"context"
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
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
	emitterConfig = fluentbit.EmitterConfig{
		InputTag:    "kube",
		BufferLimit: "10M",
		StorageType: "filesystem",
	}
)

func TestGetOrCreateWithConfigMapIsNotFoundCreatesNewWithGivenNamespacedNameAndNoError(t *testing.T) {
	mockClient := &mocks.Client{}
	notFoundErr := errors.NewNotFound(schema.GroupResource{}, "")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(notFoundErr)
	mockClient.On("Create", mock.Anything, mock.Anything).Return(nil)
	sut := NewLogPipelineSyncer(mockClient, daemonSetConfig, emitterConfig)

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
	sut := NewLogPipelineSyncer(mockClient, daemonSetConfig, emitterConfig)

	cm := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "some-cm", Namespace: "cm-ns"}}
	err := sut.getOrCreate(context.Background(), &cm)

	require.Error(t, err)
}

func TestGetOrCreateWithSecretIsNotFoundCreatesNewWithGivenNamespacedNameAndNoError(t *testing.T) {
	mockClient := &mocks.Client{}
	notFoundErr := errors.NewNotFound(schema.GroupResource{}, "")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(notFoundErr)
	mockClient.On("Create", mock.Anything, mock.Anything).Return(nil)
	sut := NewLogPipelineSyncer(mockClient, daemonSetConfig, emitterConfig)

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
	sut := NewLogPipelineSyncer(mockClient, daemonSetConfig, emitterConfig)

	secret := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "some-secret", Namespace: "secret-ns"}}
	err := sut.getOrCreate(context.Background(), &secret)

	require.Error(t, err)
}

func TestSyncSectionsConfigMapClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, daemonSetConfig, emitterConfig)

	lp := telemetryv1alpha1.LogPipeline{}
	result, err := sut.syncSectionsConfigMap(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}

func TestSyncParsersConfigMapErrorClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, daemonSetConfig, emitterConfig)

	lp := telemetryv1alpha1.LogPipeline{}
	result, err := sut.syncParsersConfigMap(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}

func TestSyncFilesConfigMapErrorClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, daemonSetConfig, emitterConfig)

	lp := telemetryv1alpha1.LogPipeline{}
	result, err := sut.syncFilesConfigMap(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}

func TestSyncSecretRefsConfigMapErrorClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := NewLogPipelineSyncer(mockClient, daemonSetConfig, emitterConfig)

	lp := telemetryv1alpha1.LogPipeline{}
	result, err := sut.syncVariables(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}

func TestFetchSecret(t *testing.T) {
	mockClient := &mocks.Client{}
	retSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fooSecret",
			Namespace: "fooNamespace",
		},
		StringData: map[string]string{"foo": "bar"},
	}
	//mockClient.CoreV1().Secrets("fooNamespace").Create(context.TODO(), &retSecret, metav1.CreateOptions{})
	mockClient.On("Get", context.Background(), types.NamespacedName{Namespace: "fooNamespace", Name: "fooSecret"}, &corev1.Secret{}).Return(nil)
	mockClient.On("fetchSecret", context.Background(), mock.Anything, mock.Anything).Return(retSecret, nil)
	sut := NewLogPipelineSyncer(mockClient, daemonSetConfig, emitterConfig)
	secretKey := telemetryv1alpha1.SecretKeyRef{
		Name:      "foo",
		Namespace: "fooNs",
		Key:       "fooKey",
	}
	fromType := telemetryv1alpha1.VariableReference{Name: "foo", ValueFrom: telemetryv1alpha1.ValueFromType{SecretKey: secretKey}}
	fromLogPipeline := telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Variables: []telemetryv1alpha1.VariableReference{fromType},
		},
	}

	gotSecret, err := sut.syncVariables(context.Background(), &fromLogPipeline)
	require.NoError(t, err)
	t.Log(gotSecret)
}

//func TestSyncSecretRefsSecretsCorrectlyAdded(t *testing.T) {
//	mockClient := &mocks.Client{}
//	secret := corev1.Secret{}
//	referencedSecret := corev1.Secret{
//		ObjectMeta: metav1.ObjectMeta{Name: "foo"},
//		StringData: map[string]string{"sec1": "val1"},
//	}
//	sut := NewLogPipelineSyncer(mockClient, daemonSetConfig, emitterConfig)
//	sut.fetchSecretData(referencedSecret, secret)
//
//}
