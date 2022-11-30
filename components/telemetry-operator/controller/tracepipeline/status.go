package tracepipeline

import (
	"context"
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *Reconciler) updateStatus(ctx context.Context, pipelineName string) error {
	return r.updateStatusConditions(ctx, pipelineName)
}

func (r *Reconciler) updateStatusConditions(ctx context.Context, pipelineName string) error {
	log := logf.FromContext(ctx)

	var pipeline telemetryv1alpha1.TracePipeline
	if err := r.Get(ctx, types.NamespacedName{Name: pipelineName}, &pipeline); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("failed to get TracePipeline: %v", err)
	}

	if pipeline.DeletionTimestamp != nil {
		return nil
	}

	openTelemetryReady, err := r.prober.IsReady(ctx, types.NamespacedName{Name: r.config.ResourceName, Namespace: r.config.CollectorNamespace})
	if err != nil {
		return err
	}

	if openTelemetryReady {
		if pipeline.Status.HasCondition(telemetryv1alpha1.TracePipelineRunning) {
			return nil
		}

		running := telemetryv1alpha1.NewTracePipelineCondition(
			telemetryv1alpha1.OpenTelemetryDReadyReason,
			telemetryv1alpha1.TracePipelineRunning,
		)

		return setCondition(ctx, r.Client, &pipeline, running)
	}

	pending := telemetryv1alpha1.NewTracePipelineCondition(
		telemetryv1alpha1.OpenTelemetryDNotReadyReason,
		telemetryv1alpha1.TracePipelinePending,
	)

	if pipeline.Status.HasCondition(telemetryv1alpha1.TracePipelineRunning) {
		log.V(1).Info(fmt.Sprintf("Updating the status of %s to %s. Resetting previous conditions", pipeline.Name, pending.Type))
		pipeline.Status.Conditions = []telemetryv1alpha1.TracePipelineCondition{}
	}

	return setCondition(ctx, r.Client, &pipeline, pending)
}

func setCondition(ctx context.Context, client client.Client, pipeline *telemetryv1alpha1.TracePipeline, condition *telemetryv1alpha1.TracePipelineCondition) error {
	log := logf.FromContext(ctx)

	log.V(1).Info(fmt.Sprintf("Updating the status of %s to %s", pipeline.Name, condition.Type))

	pipeline.Status.SetCondition(*condition)

	if err := client.Status().Update(ctx, pipeline); err != nil {
		return fmt.Errorf("failed to update LogPipeline status to %s: %v", condition.Type, err)
	}
	return nil
}
