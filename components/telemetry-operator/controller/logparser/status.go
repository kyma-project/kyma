package logparser

import (
	"context"
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *Reconciler) updateStatus(ctx context.Context, parser *telemetryv1alpha1.LogParser) error {
	fluentBitDSReady, err := r.daemonSet.IsReady(ctx, r.config.DaemonSet)
	if err != nil {
		return err
	}

	if !fluentBitDSReady {
		if err = setStatus(ctx, r.Client, parser.Name, telemetryv1alpha1.NewLogParserCondition(
			telemetryv1alpha1.FluentBitDSNotReadyReason,
			telemetryv1alpha1.LogParserPending,
		)); err != nil {
			return err
		}

		return nil
	}

	if fluentBitDSReady && parser.Status.GetCondition(telemetryv1alpha1.LogParserRunning) == nil {
		if err = setStatus(ctx, r.Client, parser.Name, telemetryv1alpha1.NewLogParserCondition(
			telemetryv1alpha1.FluentBitDSReadyReason,
			telemetryv1alpha1.LogParserRunning,
		)); err != nil {
			return err
		}
	}

	return nil
}

func setStatus(ctx context.Context, client client.Client, parserName string, condition *telemetryv1alpha1.LogParserCondition) error {
	log := logf.FromContext(ctx)

	var parser telemetryv1alpha1.LogParser
	if err := client.Get(ctx, types.NamespacedName{Name: parserName}, &parser); err != nil {
		return fmt.Errorf("failed to get LogParser: %v", err)
	}

	// Do not update status if the log parser is being deleted
	if parser.DeletionTimestamp != nil {
		return nil
	}

	// If the log parser had a running condition and then was modified, all conditions are removed.
	// In this case, condition tracking starts off from the beginning.
	if parser.Status.GetCondition(telemetryv1alpha1.LogParserRunning) != nil &&
		condition.Type == telemetryv1alpha1.LogParserPending {
		log.V(1).Info(fmt.Sprintf("Updating the status of %s to %s. Resetting previous conditions", parserName, condition.Type))
		parser.Status.Conditions = []telemetryv1alpha1.LogParserCondition{}
	} else {
		log.V(1).Info(fmt.Sprintf("Updating the status of %s to %s", parserName, condition.Type))
	}

	parser.Status.SetCondition(*condition)

	if err := client.Status().Update(ctx, &parser); err != nil {
		log.Error(err, fmt.Sprintf("Failed to update LogParser status to %s", condition.Type))
		return err
	}
	return nil
}
