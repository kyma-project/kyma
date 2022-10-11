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
}

// Reconciler reconciles a TracePipeline object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config
}

func NewReconciler(
	client client.Client,
	config Config,
	scheme *runtime.Scheme,
) *Reconciler {
	var r Reconciler
	r.Client = client
	r.Config = config
	r.Scheme = scheme
	return &r
}

//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=tracepipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=tracepipelines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
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

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	newReconciler := ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1alpha1.TracePipeline{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{})

	if r.Config.CreateServiceMonitor {
		newReconciler.Owns(&monitoringv1.ServiceMonitor{})
	}

	return newReconciler.Complete(r)
}

func (r *Reconciler) installOrUpgradeOtelCollector(ctx context.Context, tracing *telemetryv1alpha1.TracePipeline) error {
	configMap := makeConfigMap(tracing.Spec.Output)
	if err := controllerutil.SetControllerReference(tracing, configMap, r.Scheme); err != nil {
		return err
	}
	if err := createOrUpdateConfigMap(ctx, r.Client, configMap); err != nil {
		return fmt.Errorf("failed to create otel collector configmap: %w", err)
	}

	deployment := makeDeployment()
	if err := controllerutil.SetControllerReference(tracing, deployment, r.Scheme); err != nil {
		return err
	}
	if err := createOrUpdateDeployment(ctx, r.Client, deployment); err != nil {
		return fmt.Errorf("failed to create otel collector deployment: %w", err)
	}

	service := makeService()
	if err := controllerutil.SetControllerReference(tracing, service, r.Scheme); err != nil {
		return err
	}
	if err := createOrUpdateService(ctx, r.Client, service); err != nil {
		return fmt.Errorf("failed to create otel collector service: %w", err)
	}

	if r.Config.CreateServiceMonitor {
		serviceMonitor := makeServiceMonitor()
		if err := controllerutil.SetControllerReference(tracing, serviceMonitor, r.Scheme); err != nil {
			return err
		}

		if err := createOrUpdateServiceMonitor(ctx, r.Client, serviceMonitor); err != nil {
			return fmt.Errorf("failed to create otel collector prometheus service monitor: %w", err)
		}
	}

	return nil
}
