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
	"errors"
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/configchecksum"
	utils "github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/overrides"
	corev1 "k8s.io/api/core/v1"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type Config struct {
	BaseName          string
	Namespace         string
	OverrideConfigMap types.NamespacedName

	Deployment DeploymentConfig
	Service    ServiceConfig
	Overrides  overrides.Config
}

type DeploymentConfig struct {
	Image             string
	PriorityClassName string
	CPULimit          resource.Quantity
	MemoryLimit       resource.Quantity
	CPURequest        resource.Quantity
	MemoryRequest     resource.Quantity
}

type ServiceConfig struct {
	OTLPServiceName string
}

//go:generate mockery --name DeploymentProber --filename deployment_prober.go
type DeploymentProber interface {
	IsReady(ctx context.Context, name types.NamespacedName) (bool, error)
}

type Reconciler struct {
	client.Client
	config           Config
	Scheme           *runtime.Scheme
	prober           DeploymentProber
	overridesHandler overrides.GlobalConfigHandler
}

func NewReconciler(client client.Client, config Config, prober DeploymentProber, scheme *runtime.Scheme, handler *overrides.Handler) *Reconciler {
	var r Reconciler
	r.Client = client
	r.config = config
	r.Scheme = scheme
	r.prober = prober
	r.overridesHandler = handler
	return &r
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (reconcileResult ctrl.Result, reconcileErr error) {
	log := logf.FromContext(ctx)

	log.V(1).Info("Reconciliation triggered")

	overrideConfig, err := r.overridesHandler.UpdateOverrideConfig(ctx, r.config.OverrideConfigMap)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.overridesHandler.CheckGlobalConfig(overrideConfig.Global); err != nil {
		return ctrl.Result{}, err
	}
	if overrideConfig.Tracing.Paused {
		log.V(1).Info("Skipping reconciliation of tracepipeline as reconciliation is paused")
		return ctrl.Result{}, nil
	}

	var tracePipeline telemetryv1alpha1.TracePipeline
	if err := r.Get(ctx, req.NamespacedName, &tracePipeline); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, r.doReconcile(ctx, &tracePipeline)
}

func (r *Reconciler) doReconcile(ctx context.Context, pipeline *telemetryv1alpha1.TracePipeline) error {
	var err error
	lockAcquired := true

	defer func() {
		if statusErr := r.updateStatus(ctx, pipeline.Name, lockAcquired); statusErr != nil {
			if err != nil {
				err = fmt.Errorf("failed while updating status: %v: %v", statusErr, err)
			} else {
				err = fmt.Errorf("failed to update status: %v", statusErr)
			}
		}
	}()

	if err = r.tryAcquireLock(ctx, pipeline); err != nil {
		lockAcquired = false
		return err
	}

	var secretData map[string][]byte
	if secretData, err = fetchSecretData(ctx, r, pipeline.Spec.Output.Otlp); err != nil {
		return err
	}
	secret := makeSecret(r.config, secretData)
	if err = controllerutil.SetControllerReference(pipeline, secret, r.Scheme); err != nil {
		return err
	}
	if err = utils.CreateOrUpdateSecret(ctx, r.Client, secret); err != nil {
		return err
	}
	endpoint := string(secretData[otlpEndpointVariable])
	isInsecure := isInsecureOutput(endpoint)

	collectorConfig := makeOtelCollectorConfig(pipeline.Spec.Output, isInsecure)
	configMap := makeConfigMap(r.config, collectorConfig)
	if err = controllerutil.SetControllerReference(pipeline, configMap, r.Scheme); err != nil {
		return err
	}
	if err = utils.CreateOrUpdateConfigMap(ctx, r.Client, configMap); err != nil {
		return fmt.Errorf("failed to create otel collector configmap: %w", err)
	}

	configHash := configchecksum.Calculate([]corev1.ConfigMap{*configMap}, []corev1.Secret{*secret})
	deployment := makeDeployment(r.config, configHash)
	if err = controllerutil.SetControllerReference(pipeline, deployment, r.Scheme); err != nil {
		return err
	}
	if err = utils.CreateOrUpdateDeployment(ctx, r.Client, deployment); err != nil {
		return fmt.Errorf("failed to create otel collector deployment: %w", err)
	}

	otlpService := makeOTLPService(r.config)
	if err = controllerutil.SetControllerReference(pipeline, otlpService, r.Scheme); err != nil {
		return err
	}
	if err = utils.CreateOrUpdateService(ctx, r.Client, otlpService); err != nil {
		return fmt.Errorf("failed to create otel collector otlp service: %w", err)
	}

	openCensusService := makeOpenCensusService(r.config)
	if err = controllerutil.SetControllerReference(pipeline, openCensusService, r.Scheme); err != nil {
		return err
	}
	if err = utils.CreateOrUpdateService(ctx, r.Client, openCensusService); err != nil {
		return fmt.Errorf("failed to create otel collector open census service: %w", err)
	}

	metricsService := makeMetricsService(r.config)
	if err = controllerutil.SetControllerReference(pipeline, metricsService, r.Scheme); err != nil {
		return err
	}
	if err = utils.CreateOrUpdateService(ctx, r.Client, metricsService); err != nil {
		return fmt.Errorf("failed to create otel collector metrics service: %w", err)
	}

	return nil
}

func isInsecureOutput(endpoint string) bool {
	return len(strings.TrimSpace(endpoint)) > 0 && strings.HasPrefix(endpoint, "http://")
}

func (r *Reconciler) tryAcquireLock(ctx context.Context, pipeline *telemetryv1alpha1.TracePipeline) error {
	lockName := types.NamespacedName{Name: "telemetry-tracepipeline-lock", Namespace: r.config.Namespace}
	var lock corev1.ConfigMap
	if err := r.Get(ctx, lockName, &lock); err != nil {
		if apierrors.IsNotFound(err) {
			return r.createLock(ctx, lockName, pipeline)
		}
		return fmt.Errorf("failed to get lock: %v", err)
	}

	for _, ref := range lock.GetOwnerReferences() {
		if ref.Name == pipeline.Name && ref.UID == pipeline.UID {
			return nil
		}
	}

	return errors.New("lock is already acquired by another TracePipeline")
}

func (r *Reconciler) createLock(ctx context.Context, name types.NamespacedName, owner *telemetryv1alpha1.TracePipeline) error {
	lock := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
	}
	if err := controllerutil.SetControllerReference(owner, &lock, r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference: %v", err)
	}
	if err := r.Create(ctx, &lock); err != nil {
		return fmt.Errorf("failed to create lock: %v", err)
	}
	return nil
}
