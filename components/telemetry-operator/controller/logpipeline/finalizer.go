package logpipeline

import (
	"context"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	sectionsFinalizer = "FLUENT_BIT_SECTIONS_CONFIG_MAP"
	filesFinalizer    = "FLUENT_BIT_FILES"
)

func (r *Reconciler) manageFinalizers(ctx context.Context, pipeline *telemetryv1alpha1.LogPipeline) error {
	var changed bool
	if pipeline.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(pipeline, sectionsFinalizer) {
			controllerutil.AddFinalizer(pipeline, sectionsFinalizer)
			changed = true
		}

		if len(pipeline.Spec.Files) > 0 && !controllerutil.ContainsFinalizer(pipeline, filesFinalizer) {
			controllerutil.AddFinalizer(pipeline, filesFinalizer)
			changed = true
		}
	} else {
		if controllerutil.ContainsFinalizer(pipeline, sectionsFinalizer) {
			controllerutil.RemoveFinalizer(pipeline, sectionsFinalizer)
			changed = true
		}

		if controllerutil.ContainsFinalizer(pipeline, filesFinalizer) {
			controllerutil.RemoveFinalizer(pipeline, filesFinalizer)
			changed = true
		}
	}

	if !changed {
		return nil
	}

	return r.Update(ctx, pipeline)
}
