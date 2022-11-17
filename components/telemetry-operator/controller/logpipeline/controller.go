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

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/kyma-project/kyma/components/telemetry-operator/controller"
	controllermetrics "github.com/kyma-project/kyma/components/telemetry-operator/controller/metrics"
	configbuilder "github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config/builder"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes"
	"github.com/prometheus/client_golang/prometheus"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
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
	PipelineDefaults  configbuilder.PipelineDefaults
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
			handler.EnqueueRequestsFromMapFunc(r.enqueueRequests),
			builder.WithPredicates(predicate.Funcs{
				CreateFunc: func(event event.CreateEvent) bool {
					return false
				},
				DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
					return false
				},
				UpdateFunc: func(updateEvent event.UpdateEvent) bool {
					return true // only handle updates
				},
				GenericFunc: func(genericEvent event.GenericEvent) bool {
					return false
				},
			}),
		).
		Complete(r)
}

func (r *Reconciler) enqueueRequests(object client.Object) []reconcile.Request {
	secret := object.(*corev1.Secret)
	ctrl.Log.V(1).Info(fmt.Sprintf("Secret changed event: Handling Secret with name: %s\n", secret.Name))
	secretName := types.NamespacedName{Namespace: secret.Namespace, Name: secret.Name}
	pipelines := r.secrets.get(secretName)
	var requests []reconcile.Request
	for _, p := range pipelines {
		request := reconcile.Request{NamespacedName: types.NamespacedName{Name: p}}
		ctrl.Log.V(1).Info(fmt.Sprintf("Secret changed event: Creating reconciliation request for pipeline: %s\n", p))
		requests = append(requests, request)
	}
	ctrl.Log.V(1).Info(fmt.Sprintf("Secret changed event handling done: Created %d new reconciliation requests.\n", len(requests)))
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
	log.Info("Reconciliation triggered")

	var allPipelines telemetryv1alpha1.LogPipelineList
	if err := r.List(ctx, &allPipelines); err != nil {
		return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, fmt.Errorf("failed to get all log pipelines: %v", err)
	}

	r.updateMetrics(&allPipelines)

	var pipeline telemetryv1alpha1.LogPipeline
	if err := r.Get(ctx, req.NamespacedName, &pipeline); err != nil {
		log.V(1).Info("Ignoring deleted LogPipeline")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.ensureFinalizers(ctx, &pipeline); err != nil {
		return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
	}

	r.syncSecretsCache(&pipeline)

	secretsOK := r.ensureReferencedSecretsExist(ctx, &pipeline)
	if !secretsOK {
		condition := telemetryv1alpha1.NewLogPipelineCondition(
			telemetryv1alpha1.SecretsNotPresent,
			telemetryv1alpha1.LogPipelinePending,
		)
		pipelineUnsupported := pipeline.ContainsCustomPlugin()
		if err := r.updateLogPipelineStatus(ctx, req.NamespacedName, condition, pipelineUnsupported); err != nil {
			return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
		}

		return ctrl.Result{RequeueAfter: controller.RequeueTime}, nil
	}

	changed, err := r.syncer.syncAll(ctx, &pipeline, &allPipelines)
	if err != nil {
		return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
	}

	if err := r.cleanupFinalizers(ctx, &pipeline); err != nil {
		return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
	}

	if changed {
		log.Info("Fluent Bit configuration was updated. Restarting the DaemonSet")

		if err = r.daemonSetHelper.Restart(ctx, r.config.DaemonSet); err != nil {
			return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, fmt.Errorf("failed to restart Fluent Bit DaemonSet: %v", err)
		}

		condition := telemetryv1alpha1.NewLogPipelineCondition(
			telemetryv1alpha1.FluentBitDSRestartedReason,
			telemetryv1alpha1.LogPipelinePending,
		)
		pipelineUnsupported := pipeline.ContainsCustomPlugin()
		if err = r.updateLogPipelineStatus(ctx, req.NamespacedName, condition, pipelineUnsupported); err != nil {
			return ctrl.Result{RequeueAfter: controller.RequeueTime}, err
		}

		return ctrl.Result{RequeueAfter: controller.RequeueTime}, nil
	}

	if pipeline.Status.GetCondition(telemetryv1alpha1.LogPipelineRunning) == nil {
		var ready bool
		ready, err = r.daemonSetHelper.IsReady(ctx, r.config.DaemonSet)
		if err != nil {
			return ctrl.Result{RequeueAfter: controller.RequeueTime}, fmt.Errorf("failed to check Fluent Bit readiness: %v", err)
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
		pipelineUnsupported := pipeline.ContainsCustomPlugin()

		if err = r.updateLogPipelineStatus(ctx, req.NamespacedName, condition, pipelineUnsupported); err != nil {
			return ctrl.Result{RequeueAfter: controller.RequeueTime}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) syncSecretsCache(pipeline *telemetryv1alpha1.LogPipeline) {
	fields := lookupSecretRefFields(pipeline)

	if pipeline.DeletionTimestamp != nil {
		for _, f := range fields {
			ctrl.Log.V(1).Info(fmt.Sprintf("Remove secret referenced by %s from cache: %s", pipeline.Name, f.secretKeyRef.Name))
			r.secrets.delete(f.secretKeyRef.NamespacedName(), pipeline.Name)
		}
		return
	}

	for _, f := range fields {
		ctrl.Log.V(1).Info(fmt.Sprintf("Add secret referenced by %s to cache: %s", pipeline.Name, f.secretKeyRef.Name))
		r.secrets.addOrUpdate(f.secretKeyRef.NamespacedName(), pipeline.Name)
	}
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

func (r *Reconciler) updateMetrics(allPipelines *telemetryv1alpha1.LogPipelineList) {
	r.allLogPipelines.Set(float64(count(allPipelines, isNotMarkedForDeletion)))
	r.unsupportedLogPipelines.Set(float64(count(allPipelines, isUnsupported)))
}

type keepFunc func(*telemetryv1alpha1.LogPipeline) bool

func count(pipelines *telemetryv1alpha1.LogPipelineList, keep keepFunc) int {
	c := 0
	for i := range pipelines.Items {
		if keep(&pipelines.Items[i]) {
			c++
		}
	}
	return c
}

func isNotMarkedForDeletion(pipeline *telemetryv1alpha1.LogPipeline) bool {
	return pipeline.DeletionTimestamp.IsZero()
}

func isUnsupported(pipeline *telemetryv1alpha1.LogPipeline) bool {
	return isNotMarkedForDeletion(pipeline) && pipeline.ContainsCustomPlugin()
}
