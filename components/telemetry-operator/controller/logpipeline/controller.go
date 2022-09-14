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

package logpipeline

import (
	"context"
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/kyma-project/kyma/components/telemetry-operator/controller"
	controllermetrics "github.com/kyma-project/kyma/components/telemetry-operator/controller/metrics"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config/builder"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Config struct {
	DaemonSet         types.NamespacedName
	SectionsConfigMap types.NamespacedName
	FilesConfigMap    types.NamespacedName
	EnvSecret         types.NamespacedName
	PipelineDefaults  builder.PipelineDefaults
}

// Reconciler reconciles a LogPipeline object
type Reconciler struct {
	client.Client
	config                  Config
	syncer                  *syncer
	daemonSetHelper         *kubernetes.DaemonSetHelper
	allLogPipelines         prometheus.Gauge
	unsupportedLogPipelines prometheus.Gauge
	secrets                 secretsCache
}

// NewReconciler returns a new LogPipelineReconciler using the given Fluent Bit config arguments
func NewReconciler(
	client client.Client,
	config Config,
) *Reconciler {
	var r Reconciler
	r.Client = client
	r.config = config
	r.syncer = newSyncer(client, config)
	r.daemonSetHelper = kubernetes.NewDaemonSetHelper(client, controllermetrics.FluentBitTriggeredRestartsTotal)
	r.allLogPipelines = prometheus.NewGauge(prometheus.GaugeOpts{Name: "telemetry_all_logpipelines", Help: "Number of log pipelines."})
	r.unsupportedLogPipelines = prometheus.NewGauge(prometheus.GaugeOpts{Name: "telemetry_unsupported_logpipelines", Help: "Number of log pipelines with custom filters or outputs."})
	r.secrets = newSecretsCache()
	metrics.Registry.MustRegister(r.allLogPipelines, r.unsupportedLogPipelines)
	controllermetrics.RegisterMetrics()

	return &r
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1alpha1.LogPipeline{}).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.createReconReqs),
		).
		Complete(r)
}

func (r *Reconciler) createReconReqs(object client.Object) []reconcile.Request {
	secret := object.(*corev1.Secret)

	fmt.Printf("Secret changed event: Handling Secret with name: %s\n", secret.Name)

	secretName := types.NamespacedName{Namespace: secret.Namespace, Name: secret.Name}
	pipelines := r.secrets.get(secretName)
	var requests []reconcile.Request
	for _, p := range pipelines {
		request := reconcile.Request{NamespacedName: types.NamespacedName{Name: string(p)}}
		fmt.Printf("Secret changed event: Creating Reconciliation request for pipeline: %s\n", string(p))
		requests = append(requests, request)
	}

	fmt.Printf("Secret changed event: Created %d new Reconciliation requests.\n", len(requests))
	return requests
}

