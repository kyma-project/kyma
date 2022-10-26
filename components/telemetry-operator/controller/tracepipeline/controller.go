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

package tracepipeline

import (
	"context"
	"fmt"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/controller"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type Config struct {
	CreateServiceMonitor bool
	CollectorNamespace   string
	ResourceName         string
	CollectorImage       string
}

// Reconciler reconciles a TracePipeline object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	config Config
}

func NewReconciler(
	client client.Client,
	config Config,
	scheme *runtime.Scheme,
) *Reconciler {
	var r Reconciler
	r.Client = client
	r.config = config
	r.Scheme = scheme
	return &r
}

//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=tracepipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=tracepipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	logger.Info("Reconciliation triggered")

	var tracePipeline telemetryv1alpha1.TracePipeline
	if err := r.Get(ctx, req.NamespacedName, &tracePipeline); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err := r.installOrUpgradeOtelCollector(ctx, &tracePipeline)
	return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, err
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	newReconciler := ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1alpha1.TracePipeline{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{})

	if r.config.CreateServiceMonitor {
		newReconciler.Owns(&monitoringv1.ServiceMonitor{})
	}

	return newReconciler.Complete(r)
}

func (r *Reconciler) installOrUpgradeOtelCollector(ctx context.Context, tracing *telemetryv1alpha1.TracePipeline) error {
	configMap := makeConfigMap(r.config, tracing.Spec.Output)
	if err := controllerutil.SetControllerReference(tracing, configMap, r.Scheme); err != nil {
		return err
	}
	if err := createOrUpdateConfigMap(ctx, r.Client, configMap); err != nil {
		return fmt.Errorf("failed to create otel collector configmap: %w", err)
	}

	secret := makeSecret(r.config, tracing.Spec.Output.Otlp)
	if err := controllerutil.SetControllerReference(tracing, secret, r.Scheme); err != nil {
		return err
	}
	if err := createOrUpdateSecret(ctx, r.Client, secret); err != nil {
		return err
	}

	deployment := makeDeployment(r.config)
	if err := controllerutil.SetControllerReference(tracing, deployment, r.Scheme); err != nil {
		return err
	}
	if err := createOrUpdateDeployment(ctx, r.Client, deployment); err != nil {
		return fmt.Errorf("failed to create otel collector deployment: %w", err)
	}

	service := makeCollectorService(r.config)
	if err := controllerutil.SetControllerReference(tracing, service, r.Scheme); err != nil {
		return err
	}
	if err := createOrUpdateService(ctx, r.Client, service); err != nil {
		return fmt.Errorf("failed to create otel collector service: %w", err)
	}

	if r.config.CreateServiceMonitor {
		serviceMonitor := makeServiceMonitor(r.config)
		if err := controllerutil.SetControllerReference(tracing, serviceMonitor, r.Scheme); err != nil {
			return err
		}

		if err := createOrUpdateServiceMonitor(ctx, r.Client, serviceMonitor); err != nil {
			return fmt.Errorf("failed to create otel collector prometheus service monitor: %w", err)
		}

		metricsService := makeMetricsService(r.config)
		if err := controllerutil.SetControllerReference(tracing, metricsService, r.Scheme); err != nil {
			return err
		}
		if err := createOrUpdateService(ctx, r.Client, metricsService); err != nil {
			return fmt.Errorf("failed to create otel collector metrics service: %w", err)
		}
	}

	return nil
}
