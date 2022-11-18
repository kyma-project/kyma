package logparser

import (
	"context"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const finalizer = "FLUENT_BIT_PARSERS_CONFIG_MAP"

func ensureFinalizer(ctx context.Context, client client.Client, parser *telemetryv1alpha1.LogParser) error {
	if parser.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(parser, finalizer) {
			return nil
		}

		controllerutil.AddFinalizer(parser, finalizer)
		return client.Update(ctx, parser)
	}

	return nil
}

func cleanupFinalizerIfNeeded(ctx context.Context, client client.Client, parser *telemetryv1alpha1.LogParser) error {
	if parser.DeletionTimestamp.IsZero() {
		return nil
	}

	if controllerutil.ContainsFinalizer(parser, finalizer) {
		controllerutil.RemoveFinalizer(parser, finalizer)
		return client.Update(ctx, parser)
	}

	return nil
}
