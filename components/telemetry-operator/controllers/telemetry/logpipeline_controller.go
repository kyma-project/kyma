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
	"time"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/sync"

	"github.com/prometheus/client_golang/prometheus"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	requeueTime = 10 * time.Second
)

// LogPipelineReconciler reconciles a LogPipeline object
type LogPipelineReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	Syncer             *sync.LogPipelineSyncer
	FluentBitDaemonSet types.NamespacedName
	restartsTotal      prometheus.Counter
	unsupportedTotal   prometheus.Gauge
	FluentBitUtils     *fluentbit.DaemonSetUtils
}

// NewLogPipelineReconciler returns a new LogPipelineReconciler using the given FluentBit config arguments
func NewLogPipelineReconciler(client client.Client, scheme *runtime.Scheme, daemonSetConfig sync.FluentBitDaemonSetConfig, pipelineConfig fluentbit.PipelineConfig, restartsTotal prometheus.Counter) *LogPipelineReconciler {
	var lpr LogPipelineReconciler
	lpr.Client = client
	lpr.Scheme = scheme
	lpr.Syncer = sync.NewLogPipelineSyncer(client,
		daemonSetConfig,
		pipelineConfig,
	)
	lpr.FluentBitDaemonSet = daemonSetConfig.FluentBitDaemonSetName

	// Metrics
	lpr.restartsTotal = restartsTotal
	lpr.unsupportedTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "telemetry_plugins_unsupported_total",
		Help: "Number of custom filters or outputs to indicate unsupported mode.",
	})
	lpr.FluentBitUtils = fluentbit.NewDaemonSetUtils(client, daemonSetConfig.FluentBitDaemonSetName)

	metrics.Registry.MustRegister(lpr.unsupportedTotal)

	return &lpr
}

// SetupWithManager sets up the controller with the Manager.
func (r *LogPipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1alpha1.LogPipeline{}).
		Complete(r)
}

//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logpipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logpipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logpipelines/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=daemonsets,verbs=get;list;watch;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *LogPipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var logPipeline telemetryv1alpha1.LogPipeline
	if err := r.Get(ctx, req.NamespacedName, &logPipeline); err != nil {
		log.Info("Ignoring deleted LogPipeline")
		// Ignore not-found errors since we can get them on deleted requests
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	secretsOK := r.Syncer.SecretValidator.ValidateSecretsExist(ctx, &logPipeline)
	if !secretsOK {
		condition := telemetryv1alpha1.NewLogPipelineCondition(
			telemetryv1alpha1.SecretsNotPresent,
			telemetryv1alpha1.LogPipelinePending,
		)
		pipeLineUnsupported := sync.LogPipelineIsUnsupported(logPipeline)
		if err := r.updateLogPipelineStatus(ctx, req.NamespacedName, condition, pipeLineUnsupported); err != nil {
			return ctrl.Result{Requeue: shouldRetryOn(err)}, nil
		}

		return ctrl.Result{RequeueAfter: requeueTime}, nil
	}

	var changed, err = r.Syncer.SyncAll(ctx, &logPipeline)
	if err != nil {
		return ctrl.Result{Requeue: shouldRetryOn(err)}, nil
	}

	r.unsupportedTotal.Set(float64(r.Syncer.UnsupportedPluginsTotal))

	if changed {
		log.V(1).Info("Fluent Bit configuration was updated. Restarting the DaemonSet due to logpipeline change")

		if err = r.Update(ctx, &logPipeline); err != nil {
			log.Error(err, "Failed updating log pipeline")
			return ctrl.Result{Requeue: shouldRetryOn(err)}, err
		}

		if err = r.FluentBitUtils.RestartFluentBit(ctx, r.restartsTotal); err != nil {
			log.Error(err, "Failed restarting fluent bit daemon set")
			return ctrl.Result{Requeue: shouldRetryOn(err)}, err
		}

		condition := telemetryv1alpha1.NewLogPipelineCondition(
			telemetryv1alpha1.FluentBitDSRestartedReason,
			telemetryv1alpha1.LogPipelinePending,
		)
		pipeLineUnsupported := sync.LogPipelineIsUnsupported(logPipeline)
		if err = r.updateLogPipelineStatus(ctx, req.NamespacedName, condition, pipeLineUnsupported); err != nil {
			return ctrl.Result{RequeueAfter: requeueTime}, err
		}

		return ctrl.Result{RequeueAfter: requeueTime}, nil
	}

	if logPipeline.Status.GetCondition(telemetryv1alpha1.LogPipelineRunning) == nil {
		var ready bool
		ready, err = r.FluentBitUtils.IsFluentBitDaemonSetReady(ctx)
		if err != nil {
			log.Error(err, "Failed to check fluent bit readiness")
			return ctrl.Result{RequeueAfter: requeueTime}, err
		}
		if !ready {
			log.V(1).Info(fmt.Sprintf("Checked %s - not yet ready. Requeueing...", req.NamespacedName.Name))
			return ctrl.Result{RequeueAfter: requeueTime}, nil
		}
		log.V(1).Info(fmt.Sprintf("Checked %s - ready", req.NamespacedName.Name))

		condition := telemetryv1alpha1.NewLogPipelineCondition(
			telemetryv1alpha1.FluentBitDSRestartCompletedReason,
			telemetryv1alpha1.LogPipelineRunning,
		)
		pipeLineUnsupported := sync.LogPipelineIsUnsupported(logPipeline)

		if err = r.updateLogPipelineStatus(ctx, req.NamespacedName, condition, pipeLineUnsupported); err != nil {
			return ctrl.Result{RequeueAfter: requeueTime}, err
		}
	}

	return ctrl.Result{}, nil
}

// Indicate if an error from the kubernetes client should be retried. Errors caused by a bad request or configuration should not be retried.
func shouldRetryOn(err error) bool {
	return !errors.IsInvalid(err) &&
		!errors.IsNotAcceptable(err) &&
		!errors.IsUnsupportedMediaType(err) &&
		!errors.IsMethodNotSupported(err) &&
		!errors.IsBadRequest(err) &&
		!errors.IsUnauthorized(err) &&
		!errors.IsForbidden(err)
}

func (r *LogPipelineReconciler) updateLogPipelineStatus(ctx context.Context, name types.NamespacedName, condition *telemetryv1alpha1.LogPipelineCondition, unSupported bool) error {
	log := logf.FromContext(ctx)

	var logPipeline telemetryv1alpha1.LogPipeline
	if err := r.Get(ctx, name, &logPipeline); err != nil {
		log.Error(err, "Failed getting log pipeline")
		return err
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
		log.Error(err, fmt.Sprintf("Failed updating log pipeline status to %s", condition.Type))
		return err
	}
	return nil
}
