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

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/sync"

	"github.com/prometheus/client_golang/prometheus"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
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
	Scheme                 *runtime.Scheme
	Syncer                 *sync.LogPipelineSyncer
	FluentBitDaemonSet     types.NamespacedName
	FluentBitRestartsCount prometheus.Counter
}

// NewLogPipelineReconciler returns a new LogPipelineReconciler using the given FluentBit config arguments
func NewLogPipelineReconciler(client client.Client, scheme *runtime.Scheme, daemonSetConfig sync.FluentBitDaemonSetConfig, emitterConfig fluentbit.EmitterConfig) *LogPipelineReconciler {
	var lpr LogPipelineReconciler
	lpr.Client = client
	lpr.Scheme = scheme
	lpr.Syncer = sync.NewLogPipelineSyncer(client,
		daemonSetConfig,
		emitterConfig,
	)
	lpr.FluentBitDaemonSet = daemonSetConfig.FluentBitDaemonSetName

	lpr.FluentBitRestartsCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "telemetry_operator_fluentbit_restarts_total",
		Help: "Number of triggered FluentBit restarts",
	})
	metrics.Registry.MustRegister(lpr.FluentBitRestartsCount)

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
		// Ignore not-found errors since we can get them on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var changed, err = r.Syncer.SyncAll(ctx, &logPipeline)
	if err != nil {
		return ctrl.Result{Requeue: shouldRetryOn(err)}, nil
	}

	if changed {
		log.V(1).Info("Fluent bit configuration was updated. Restarting the daemon set")

		if err = r.Update(ctx, &logPipeline); err != nil {
			log.Error(err, "Failed updating log pipeline")
			return ctrl.Result{Requeue: shouldRetryOn(err)}, err
		}

		if err = r.restartFluentBit(ctx); err != nil {
			log.Error(err, "Failed restarting fluent bit daemon set")
			return ctrl.Result{Requeue: shouldRetryOn(err)}, err
		}

		condition := telemetryv1alpha1.NewLogPipelineCondition(
			telemetryv1alpha1.FluentBitDSRestartedReason,
			telemetryv1alpha1.LogPipelinePending,
		)
		if err = r.updateLogPipelineStatus(ctx, req.NamespacedName, condition); err != nil {
			return ctrl.Result{RequeueAfter: requeueTime}, err
		}

		return ctrl.Result{RequeueAfter: requeueTime}, nil
	}

	if logPipeline.Status.GetCondition(telemetryv1alpha1.LogPipelineRunning) == nil {
		var ready bool
		ready, err = r.isFluentBitDaemonSetReady(ctx)
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
		if err = r.updateLogPipelineStatus(ctx, req.NamespacedName, condition); err != nil {
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

// Delete all Fluent Bit pods to apply new configuration.
func (r *LogPipelineReconciler) restartFluentBit(ctx context.Context) error {
	log := logf.FromContext(ctx)
	var ds appsv1.DaemonSet
	if err := r.Get(ctx, r.FluentBitDaemonSet, &ds); err != nil {
		log.Error(err, "Failed getting fluent bit DaemonSet")
		return err
	}

	patchedDS := *ds.DeepCopy()
	if patchedDS.Spec.Template.ObjectMeta.Annotations == nil {
		patchedDS.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	patchedDS.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	if err := r.Patch(ctx, &patchedDS, client.MergeFrom(&ds)); err != nil {
		log.Error(err, "Failed to patch fluent bit to trigger rolling update")
		return err
	}
	r.FluentBitRestartsCount.Inc()
	return nil
}

func (r *LogPipelineReconciler) isFluentBitDaemonSetReady(ctx context.Context) (bool, error) {
	log := logf.FromContext(ctx)
	var ds appsv1.DaemonSet
	if err := r.Get(ctx, r.FluentBitDaemonSet, &ds); err != nil {
		log.Error(err, "Failed getting fluent bit daemon set")
		return false, err
	}

	generation := ds.Generation
	observedGeneration := ds.Status.ObservedGeneration
	updated := ds.Status.UpdatedNumberScheduled
	desired := ds.Status.DesiredNumberScheduled
	ready := ds.Status.NumberReady

	log.V(1).Info(fmt.Sprintf("Checking fluent bit: updated: %d, desired: %d, ready: %d, generation: %d, observed generation: %d",
		updated, desired, ready, generation, observedGeneration))

	return observedGeneration == generation && updated == desired && ready >= desired, nil
}

func (r *LogPipelineReconciler) updateLogPipelineStatus(ctx context.Context,
	name types.NamespacedName,
	condition *telemetryv1alpha1.LogPipelineCondition) error {
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

	if err := r.Status().Update(ctx, &logPipeline); err != nil {
		log.Error(err, fmt.Sprintf("Failed updating log pipeline status to %s", condition.Type))
		return err
	}
	return nil
}
