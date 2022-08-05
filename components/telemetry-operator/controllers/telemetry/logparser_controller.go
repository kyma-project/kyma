/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package telemetry

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/controller/logparser/sync"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// LogParserReconciler reconciles a LogParser object
type LogParserReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	FluentBitDaemonSet types.NamespacedName
	Parser             *sync.Syncer
	DaemonSetUtils     *fluentbit.DaemonSetUtils
}

func NewLogParserReconciler(client client.Client, scheme *runtime.Scheme, daemonSetConfig sync.Config, restartsTotal prometheus.Counter) *LogParserReconciler {
	var lpr LogParserReconciler
	lpr.Client = client
	lpr.Scheme = scheme
	lpr.FluentBitDaemonSet = daemonSetConfig.FluentBitDaemonSetName
	lpr.DaemonSetUtils = fluentbit.NewDaemonSetUtils(client, daemonSetConfig.FluentBitDaemonSetName, restartsTotal)
	lpr.Parser = sync.NewSyncer(client, daemonSetConfig)

	return &lpr
}

//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logparsers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logparsers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logparsers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *LogParserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	var logParser telemetryv1alpha1.LogParser
	if err := r.Get(ctx, req.NamespacedName, &logParser); err != nil {
		log.Info("Ignoring deleted LogParser")
		// Ignore not-found errors since we can get them on deleted requests
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var changed, err = r.Parser.SyncParsersConfigMap(ctx, &logParser)
	if err != nil {
		return ctrl.Result{Requeue: shouldRetryOn(err)}, nil
	}

	if changed {
		log.V(1).Info("Fluent Bit configuration was updated. Restarting the DaemonSet due to log parser change")

		if err = r.Update(ctx, &logParser); err != nil {
			log.Error(err, "Failed updating log parser")
			return ctrl.Result{Requeue: shouldRetryOn(err)}, err
		}

		if err = r.DaemonSetUtils.RestartFluentBit(ctx); err != nil {
			log.Error(err, "Failed restarting fluent bit daemon set")
			return ctrl.Result{Requeue: shouldRetryOn(err)}, err
		}

		condition := telemetryv1alpha1.NewLogParserCondition(
			telemetryv1alpha1.FluentBitDSRestartedReason,
			telemetryv1alpha1.LogParserPending,
		)
		if err = r.updateLogPipelineStatus(ctx, req.NamespacedName, condition); err != nil {
			return ctrl.Result{Requeue: shouldRetryOn(err)}, err
		}
	}

	if logParser.Status.GetCondition(telemetryv1alpha1.LogParserRunning) == nil {
		var ready bool
		ready, err = r.DaemonSetUtils.IsFluentBitDaemonSetReady(ctx)
		if err != nil {
			log.Error(err, "Failed to check fluent bit readiness")
			return ctrl.Result{RequeueAfter: requeueTime}, err
		}
		if !ready {
			log.V(1).Info(fmt.Sprintf("Checked %s - not yet ready. Requeueing...", req.NamespacedName.Name))
			return ctrl.Result{RequeueAfter: requeueTime}, nil
		}
		log.V(1).Info(fmt.Sprintf("Checked %s - ready", req.NamespacedName.Name))

		condition := telemetryv1alpha1.NewLogParserCondition(
			telemetryv1alpha1.FluentBitDSRestartCompletedReason,
			telemetryv1alpha1.LogParserRunning,
		)

		if err = r.updateLogPipelineStatus(ctx, req.NamespacedName, condition); err != nil {
			return ctrl.Result{RequeueAfter: requeueTime}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LogParserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1alpha1.LogParser{}).
		Complete(r)
}

func (r *LogParserReconciler) updateLogPipelineStatus(ctx context.Context, name types.NamespacedName, condition *telemetryv1alpha1.LogParserCondition) error {
	log := logf.FromContext(ctx)

	var logParser telemetryv1alpha1.LogParser
	if err := r.Get(ctx, name, &logParser); err != nil {
		log.Error(err, "Failed getting log parser")
		return err
	}

	// Do not update status if the log pipeline is being deleted
	if logParser.DeletionTimestamp != nil {
		return nil
	}

	// If the log parser had a running condition and then was modified, all conditions are removed.
	// In this case, condition tracking starts off from the beginning.
	if logParser.Status.GetCondition(telemetryv1alpha1.LogParserRunning) != nil &&
		condition.Type == telemetryv1alpha1.LogParserPending {
		log.V(1).Info(fmt.Sprintf("Updating the status of %s to %s. Resetting previous conditions", name.Name, condition.Type))
		logParser.Status.Conditions = []telemetryv1alpha1.LogParserCondition{}
	} else {
		log.V(1).Info(fmt.Sprintf("Updating the status of %s to %s", name.Name, condition.Type))
	}

	logParser.Status.SetCondition(*condition)

	if err := r.Status().Update(ctx, &logParser); err != nil {
		log.Error(err, fmt.Sprintf("Failed updating log pipeline status to %s", condition.Type))
		return err
	}
	return nil
}
