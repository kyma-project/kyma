package logparser

import (
	"context"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"testing"
)

func TestEnsureFinalizer(t *testing.T) {
	t.Run("without DeletionTimestamp", func(t *testing.T) {
		ctx := context.Background()
		scheme := runtime.NewScheme()
		_ = telemetryv1alpha1.AddToScheme(scheme)
		parser := &telemetryv1alpha1.LogParser{ObjectMeta: v1.ObjectMeta{Name: "parser"}}
		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(parser).Build()

		err := ensureFinalizer(ctx, client, parser)
		require.NoError(t, err)

		var updatedParser telemetryv1alpha1.LogParser
		_ = client.Get(ctx, types.NamespacedName{Name: parser.Name}, &updatedParser)

		require.True(t, controllerutil.ContainsFinalizer(&updatedParser, finalizer))
	})

	t.Run("with DeletionTimestamp", func(t *testing.T) {
		ctx := context.Background()
		client := fake.NewClientBuilder().Build()
		timestamp := v1.Now()
		parser := &telemetryv1alpha1.LogParser{
			ObjectMeta: v1.ObjectMeta{
				DeletionTimestamp: &timestamp,
				Name:              "parser"}}

		err := ensureFinalizer(ctx, client, parser)
		require.NoError(t, err)
	})
}

func TestCleanupFinalizer(t *testing.T) {
	t.Run("without DeletionTimestamp", func(t *testing.T) {
		ctx := context.Background()
		parser := &telemetryv1alpha1.LogParser{
			ObjectMeta: v1.ObjectMeta{
				Name:       "parser",
				Finalizers: []string{finalizer},
			},
		}
		client := fake.NewClientBuilder().Build()

		err := cleanupFinalizer(ctx, client, parser)
		require.NoError(t, err)
	})

	t.Run("with DeletionTimestamp", func(t *testing.T) {
		ctx := context.Background()
		scheme := runtime.NewScheme()
		_ = telemetryv1alpha1.AddToScheme(scheme)
		timestamp := v1.Now()
		parser := &telemetryv1alpha1.LogParser{
			ObjectMeta: v1.ObjectMeta{
				Name:              "parser",
				Finalizers:        []string{finalizer},
				DeletionTimestamp: &timestamp,
			},
		}
		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(parser).Build()

		err := cleanupFinalizer(ctx, client, parser)
		require.NoError(t, err)

		var updatedParser telemetryv1alpha1.LogPipeline
		_ = client.Get(ctx, types.NamespacedName{Name: parser.Name}, &updatedParser)

		require.False(t, controllerutil.ContainsFinalizer(&updatedParser, finalizer))
	})
}
