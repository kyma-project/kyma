package logpipeline

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/controller/logpipeline/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestUpdateStatus(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = telemetryv1alpha1.AddToScheme(scheme)

	t.Run("should add pending condition if some referenced secret does not exist", func(t *testing.T) {
		pipelineName := "pipeline"
		pipeline := &telemetryv1alpha1.LogPipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name: pipelineName,
			},
			Spec: telemetryv1alpha1.LogPipelineSpec{
				Output: telemetryv1alpha1.Output{
					HTTP: &telemetryv1alpha1.HTTPOutput{
						Host: telemetryv1alpha1.ValueType{
							ValueFrom: &telemetryv1alpha1.ValueFromSource{
								SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
									Name:      "some-secret",
									Namespace: "some-namespace",
									Key:       "host",
								},
							},
						},
					},
				}},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pipeline).Build()

		sut := Reconciler{
			Client: fakeClient,
			config: Config{DaemonSet: types.NamespacedName{Name: "fluent-bit"}},
			prober: &mocks.DaemonSetProber{},
		}

		err := sut.updateStatus(context.Background(), pipeline.Name)
		require.NoError(t, err)

		var updatedPipeline telemetryv1alpha1.LogPipeline
		_ = fakeClient.Get(context.Background(), types.NamespacedName{Name: pipelineName}, &updatedPipeline)
		require.Len(t, updatedPipeline.Status.Conditions, 1)
		require.Equal(t, updatedPipeline.Status.Conditions[0].Type, telemetryv1alpha1.LogPipelinePending)
		require.Equal(t, updatedPipeline.Status.Conditions[0].Reason, reasonReferencedSecretMissing)
	})

	t.Run("should add pending condition if referenced secret exists but fluent bit is not ready", func(t *testing.T) {
		pipelineName := "pipeline"
		pipeline := &telemetryv1alpha1.LogPipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name: pipelineName,
			},
			Spec: telemetryv1alpha1.LogPipelineSpec{
				Output: telemetryv1alpha1.Output{
					HTTP: &telemetryv1alpha1.HTTPOutput{
						Host: telemetryv1alpha1.ValueType{
							ValueFrom: &telemetryv1alpha1.ValueFromSource{
								SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
									Name:      "some-secret",
									Namespace: "some-namespace",
									Key:       "host",
								},
							},
						},
					},
				}},
		}
		secret := &corev1.Secret{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-secret",
				Namespace: "some-namespace",
			},
			Data: map[string][]byte{"host": nil},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pipeline, secret).Build()

		proberStub := &mocks.DaemonSetProber{}
		proberStub.On("IsReady", mock.Anything, mock.Anything).Return(false, nil)

		sut := Reconciler{
			Client: fakeClient,
			config: Config{DaemonSet: types.NamespacedName{Name: "fluent-bit"}},
			prober: proberStub,
		}

		err := sut.updateStatus(context.Background(), pipeline.Name)
		require.NoError(t, err)

		var updatedPipeline telemetryv1alpha1.LogPipeline
		_ = fakeClient.Get(context.Background(), types.NamespacedName{Name: pipelineName}, &updatedPipeline)
		require.Len(t, updatedPipeline.Status.Conditions, 1)
		require.Equal(t, updatedPipeline.Status.Conditions[0].Type, telemetryv1alpha1.LogPipelinePending)
		require.Equal(t, updatedPipeline.Status.Conditions[0].Reason, reasonFluentBitDSNotReady)
	})

	t.Run("should add running condition if referenced secret exists and fluent bit is ready", func(t *testing.T) {
		pipelineName := "pipeline"
		pipeline := &telemetryv1alpha1.LogPipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name: pipelineName,
			},
			Spec: telemetryv1alpha1.LogPipelineSpec{
				Output: telemetryv1alpha1.Output{
					HTTP: &telemetryv1alpha1.HTTPOutput{
						Host: telemetryv1alpha1.ValueType{
							ValueFrom: &telemetryv1alpha1.ValueFromSource{
								SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
									Name:      "some-secret",
									Namespace: "some-namespace",
									Key:       "host",
								},
							},
						},
					},
				}},
		}
		secret := &corev1.Secret{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-secret",
				Namespace: "some-namespace",
			},
			Data: map[string][]byte{"host": nil},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pipeline, secret).Build()

		proberStub := &mocks.DaemonSetProber{}
		proberStub.On("IsReady", mock.Anything, mock.Anything).Return(true, nil)

		sut := Reconciler{
			Client: fakeClient,
			config: Config{DaemonSet: types.NamespacedName{Name: "fluent-bit"}},
			prober: proberStub,
		}

		err := sut.updateStatus(context.Background(), pipeline.Name)
		require.NoError(t, err)

		var updatedPipeline telemetryv1alpha1.LogPipeline
		_ = fakeClient.Get(context.Background(), types.NamespacedName{Name: pipelineName}, &updatedPipeline)
		require.Len(t, updatedPipeline.Status.Conditions, 1)
		require.Equal(t, updatedPipeline.Status.Conditions[0].Type, telemetryv1alpha1.LogPipelineRunning)
		require.Equal(t, updatedPipeline.Status.Conditions[0].Reason, reasonFluentBitDSReady)
	})

	t.Run("should reset conditions and add pending if fluent bit becomes not ready again", func(t *testing.T) {
		pipelineName := "pipeline"
		pipeline := &telemetryv1alpha1.LogPipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name: pipelineName,
			},
			Status: telemetryv1alpha1.LogPipelineStatus{
				Conditions: []telemetryv1alpha1.LogPipelineCondition{
					{Reason: reasonFluentBitDSNotReady, Type: telemetryv1alpha1.LogPipelinePending},
					{Reason: reasonFluentBitDSReady, Type: telemetryv1alpha1.LogPipelineRunning},
				},
			},
			Spec: telemetryv1alpha1.LogPipelineSpec{
				Output: telemetryv1alpha1.Output{
					HTTP: &telemetryv1alpha1.HTTPOutput{
						Host: telemetryv1alpha1.ValueType{
							ValueFrom: &telemetryv1alpha1.ValueFromSource{
								SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
									Name:      "some-secret",
									Namespace: "some-namespace",
									Key:       "host",
								},
							},
						},
					},
				}},
		}
		secret := &corev1.Secret{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-secret",
				Namespace: "some-namespace",
			},
			Data: map[string][]byte{"host": nil},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pipeline, secret).Build()

		proberStub := &mocks.DaemonSetProber{}
		proberStub.On("IsReady", mock.Anything, mock.Anything).Return(false, nil)

		sut := Reconciler{
			Client: fakeClient,
			config: Config{DaemonSet: types.NamespacedName{Name: "fluent-bit"}},
			prober: proberStub,
		}

		err := sut.updateStatus(context.Background(), pipeline.Name)
		require.NoError(t, err)

		var updatedPipeline telemetryv1alpha1.LogPipeline
		_ = fakeClient.Get(context.Background(), types.NamespacedName{Name: pipelineName}, &updatedPipeline)
		require.Len(t, updatedPipeline.Status.Conditions, 1)
		require.Equal(t, updatedPipeline.Status.Conditions[0].Type, telemetryv1alpha1.LogPipelinePending)
		require.Equal(t, updatedPipeline.Status.Conditions[0].Reason, reasonFluentBitDSNotReady)
	})

	t.Run("should reset conditions and add pending if some referenced secret does not exist anymore", func(t *testing.T) {
		pipelineName := "pipeline"
		pipeline := &telemetryv1alpha1.LogPipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name: pipelineName,
			},
			Status: telemetryv1alpha1.LogPipelineStatus{
				Conditions: []telemetryv1alpha1.LogPipelineCondition{
					{Reason: reasonFluentBitDSNotReady, Type: telemetryv1alpha1.LogPipelinePending},
					{Reason: reasonFluentBitDSReady, Type: telemetryv1alpha1.LogPipelineRunning},
				},
			},
			Spec: telemetryv1alpha1.LogPipelineSpec{
				Output: telemetryv1alpha1.Output{
					HTTP: &telemetryv1alpha1.HTTPOutput{
						Host: telemetryv1alpha1.ValueType{
							ValueFrom: &telemetryv1alpha1.ValueFromSource{
								SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
									Name:      "some-secret",
									Namespace: "some-namespace",
									Key:       "host",
								},
							},
						},
					},
				}},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pipeline).Build()

		sut := Reconciler{
			Client: fakeClient,
			config: Config{DaemonSet: types.NamespacedName{Name: "fluent-bit"}},
			prober: &mocks.DaemonSetProber{},
		}

		err := sut.updateStatus(context.Background(), pipeline.Name)
		require.NoError(t, err)

		var updatedPipeline telemetryv1alpha1.LogPipeline
		_ = fakeClient.Get(context.Background(), types.NamespacedName{Name: pipelineName}, &updatedPipeline)
		require.Len(t, updatedPipeline.Status.Conditions, 1)
		require.Equal(t, updatedPipeline.Status.Conditions[0].Type, telemetryv1alpha1.LogPipelinePending)
		require.Equal(t, updatedPipeline.Status.Conditions[0].Reason, reasonReferencedSecretMissing)
	})

	t.Run("should set status UnsupportedMode true if contains custom plugin", func(t *testing.T) {
		pipelineName := "pipeline"
		pipeline := &telemetryv1alpha1.LogPipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name: pipelineName,
			},
			Status: telemetryv1alpha1.LogPipelineStatus{
				Conditions: []telemetryv1alpha1.LogPipelineCondition{
					{Reason: reasonFluentBitDSReady, Type: telemetryv1alpha1.LogPipelineRunning},
				},
				UnsupportedMode: false,
			},
			Spec: telemetryv1alpha1.LogPipelineSpec{
				Filters: []telemetryv1alpha1.Filter{{Custom: "some-filter"}},
			},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pipeline).Build()
		proberStub := &mocks.DaemonSetProber{}
		proberStub.On("IsReady", mock.Anything, mock.Anything).Return(true, nil)
		sut := Reconciler{
			Client: fakeClient,
			config: Config{DaemonSet: types.NamespacedName{Name: "fluent-bit"}},
			prober: proberStub,
		}

		err := sut.updateStatus(context.Background(), pipeline.Name)
		require.NoError(t, err)

		var updatedPipeline telemetryv1alpha1.LogPipeline
		_ = fakeClient.Get(context.Background(), types.NamespacedName{Name: pipelineName}, &updatedPipeline)

		require.True(t, updatedPipeline.Status.UnsupportedMode)
	})

	t.Run("should set status UnsupportedMode false if does not contains custom plugin", func(t *testing.T) {
		pipelineName := "pipeline"
		pipeline := &telemetryv1alpha1.LogPipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name: pipelineName,
			},
			Status: telemetryv1alpha1.LogPipelineStatus{
				Conditions: []telemetryv1alpha1.LogPipelineCondition{
					{Reason: reasonFluentBitDSReady, Type: telemetryv1alpha1.LogPipelineRunning},
				},
				UnsupportedMode: true,
			},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pipeline).Build()
		proberStub := &mocks.DaemonSetProber{}
		proberStub.On("IsReady", mock.Anything, mock.Anything).Return(true, nil)
		sut := Reconciler{
			Client: fakeClient,
			config: Config{DaemonSet: types.NamespacedName{Name: "fluent-bit"}},
			prober: proberStub,
		}

		err := sut.updateStatus(context.Background(), pipeline.Name)
		require.NoError(t, err)

		var updatedPipeline telemetryv1alpha1.LogPipeline
		_ = fakeClient.Get(context.Background(), types.NamespacedName{Name: pipelineName}, &updatedPipeline)

		require.False(t, updatedPipeline.Status.UnsupportedMode)
	})
}
