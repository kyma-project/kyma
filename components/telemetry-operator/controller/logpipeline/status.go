package logpipeline

import (
	"context"
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *Reconciler) updateStatus(ctx context.Context, pipeline *telemetryv1alpha1.LogPipeline) error {
	secretsExist := checkReferencedSecretsExist(ctx, r.Client, pipeline)
	if !secretsExist {
		if err := setStatus(ctx, r.Client, pipeline.Name, telemetryv1alpha1.NewLogPipelineCondition(
			telemetryv1alpha1.ReferencedSecretMissingReason,
			telemetryv1alpha1.LogPipelinePending,
		)); err != nil {
			return err
		}

		return nil
	}

	fluentBitDSReady, err := r.prober.IsReady(ctx, r.config.DaemonSet)
	if err != nil {
		return err
	}

	if !fluentBitDSReady {
		if err = setStatus(ctx, r.Client, pipeline.Name, telemetryv1alpha1.NewLogPipelineCondition(
			telemetryv1alpha1.FluentBitDSNotReadyReason,
			telemetryv1alpha1.LogPipelinePending,
		)); err != nil {
			return err
		}

		return nil
	}

	if fluentBitDSReady && pipeline.Status.GetCondition(telemetryv1alpha1.LogPipelineRunning) == nil {
		if err = setStatus(ctx, r.Client, pipeline.Name, telemetryv1alpha1.NewLogPipelineCondition(
			telemetryv1alpha1.FluentBitDSReadyReason,
			telemetryv1alpha1.LogPipelineRunning,
		)); err != nil {
			return err
		}
	}

	return nil
}

func setStatus(ctx context.Context, client client.Client, pipelineName string, condition *telemetryv1alpha1.LogPipelineCondition) error {
	log := logf.FromContext(ctx)

	var pipeline telemetryv1alpha1.LogPipeline
	if err := client.Get(ctx, types.NamespacedName{Name: pipelineName}, &pipeline); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed get LogPipeline: %v", err)
	}

	// Do not update status if the log pipeline is being deleted
	if pipeline.DeletionTimestamp != nil {
		return nil
	}

	// If the log pipeline had a running condition and then was modified, all conditions are removed.
	// In this case, condition tracking starts off from the beginning.
	if pipeline.Status.GetCondition(telemetryv1alpha1.LogPipelineRunning) != nil &&
		condition.Type == telemetryv1alpha1.LogPipelinePending {
		log.V(1).Info(fmt.Sprintf("Updating the status of %s to %s. Resetting previous conditions", pipelineName, condition.Type))
		pipeline.Status.Conditions = []telemetryv1alpha1.LogPipelineCondition{}
	} else {
		log.V(1).Info(fmt.Sprintf("Updating the status of %s to %s", pipelineName, condition.Type))
	}

	pipeline.Status.SetCondition(*condition)
	pipeline.Status.UnsupportedMode = pipeline.ContainsCustomPlugin()

	if err := client.Status().Update(ctx, &pipeline); err != nil {
		return fmt.Errorf("failed to update LogPipeline status to %s: %v", condition.Type, err)
	}
	return nil
}

func checkReferencedSecretsExist(ctx context.Context, client client.Client, pipeline *telemetryv1alpha1.LogPipeline) bool {
	secretRefFields := lookupSecretRefFields(pipeline)
	for _, field := range secretRefFields {
		hasKey := checkSecretHasKey(ctx, client, field.secretKeyRef)
		if !hasKey {
			return false
		}
	}

	return true
}

func checkSecretHasKey(ctx context.Context, client client.Client, from telemetryv1alpha1.SecretKeyRef) bool {
	log := logf.FromContext(ctx)

	var secret corev1.Secret
	if err := client.Get(ctx, types.NamespacedName{Name: from.Name, Namespace: from.Namespace}, &secret); err != nil {
		log.V(1).Info(fmt.Sprintf("Unable to get secret '%s' from namespace '%s'", from.Name, from.Namespace))
		return false
	}
	if _, ok := secret.Data[from.Key]; !ok {
		log.V(1).Info(fmt.Sprintf("Unable to find key '%s' in secret '%s'", from.Key, from.Name))
		return false
	}

	return true
}
