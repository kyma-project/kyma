package logpipeline

import (
	"context"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"testing"
)

func TestEnsureFinalizers(t *testing.T) {
	t.Run("without files", func(t *testing.T) {
		scheme := runtime.NewScheme()
		_ = telemetryv1alpha1.AddToScheme(scheme)
		pipeline := &telemetryv1alpha1.LogPipeline{ObjectMeta: metav1.ObjectMeta{Name: "pipeline"}}
		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pipeline).Build()
		sut := Reconciler{Client: client}

		err := sut.ensureFinalizers(context.Background(), pipeline)
		require.NoError(t, err)

		var updatedPipeline telemetryv1alpha1.LogPipeline
		_ = client.Get(ctx, types.NamespacedName{Name: pipeline.Name}, &updatedPipeline)

		require.True(t, controllerutil.ContainsFinalizer(&updatedPipeline, sectionsFinalizer))
		require.False(t, controllerutil.ContainsFinalizer(&updatedPipeline, filesFinalizer))
	})

	t.Run("with files", func(t *testing.T) {
		pipeline := &telemetryv1alpha1.LogPipeline{
			ObjectMeta: metav1.ObjectMeta{Name: "pipeline"},
			Spec: telemetryv1alpha1.LogPipelineSpec{
				Files: []telemetryv1alpha1.FileMount{
					{
						Name:    "script.js",
						Content: "",
					},
				},
			},
		}

		scheme := runtime.NewScheme()
		_ = telemetryv1alpha1.AddToScheme(scheme)
		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pipeline).Build()
		sut := Reconciler{Client: client}

		err := sut.ensureFinalizers(context.Background(), pipeline)
		require.NoError(t, err)

		var updatedPipeline telemetryv1alpha1.LogPipeline
		_ = client.Get(ctx, types.NamespacedName{Name: pipeline.Name}, &updatedPipeline)

		require.True(t, controllerutil.ContainsFinalizer(&updatedPipeline, sectionsFinalizer))
		require.True(t, controllerutil.ContainsFinalizer(&updatedPipeline, filesFinalizer))
	})
}

func TestCleanupFinalizers(t *testing.T) {
	t.Run("without files", func(t *testing.T) {
		ts := metav1.Now()
		pipeline := &telemetryv1alpha1.LogPipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "pipeline",
				Finalizers:        []string{sectionsFinalizer},
				DeletionTimestamp: &ts,
			},
		}

		scheme := runtime.NewScheme()
		_ = telemetryv1alpha1.AddToScheme(scheme)
		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pipeline).Build()
		sut := Reconciler{Client: client}

		err := sut.cleanupFinalizersIfNeeded(context.Background(), pipeline)
		require.NoError(t, err)

		var updatedPipeline telemetryv1alpha1.LogPipeline
		_ = client.Get(ctx, types.NamespacedName{Name: pipeline.Name}, &updatedPipeline)

		require.False(t, controllerutil.ContainsFinalizer(&updatedPipeline, sectionsFinalizer))
	})

	t.Run("with files", func(t *testing.T) {
		ts := metav1.Now()
		pipeline := &telemetryv1alpha1.LogPipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "pipeline",
				Finalizers:        []string{sectionsFinalizer, filesFinalizer},
				DeletionTimestamp: &ts,
			},
		}

		scheme := runtime.NewScheme()
		_ = telemetryv1alpha1.AddToScheme(scheme)
		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pipeline).Build()
		sut := Reconciler{Client: client}

		err := sut.cleanupFinalizersIfNeeded(context.Background(), pipeline)
		require.NoError(t, err)

		var updatedPipeline telemetryv1alpha1.LogPipeline
		_ = client.Get(ctx, types.NamespacedName{Name: pipeline.Name}, &updatedPipeline)

		require.False(t, controllerutil.ContainsFinalizer(&updatedPipeline, sectionsFinalizer))
		require.False(t, controllerutil.ContainsFinalizer(&updatedPipeline, filesFinalizer))
	})
}
