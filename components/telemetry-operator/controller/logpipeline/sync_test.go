package logpipeline

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes/mocks"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/utils/envvar"

	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

func TestSyncSectionsConfigMapClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := newSyncer(mockClient, testConfig)

	lp := telemetryv1alpha1.LogPipeline{}
	result, err := sut.syncSectionsConfigMap(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}

func TestSyncFilesConfigMapErrorClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := newSyncer(mockClient, testConfig)

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

	mockClient := &mocks.Client{}
	sut := newSyncer(mockClient, testConfig)
	sut.syncUnsupportedPluginsTotal(logPipelines)
	require.Equal(t, 2, sut.unsupportedPluginsTotal)
}

func TestSyncVariablesFromHttpOutput(t *testing.T) {
	s := scheme.Scheme
	err := telemetryv1alpha1.AddToScheme(s)
	require.NoError(t, err)

	secretData := map[string][]byte{
		"host": []byte("my-host"),
	}
	referencedSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "referenced-secret",
			Namespace: "default",
		},
		Data: secretData,
	}
	require.NoError(t, err)

	secretKeyRef := telemetryv1alpha1.SecretKeyRef{
		Name:      "referenced-secret",
		Key:       "host",
		Namespace: "default",
	}
	lp := telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "my-pipeline"},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				HTTP: &telemetryv1alpha1.HTTPOutput{
					Host: telemetryv1alpha1.ValueType{
						ValueFrom: &telemetryv1alpha1.ValueFromSource{
							SecretKey: secretKeyRef,
						},
					},
				},
			},
		},
	}
	logPipelines := telemetryv1alpha1.LogPipelineList{
		Items: []telemetryv1alpha1.LogPipeline{lp},
	}
	mockClient := fake.NewClientBuilder().WithScheme(s).WithObjects(&referencedSecret).Build()

	sut := newSyncer(mockClient, testConfig)
	restartRequired, err := sut.syncVariables(context.Background(), &logPipelines)
	require.NoError(t, err)
	require.True(t, restartRequired)

	var envSecret corev1.Secret
	err = mockClient.Get(context.Background(), types.NamespacedName{Name: "test-telemetry-fluent-bit-env", Namespace: "default"}, &envSecret)
	require.NoError(t, err)
	targetSecretKey := envvar.GenerateName("my-pipeline", secretKeyRef)
	require.Equal(t, []byte("my-host"), envSecret.Data[targetSecretKey])
}