//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logpipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logpipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logpipelines/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=daemonsets,verbs=get;list;watch;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var allPipelines telemetryv1alpha1.LogPipelineList
	if err := r.List(ctx, &allPipelines); err != nil {
		log.Error(err, "Failed to get all log pipelines")
		return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, err
	}
	r.updateMetrics(&allPipelines)

	var logPipeline telemetryv1alpha1.LogPipeline
	if err := r.Get(ctx, req.NamespacedName, &logPipeline); err != nil {
		log.Info("Ignoring deleted LogPipeline")
		// Ignore not-found errors since we can get them on deleted requests
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	secretsOK := r.syncer.secretHelper.ValidatePipelineSecretsExist(ctx, &logPipeline)
	if !secretsOK {
		condition := telemetryv1alpha1.NewLogPipelineCondition(
			telemetryv1alpha1.SecretsNotPresent,
			telemetryv1alpha1.LogPipelinePending,
		)
		pipelineUnsupported := logPipeline.ContainsCustomPlugin()
		if err := r.updateLogPipelineStatus(ctx, req.NamespacedName, condition, pipelineUnsupported); err != nil {
			return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
		}

		return ctrl.Result{RequeueAfter: controller.RequeueTime}, nil
	}

	changed, err := r.syncer.syncAll(ctx, &logPipeline, &allPipelines)
	if err != nil {
		return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
	}

	if changed {
		log.V(1).Info("Fluent Bit configuration was updated. Restarting the DaemonSet")

		if err = r.Update(ctx, &logPipeline); err != nil {
			log.Error(err, "Failed to update log pipeline")
			return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, err
		}

		if err = r.daemonSetHelper.Restart(ctx, r.config.DaemonSet); err != nil {
			log.Error(err, "Failed to restart Fluent Bit DaemonSet")
			return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, err
		}

		condition := telemetryv1alpha1.NewLogPipelineCondition(
			telemetryv1alpha1.FluentBitDSRestartedReason,
			telemetryv1alpha1.LogPipelinePending,
		)
		pipelineUnsupported := logPipeline.ContainsCustomPlugin()
		if err = r.updateLogPipelineStatus(ctx, req.NamespacedName, condition, pipelineUnsupported); err != nil {
			return ctrl.Result{RequeueAfter: controller.RequeueTime}, err
		}

		return ctrl.Result{RequeueAfter: controller.RequeueTime}, nil
	}

	if logPipeline.Status.GetCondition(telemetryv1alpha1.LogPipelineRunning) == nil {
		var ready bool
		ready, err = r.daemonSetHelper.IsReady(ctx, r.config.DaemonSet)
		if err != nil {
			log.Error(err, "Failed to check Fluent Bit readiness")
			return ctrl.Result{RequeueAfter: controller.RequeueTime}, err
		}
		if !ready {
			log.V(1).Info(fmt.Sprintf("Checked %s - not yet ready. Requeueing...", req.NamespacedName.Name))
			return ctrl.Result{RequeueAfter: controller.RequeueTime}, nil
		}
		log.V(1).Info(fmt.Sprintf("Checked %s - ready", req.NamespacedName.Name))

		condition := telemetryv1alpha1.NewLogPipelineCondition(
			telemetryv1alpha1.FluentBitDSRestartCompletedReason,
			telemetryv1alpha1.LogPipelineRunning,
		)
		pipelineUnsupported := logPipeline.ContainsCustomPlugin()

		if err = r.updateLogPipelineStatus(ctx, req.NamespacedName, condition, pipelineUnsupported); err != nil {
			return ctrl.Result{RequeueAfter: controller.RequeueTime}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) updateLogPipelineStatus(ctx context.Context, name types.NamespacedName, condition *telemetryv1alpha1.LogPipelineCondition, unSupported bool) error {
	log := logf.FromContext(ctx)

	var logPipeline telemetryv1alpha1.LogPipeline
	if err := r.Get(ctx, name, &logPipeline); err != nil {
		log.Error(err, "Failed to get LogPipeline")
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
		log.Error(err, fmt.Sprintf("Failed to update LogPipeline status to %s", condition.Type))
		return err
	}
	return nil
}

func (r *Reconciler) updateMetrics(allPipelines *telemetryv1alpha1.LogPipelineList) {
	r.allLogPipelines.Set(float64(count(allPipelines, isNotMarkedForDeletion)))
	r.unsupportedLogPipelines.Set(float64(count(allPipelines, isUnsupported)))
}

type keepFunc func(*telemetryv1alpha1.LogPipeline) bool

func count(pipelines *telemetryv1alpha1.LogPipelineList, keep keepFunc) int {
	count := 0
	for i := range pipelines.Items {
		if keep(&pipelines.Items[i]) {
			count++
		}
	}
	return count
}

func isNotMarkedForDeletion(pipeline *telemetryv1alpha1.LogPipeline) bool {
	return pipeline.DeletionTimestamp.IsZero()
}

func isUnsupported(pipeline *telemetryv1alpha1.LogPipeline) bool {
	return isNotMarkedForDeletion(pipeline) && pipeline.ContainsCustomPlugin()
}
