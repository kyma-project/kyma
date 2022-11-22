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

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/configchecksum"
	appsv1 "k8s.io/api/apps/v1"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/kyma-project/kyma/components/telemetry-operator/controller"
	controllermetrics "github.com/kyma-project/kyma/components/telemetry-operator/controller/metrics"
	configbuilder "github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config/builder"
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

const checksumAnnotationKey = "checksum/logpipeline-config"

type Config struct {
	DaemonSet         types.NamespacedName
	SectionsConfigMap types.NamespacedName
	FilesConfigMap    types.NamespacedName
	EnvSecret         types.NamespacedName
	PipelineDefaults  configbuilder.PipelineDefaults
}

//go:generate mockery --name DaemonSetProber --filename daemon_set_prober.go
type DaemonSetProber interface {
	IsReady(ctx context.Context, name types.NamespacedName) (bool, error)
}

//go:generate mockery --name DaemonSetAnnotator --filename daemon_set_annotator.go
type DaemonSetAnnotator interface {
	SetAnnotation(ctx context.Context, name types.NamespacedName, key, value string) (patched bool, err error)
}

type Reconciler struct {
	client.Client
	config                  Config
	syncer                  *syncer
	allLogPipelines         prometheus.Gauge
	unsupportedLogPipelines prometheus.Gauge
	restartsTotal           prometheus.Counter
	prober                  DaemonSetProber
	annotator               DaemonSetAnnotator
	secrets                 secretsCache
}

func NewReconciler(client client.Client, config Config, prober DaemonSetProber, annotator DaemonSetAnnotator) *Reconciler {
	var r Reconciler
	r.Client = client
	r.config = config
	r.syncer = newSyncer(client, config)
	r.allLogPipelines = prometheus.NewGauge(prometheus.GaugeOpts{Name: "telemetry_all_logpipelines", Help: "Number of log pipelines."})
	r.unsupportedLogPipelines = prometheus.NewGauge(prometheus.GaugeOpts{Name: "telemetry_unsupported_logpipelines", Help: "Number of log pipelines with custom filters or outputs."})
	r.restartsTotal = controllermetrics.FluentBitTriggeredRestartsTotal
	r.prober = prober
	r.annotator = annotator
	r.secrets = newSecretsCache()
	metrics.Registry.MustRegister(r.allLogPipelines, r.unsupportedLogPipelines)
	controllermetrics.RegisterMetrics()

	return &r
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1alpha1.LogPipeline{}).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.mapSecrets),
			builder.WithPredicates(onlyUpdate()),
		).
		Watches(
			&source.Kind{Type: &appsv1.DaemonSet{}},
			handler.EnqueueRequestsFromMapFunc(r.mapDaemonSets),
			builder.WithPredicates(onlyUpdate()),
		).
		Complete(r)
}

func onlyUpdate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(event event.CreateEvent) bool { return false },
		DeleteFunc:  func(deleteEvent event.DeleteEvent) bool { return false },
		UpdateFunc:  func(updateEvent event.UpdateEvent) bool { return true },
		GenericFunc: func(genericEvent event.GenericEvent) bool { return false },
	}
}

func (r *Reconciler) mapSecrets(object client.Object) []reconcile.Request {
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

func (r *Reconciler) mapDaemonSets(object client.Object) []reconcile.Request {
	daemonSet := object.(*appsv1.DaemonSet)

	var requests []reconcile.Request
	if daemonSet.Name != r.config.DaemonSet.Name || daemonSet.Namespace != r.config.DaemonSet.Namespace {
		return requests
	}

	var allPipelines telemetryv1alpha1.LogPipelineList
	if err := r.List(context.Background(), &allPipelines); err != nil {
		ctrl.Log.Error(err, "DamonSet UpdateEvent: fetching LogPipelineList failed!", err.Error())
		return requests
	}

	for _, pipeline := range allPipelines.Items {
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{Name: pipeline.Name}})
	}
	ctrl.Log.V(1).Info(fmt.Sprintf("DaemonSet changed event handling done: Created %d new reconciliation requests.\n", len(requests)))
	return requests
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (reconcileResult ctrl.Result, reconcileErr error) {
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

	defer func() {
		if err := r.updateStatus(ctx, pipeline.Name); err != nil {
			reconcileResult = ctrl.Result{Requeue: controller.ShouldRetryOn(err)}
			reconcileErr = fmt.Errorf("failed to update LogPipeline status: %v", err)
		}
	}()

	if err := r.ensureFinalizers(ctx, &pipeline); err != nil {
		return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
	}

	r.syncSecretsCache(&pipeline)

	if err := r.syncer.syncAll(ctx, &pipeline, &allPipelines); err != nil {
		return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
	}

	if err := r.cleanupFinalizersIfNeeded(ctx, &pipeline); err != nil {
		return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
	}

	checksum, err := r.calculateChecksum(ctx)
	if err != nil {
		return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
	}

	patched, err := r.annotator.SetAnnotation(ctx, r.config.DaemonSet, checksumAnnotationKey, checksum)
	if err != nil {
		return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
	}

	if patched {
		r.restartsTotal.Inc()
	}

	return reconcileResult, reconcileErr
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

func (r *Reconciler) calculateChecksum(ctx context.Context) (string, error) {
	var sectionsCm corev1.ConfigMap
	if err := r.Get(ctx, r.config.SectionsConfigMap, &sectionsCm); err != nil {
		return "", fmt.Errorf("failed to get %s/%s ConfigMap: %v", r.config.SectionsConfigMap.Namespace, r.config.SectionsConfigMap.Name, err)
	}

	var filesCm corev1.ConfigMap
	if err := r.Get(ctx, r.config.FilesConfigMap, &filesCm); err != nil {
		return "", fmt.Errorf("failed to get %s/%s ConfigMap: %v", r.config.FilesConfigMap.Namespace, r.config.FilesConfigMap.Name, err)
	}

	var envSecret corev1.Secret
	if err := r.Get(ctx, r.config.EnvSecret, &envSecret); err != nil {
		return "", fmt.Errorf("failed to get %s/%s ConfigMap: %v", r.config.EnvSecret.Namespace, r.config.EnvSecret.Name, err)
	}

	return configchecksum.Calculate([]corev1.ConfigMap{sectionsCm, filesCm}, []corev1.Secret{envSecret}), nil
}
