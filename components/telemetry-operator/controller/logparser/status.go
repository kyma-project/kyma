package logparser

import (
	"context"
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *Reconciler) updateStatus(ctx context.Context, parserName string) error {
	log := logf.FromContext(ctx)
	var parser telemetryv1alpha1.LogParser
	if err := r.Get(ctx, types.NamespacedName{Name: parserName}, &parser); err != nil {
		return fmt.Errorf("failed to get LogParser: %v", err)
	}

	if parser.DeletionTimestamp != nil {
		return nil
	}

	fluentBitReady, err := r.prober.IsReady(ctx, r.config.DaemonSet)
	if err != nil {
		return err
	}

	if fluentBitReady {
		if parser.Status.HasCondition(telemetryv1alpha1.LogParserRunning) {
			return nil
		}

		running := telemetryv1alpha1.NewLogParserCondition(
			telemetryv1alpha1.FluentBitDSReadyReason,
			telemetryv1alpha1.LogParserRunning,
		)

		return setCondition(ctx, r.Client, &parser, running)
	}

	pending := telemetryv1alpha1.NewLogParserCondition(
		telemetryv1alpha1.FluentBitDSNotReadyReason,
		telemetryv1alpha1.LogParserPending,
	)

	if parser.Status.HasCondition(telemetryv1alpha1.LogParserRunning) {
		log.V(1).Info(fmt.Sprintf("Updating the status of %s to %s. Resetting previous conditions", parser.Name, pending.Type))
		parser.Status.Conditions = []telemetryv1alpha1.LogParserCondition{}
	}

	return setCondition(ctx, r.Client, &parser, pending)
}

func setCondition(ctx context.Context, client client.Client, parser *telemetryv1alpha1.LogParser, condition *telemetryv1alpha1.LogParserCondition) error {
	log := logf.FromContext(ctx)

	log.V(1).Info(fmt.Sprintf("Updating the status of %s to %s", parser.Name, condition.Type))

	parser.Status.SetCondition(*condition)

	if err := client.Status().Update(ctx, parser); err != nil {
		return fmt.Errorf("failed to update LogParser status to %s: %v", condition.Type, err)
	}
	return nil
}
