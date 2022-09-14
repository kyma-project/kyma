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

	"github.com/kyma-project/kyma/components/telemetry-operator/controller"
	controllermetrics "github.com/kyma-project/kyma/components/telemetry-operator/controller/metrics"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config/builder"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/prometheus/client_golang/prometheus"
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
	secrets                 []string
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
	r.allLogPipelines = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "telemetry_all_logpipelines",
		Help: "Number of log pipelines.",
	})
	r.unsupportedLogPipelines = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "telemetry_unsupported_logpipelines",
		Help: "Number of log pipelines with custom filters or outputs.",
	})
	r.daemonSetHelper = kubernetes.NewDaemonSetHelper(client, controllermetrics.FluentBitTriggeredRestartsTotal)

	metrics.Registry.MustRegister(r.allLogPipelines, r.unsupportedLogPipelines)
	controllermetrics.RegisterMetrics()

	return &r
}

func (r *Reconciler) getDesiredSecrets() []string {
	return r.secrets
}

func (r *Reconciler) checkSecrets() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			desiredSecrets := r.getDesiredSecrets()
			_, ok := e.ObjectNew.(*corev1.Secret)
			if !ok {
				return true
			}

			// Ignore updates to CR status in which case metadata.Generation does not change
			if slices.Contains(desiredSecrets, e.ObjectNew.GetName()) {
				fmt.Printf("Update event ALLOW: %s", e.ObjectNew.GetName())
				return true
			} else {
				//fmt.Printf("Update event DENY: %s", e.ObjectNew.GetName())
				return false
			}
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			_, ok := e.Object.(*corev1.Secret)
			if !ok {
				return true
			}
			// Evaluates to false if the object has been confirmed deleted.
			if slices.Contains(r.secrets, e.Object.GetName()) {
				fmt.Printf("Delete event ALLOW: %s", e.Object.GetName())
				return true
			} else {
				//fmt.Printf("Delete event DENY: %s", e.Object.GetName())
				return false
			}
		},
		CreateFunc: func(e event.CreateEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			_, ok := e.Object.(*corev1.Secret)
			if !ok {
				return true
			}
			if slices.Contains(r.secrets, e.Object.GetName()) {
				fmt.Printf("Create event ALLOW: %s", e.Object.GetName())
				return true
			} else {
				//fmt.Printf("Create event DENY: %s", e.Object.GetName())
				return false
			}
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1alpha1.LogPipeline{}).
		WithEventFilter(r.checkSecrets()).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.prepareReconciliationsForSecret),
		).
		Complete(r)
}
func (r *Reconciler) prepareReconciliationsForSecret(secret client.Object) []reconcile.Request {
	fmt.Printf("secret Name: %s\n", secret.GetName())

	if slices.Contains(r.secrets, secret.GetName()) {
		fmt.Printf("It contains: %s\n", secret.GetName())
	} else {
		fmt.Printf("not present: %s\n", secret.GetName())

	}

	//pipelines := r.findLogPipelinesForSecret(secret)
	requests := make([]reconcile.Request, 0)
	//for i := range pipelines.Items {
	//	requests[i] = reconcile.Request{
	//		NamespacedName: types.NamespacedName{
	//			Name: pipelines.Items[i].GetName(),
	//		},
	//	}
	//}
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

	log.Info("RECON")

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
	r.secrets = []string{"mysecret"}
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
