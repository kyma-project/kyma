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
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/configchecksum"
	configbuilder "github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config/builder"
	utils "github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/overrides"
	resources "github.com/kyma-project/kyma/components/telemetry-operator/internal/resources/logpipeline"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

type Config struct {
	DaemonSet         types.NamespacedName
	SectionsConfigMap types.NamespacedName
	FilesConfigMap    types.NamespacedName
	EnvSecret         types.NamespacedName
	OverrideConfigMap types.NamespacedName
	PipelineDefaults  configbuilder.PipelineDefaults
	Overrides         overrides.Config
	DaemonSetConfig   resources.DaemonSetConfig
}

//go:generate mockery --name DaemonSetProber --filename daemon_set_prober.go
type DaemonSetProber interface {
	IsReady(ctx context.Context, name types.NamespacedName) (bool, error)
}

//go:generate mockery --name DaemonSetAnnotator --filename daemon_set_annotator.go
type DaemonSetAnnotator interface {
	SetAnnotation(ctx context.Context, name types.NamespacedName, key, value string) error
}

type Reconciler struct {
	client.Client
	config                  Config
	prober                  DaemonSetProber
	allLogPipelines         prometheus.Gauge
	unsupportedLogPipelines prometheus.Gauge
	syncer                  syncer
	globalConfig            overrides.GlobalConfigHandler
}

func NewReconciler(client client.Client, config Config, prober DaemonSetProber, handler *overrides.Handler) *Reconciler {
	var r Reconciler
	r.Client = client
	r.config = config
	r.prober = prober
	r.allLogPipelines = prometheus.NewGauge(prometheus.GaugeOpts{Name: "telemetry_all_logpipelines", Help: "Number of log pipelines."})
	r.unsupportedLogPipelines = prometheus.NewGauge(prometheus.GaugeOpts{Name: "telemetry_unsupported_logpipelines", Help: "Number of log pipelines with custom filters or outputs."})
	metrics.Registry.MustRegister(r.allLogPipelines, r.unsupportedLogPipelines)
	r.syncer = syncer{client, config}
	r.globalConfig = handler

	return &r
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.V(1).Info("Reconciliation triggered")

	overrideConfig, err := r.globalConfig.UpdateOverrideConfig(ctx, r.config.OverrideConfigMap)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.globalConfig.CheckGlobalConfig(overrideConfig.Global); err != nil {
		return ctrl.Result{}, err
	}

	if overrideConfig.Logging.Paused {
		log.V(1).Info("Skipping reconciliation of logpipeline as reconciliation is paused.")
		return ctrl.Result{}, nil
	}

	if err := r.updateMetrics(ctx); err != nil {
		log.Error(err, "Failed to get all LogPipelines while updating metrics")
	}

	var pipeline telemetryv1alpha1.LogPipeline
	if err := r.Get(ctx, req.NamespacedName, &pipeline); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, r.doReconcile(ctx, &pipeline)
}

func (r *Reconciler) doReconcile(ctx context.Context, pipeline *telemetryv1alpha1.LogPipeline) (err error) {
	// defer the updating of status to ensure that the status is updated regardless of the outcome of the reconciliation
	defer func() {
		if statusErr := r.updateStatus(ctx, pipeline.Name); statusErr != nil {
			if err != nil {
				err = fmt.Errorf("failed while updating status: %v: %v", statusErr, err)
			} else {
				err = fmt.Errorf("failed to update status: %v", statusErr)
			}
		}
	}()

	if err = ensureFinalizers(ctx, r.Client, pipeline); err != nil {
		return err
	}

	if err = r.syncer.syncFluentBitConfig(ctx, pipeline); err != nil {
		return err
	}

	var checksum string
	if checksum, err = r.calculateChecksum(ctx); err != nil {
		return err
	}

	name := r.config.DaemonSet
	if err = r.reconcileFluentBit(ctx, name, pipeline, checksum); err != nil {
		return err
	}

	if err = cleanupFinalizersIfNeeded(ctx, r.Client, pipeline); err != nil {
		return err
	}

	return err
}

