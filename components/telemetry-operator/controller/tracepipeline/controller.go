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
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/configchecksum"
	utils "github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes"
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

type Reconciler struct {
	client.Client
	config Config
	Scheme *runtime.Scheme
}

func NewReconciler(client client.Client, config Config, scheme *runtime.Scheme) *Reconciler {
	var r Reconciler
	r.Client = client
	r.config = config
	r.Scheme = scheme
	return &r
}

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

func (r *Reconciler) installOrUpgradeOtelCollector(ctx context.Context, tracing *telemetryv1alpha1.TracePipeline) error {
	var err error

	var secretData map[string][]byte
	if secretData, err = fetchSecretData(ctx, r, tracing.Spec.Output.Otlp); err != nil {
		return err
	}
	secret := makeSecret(r.config, secretData)
	if err = controllerutil.SetControllerReference(tracing, secret, r.Scheme); err != nil {
		return err
	}
	if err = utils.CreateOrUpdateSecret(ctx, r.Client, secret); err != nil {
		return err
	}

	configMap := makeConfigMap(r.config, tracing.Spec.Output)
	if err = controllerutil.SetControllerReference(tracing, configMap, r.Scheme); err != nil {
		return err
	}
	if err = utils.CreateOrUpdateConfigMap(ctx, r.Client, configMap); err != nil {
		return fmt.Errorf("failed to create otel collector configmap: %w", err)
	}

	configHash := configchecksum.Calculate([]corev1.ConfigMap{*configMap}, []corev1.Secret{*secret})
	deployment := makeDeployment(r.config, configHash)
	if err = controllerutil.SetControllerReference(tracing, deployment, r.Scheme); err != nil {
		return err
	}
	if err = utils.CreateOrUpdateDeployment(ctx, r.Client, deployment); err != nil {
		return fmt.Errorf("failed to create otel collector deployment: %w", err)
	}

	service := makeCollectorService(r.config)
	if err = controllerutil.SetControllerReference(tracing, service, r.Scheme); err != nil {
		return err
	}
	if err = utils.CreateOrUpdateService(ctx, r.Client, service); err != nil {
		return fmt.Errorf("failed to create otel collector service: %w", err)
	}

	if r.config.CreateServiceMonitor {
		serviceMonitor := makeServiceMonitor(r.config)
		if err = controllerutil.SetControllerReference(tracing, serviceMonitor, r.Scheme); err != nil {
			return err
		}

		if err = utils.CreateOrUpdateServiceMonitor(ctx, r.Client, serviceMonitor); err != nil {
			return fmt.Errorf("failed to create otel collector prometheus service monitor: %w", err)
		}

		metricsService := makeMetricsService(r.config)
		if err = controllerutil.SetControllerReference(tracing, metricsService, r.Scheme); err != nil {
			return err
		}
		if err = utils.CreateOrUpdateService(ctx, r.Client, metricsService); err != nil {
			return fmt.Errorf("failed to create otel collector metrics service: %w", err)
		}
	}

	return nil
}
