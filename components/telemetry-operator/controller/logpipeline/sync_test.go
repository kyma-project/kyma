package logpipeline

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes/mocks"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/utils/envvar"

	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

func TestSyncSectionsConfigMap(t *testing.T) {
	sectionsCmName := types.NamespacedName{Name: "cm", Namespace: "kyma-system"}
	fakeClient := fake.NewClientBuilder().WithObjects(
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sectionsCmName.Name,
				Namespace: sectionsCmName.Namespace,
			},
		}).Build()

	t.Run("should add section during first sync", func(t *testing.T) {
		sut := newSyncer(fakeClient, Config{SectionsConfigMap: sectionsCmName})

		pipeline := &telemetryv1alpha1.LogPipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name: "dummy",
			},
			Spec: telemetryv1alpha1.LogPipelineSpec{
				Output: telemetryv1alpha1.Output{
					Custom: `
name  dummy
alias foo`,
				},
			},
		}
		result, err := sut.syncSectionsConfigMap(context.Background(), pipeline)
		require.NoError(t, err)
		require.True(t, result)
		require.True(t, controllerutil.ContainsFinalizer(pipeline, sectionsConfigMapFinalizer))

		var sectionsCm corev1.ConfigMap
		err = fakeClient.Get(context.Background(), sectionsCmName, &sectionsCm)
		require.NoError(t, err)
		require.Contains(t, sectionsCm.Data, "dummy.conf")
		require.Contains(t, sectionsCm.Data["dummy.conf"], "foo")
	})

	t.Run("should update section during subsequent sync", func(t *testing.T) {
		sut := newSyncer(fakeClient, Config{SectionsConfigMap: sectionsCmName})

		pipeline := &telemetryv1alpha1.LogPipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name: "dummy",
			},
			Spec: telemetryv1alpha1.LogPipelineSpec{
				Output: telemetryv1alpha1.Output{
					Custom: `
name  dummy
alias foo`,
				},
			},
		}

		_, err := sut.syncSectionsConfigMap(context.Background(), pipeline)
		require.NoError(t, err)

		pipeline.Spec.Output.Custom = `
name  dummy
alias bar`
		result, err := sut.syncSectionsConfigMap(context.Background(), pipeline)
		require.NoError(t, err)
		require.True(t, result)
		require.True(t, controllerutil.ContainsFinalizer(pipeline, sectionsConfigMapFinalizer))

		var sectionsCm corev1.ConfigMap
		err = fakeClient.Get(context.Background(), sectionsCmName, &sectionsCm)
		require.NoError(t, err)
		require.Contains(t, sectionsCm.Data, "dummy.conf")
		require.NotContains(t, sectionsCm.Data["dummy.conf"], "foo")
		require.Contains(t, sectionsCm.Data["dummy.conf"], "bar")
	})

	t.Run("should remove section if marked for deletion", func(t *testing.T) {
		sut := newSyncer(fakeClient, Config{SectionsConfigMap: sectionsCmName})

		pipeline := &telemetryv1alpha1.LogPipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name: "dummy",
			},
			Spec: telemetryv1alpha1.LogPipelineSpec{
				Output: telemetryv1alpha1.Output{
					Custom: `
name  dummy
alias foo`,
				},
			},
		}

		_, err := sut.syncSectionsConfigMap(context.Background(), pipeline)
		require.NoError(t, err)

		now := metav1.Now()
		pipeline.SetDeletionTimestamp(&now)
		result, err := sut.syncSectionsConfigMap(context.Background(), pipeline)
		require.NoError(t, err)
		require.True(t, result)
		require.False(t, controllerutil.ContainsFinalizer(pipeline, sectionsConfigMapFinalizer))

		var sectionsCm corev1.ConfigMap
		err = fakeClient.Get(context.Background(), sectionsCmName, &sectionsCm)
		require.NoError(t, err)
		require.NotContains(t, sectionsCm.Data, "dummy.conf")
	})

	t.Run("should fail if client fails", func(t *testing.T) {
		badReqClient := &mocks.Client{}
		badReqErr := errors.NewBadRequest("")
		badReqClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
		sut := newSyncer(badReqClient, testConfig)

		lp := telemetryv1alpha1.LogPipeline{}
		result, err := sut.syncFilesConfigMap(context.Background(), &lp)

		require.Error(t, err)
		require.Equal(t, result, false)
	})
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
							SecretKeyRef: &secretKeyRef,
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
	restartRequired, err := sut.syncReferencedSecrets(context.Background(), &logPipelines)
	require.NoError(t, err)
	require.True(t, restartRequired)

	var envSecret corev1.Secret
	err = mockClient.Get(context.Background(), types.NamespacedName{Name: "test-telemetry-fluent-bit-env", Namespace: "default"}, &envSecret)
	require.NoError(t, err)
	targetSecretKey := envvar.GenerateName("my-pipeline", secretKeyRef)
	require.Equal(t, []byte("my-host"), envSecret.Data[targetSecretKey])
}