func (r *Reconciler) reconcileFluentBit(ctx context.Context, name types.NamespacedName, pipeline *telemetryv1alpha1.LogPipeline, checksum string) error {
	shouldDeleteFluentBit, err := r.isLastPipelineMarkedForDeletion(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("failed to check if LogPipeline is last marked for deletion: %v", err)
	}

	if shouldDeleteFluentBit {
		return utils.DeleteFluentBit(ctx, r, name)
	}

	serviceAccount := resources.MakeServiceAccount(name)
	if err := utils.CreateOrUpdateServiceAccount(ctx, r, serviceAccount); err != nil {
		return fmt.Errorf("failed to create fluent bit service account: %w", err)
	}
	clusterRole := resources.MakeClusterRole(name)
	if err := utils.CreateOrUpdateClusterRole(ctx, r, clusterRole); err != nil {
		return fmt.Errorf("failed to create fluent bit cluster role: %w", err)
	}
	clusterRoleBinding := resources.MakeClusterRoleBinding(name)
	if err := utils.CreateOrUpdateClusterRoleBinding(ctx, r, clusterRoleBinding); err != nil {
		return fmt.Errorf("failed to create fluent bit cluster role Binding: %w", err)
	}
	daemonSet := resources.MakeDaemonSet(name, checksum, r.config.DaemonSetConfig)
	if err := utils.CreateOrUpdateDaemonSet(ctx, r, daemonSet); err != nil {
		return fmt.Errorf("failed to reconcile fluent bit daemonset: %w", err)
	}
	exporterMetricsService := resources.MakeExporterMetricsService(name)
	if err := utils.CreateOrUpdateService(ctx, r, exporterMetricsService); err != nil {
		return fmt.Errorf("failed to reconcile exporter metrics service: %w", err)
	}
	metricsService := resources.MakeMetricsService(name)
	if err := utils.CreateOrUpdateService(ctx, r, metricsService); err != nil {
		return fmt.Errorf("failed to reconcile fluent bit metrics service: %w", err)
	}
	cm := resources.MakeConfigMap(name)
	if err := utils.CreateOrUpdateConfigMap(ctx, r, cm); err != nil {
		return fmt.Errorf("failed to reconcile fluent bit configmap: %w", err)
	}
	luaCm := resources.MakeLuaConfigMap(name)
	if err := utils.CreateOrUpdateConfigMap(ctx, r, luaCm); err != nil {
		return fmt.Errorf("failed to reconcile fluent bit lua configmap: %w", err)
	}
	parsersCm := resources.MakeDynamicParserConfigmap(name)
	if err := utils.CreateIfNotExistsConfigMap(ctx, r, parsersCm); err != nil {
		return fmt.Errorf("failed to reconcile fluent bit parser configmap: %w", err)
	}
	return nil
}

func (r *Reconciler) isLastPipelineMarkedForDeletion(ctx context.Context, pipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	if isNotMarkedForDeletion(pipeline) {
		return false, nil
	}

	var allPipelines telemetryv1alpha1.LogPipelineList
	if err := r.List(ctx, &allPipelines); err != nil {
		return false, fmt.Errorf("failed to list LogPipelines: %v", err)
	}

	return len(allPipelines.Items) == 1 && allPipelines.Items[0].Name == pipeline.Name, nil
}

func (r *Reconciler) updateMetrics(ctx context.Context) error {
	var allPipelines telemetryv1alpha1.LogPipelineList
	if err := r.List(ctx, &allPipelines); err != nil {
		return err
	}

	r.allLogPipelines.Set(float64(count(&allPipelines, isNotMarkedForDeletion)))
	r.unsupportedLogPipelines.Set(float64(count(&allPipelines, isUnsupported)))

	return nil
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
	return pipeline.ObjectMeta.DeletionTimestamp.IsZero()
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
