package logpipeline

import (
	"context"
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *Reconciler) updateStatus(ctx context.Context, pipeline *telemetryv1alpha1.LogPipeline) error {
	//log := logf.FromContext(ctx)
	//
	//secretsOK := r.ensureReferencedSecretsExist(ctx, pipeline)
	//if !secretsOK {
	//	condition := telemetryv1alpha1.NewLogPipelineCondition(
	//		telemetryv1alpha1.SecretsNotPresent,
	//		telemetryv1alpha1.LogPipelinePending,
	//	)
	//	pipelineUnsupported := pipeline.ContainsCustomPlugin()
	//	if err := r.updateLogPipelineStatus(ctx, req.NamespacedName, condition, pipelineUnsupported); err != nil {
	//		return err
	//	}
	//
	//	return nil
	//}
	//
	//condition := telemetryv1alpha1.NewLogPipelineCondition(
	//	telemetryv1alpha1.FluentBitDSRestartedReason,
	//	telemetryv1alpha1.LogPipelinePending,
	//)
	//pipelineUnsupported := pipeline.ContainsCustomPlugin()
	//if err := r.updateLogPipelineStatus(ctx, req.NamespacedName, condition, pipelineUnsupported); err != nil {
	//	return err
	//}

	return nil

	//if pipeline.Status.GetCondition(telemetryv1alpha1.LogPipelineRunning) == nil {
	//	var ready bool
	//	ready, err := r.daemonSetHelper.IsReady(ctx, r.config.DaemonSet)
	//	if err != nil {
	//		return fmt.Errorf("failed to check Fluent Bit readiness: %v", err)
	//	}
	//	if !ready {
	//		log.V(1).Info(fmt.Sprintf("Checked %s - not yet ready. Requeueing...", req.NamespacedName.Name))
	//		return nil
	//	}
	//	log.V(1).Info(fmt.Sprintf("Checked %s - ready", req.NamespacedName.Name))
	//
	//	condition := telemetryv1alpha1.NewLogPipelineCondition(
	//		telemetryv1alpha1.FluentBitDSRestartCompletedReason,
	//		telemetryv1alpha1.LogPipelineRunning,
	//	)
	//	pipelineUnsupported := pipeline.ContainsCustomPlugin()
	//
	//	if err = r.updateLogPipelineStatus(ctx, req.NamespacedName, condition, pipelineUnsupported); err != nil {
	//		return err
	//	}
	//}
}

func (r *Reconciler) updateLogPipelineStatus(ctx context.Context, name types.NamespacedName, condition *telemetryv1alpha1.LogPipelineCondition, unSupported bool) error {
	log := logf.FromContext(ctx)

	var logPipeline telemetryv1alpha1.LogPipeline
	if err := r.Get(ctx, name, &logPipeline); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed get LogPipeline: %v", err)
	}

	// Do not update status if the log pipeline is being deleted
	if logPipeline.DeletionTimestamp != nil {
		return nil
	}

	// If the log pipeline had a running condition and then was modified, all conditions are removed.
	// In this case, condition tracking starts off from the beginning.
	if logPipeline.Status.GetCondition(telemetryv1alpha1.LogPipelineRunning) != nil &&
		condition.Type == telemetryv1alpha1.LogPipelinePending {
		log.V(1).Info(fmt.Sprintf("Updating the status of %s to %s. Resetting previous conditions", name.Name, condition.Type))
		logPipeline.Status.Conditions = []telemetryv1alpha1.LogPipelineCondition{}
	} else {
		log.V(1).Info(fmt.Sprintf("Updating the status of %s to %s", name.Name, condition.Type))
	}

	logPipeline.Status.SetCondition(*condition)
	logPipeline.Status.UnsupportedMode = unSupported

	if err := r.Status().Update(ctx, &logPipeline); err != nil {
		return fmt.Errorf("failed to update LogPipeline status to %s: %v", condition.Type, err)
	}
	return nil
}

func (r *Reconciler) ensureReferencedSecretsExist(ctx context.Context, pipeline *telemetryv1alpha1.LogPipeline) bool {
	secretRefFields := lookupSecretRefFields(pipeline)
	for _, field := range secretRefFields {
		hasKey := r.ensureSecretHasKey(ctx, field.secretKeyRef)
		if !hasKey {
			return false
		}
	}

	return true
}

func (r *Reconciler) ensureSecretHasKey(ctx context.Context, from telemetryv1alpha1.SecretKeyRef) bool {
	log := logf.FromContext(ctx)

	var secret corev1.Secret
	if err := r.Get(ctx, types.NamespacedName{Name: from.Name, Namespace: from.Namespace}, &secret); err != nil {
		log.V(1).Info(fmt.Sprintf("Unable to get secret '%s' from namespace '%s'", from.Name, from.Namespace))
		return false
	}
	if _, ok := secret.Data[from.Key]; !ok {
		log.V(1).Info(fmt.Sprintf("Unable to find key '%s' in secret '%s'", from.Key, from.Name))
		return false
	}

	return true
}
