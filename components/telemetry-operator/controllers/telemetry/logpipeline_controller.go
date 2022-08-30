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

package telemetry

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/handler" // Required for Watching
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config/builder"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/sync"

	"github.com/prometheus/client_golang/prometheus"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/source" // Required for Watching
)

var (
	requeueTime = 10 * time.Second
)

const (
	httpOutputSecretPathRef = "spec.output.http"
	varSecretPathRef        = "spec.variables"
)

// LogPipelineReconciler reconciles a LogPipeline object
type LogPipelineReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	Syncer             *sync.LogPipelineSyncer
	FluentBitDaemonSet types.NamespacedName
	unsupportedTotal   prometheus.Gauge
	DaemonSetUtils     *fluentbit.DaemonSetUtils
}

// NewLogPipelineReconciler returns a new LogPipelineReconciler using the given FluentBit config arguments
func NewLogPipelineReconciler(client client.Client, scheme *runtime.Scheme, daemonSetConfig sync.FluentBitDaemonSetConfig, pipelineConfig builder.PipelineConfig, restartsTotal prometheus.Counter) *LogPipelineReconciler {
	var lpr LogPipelineReconciler
	lpr.Client = client
	lpr.Scheme = scheme
	lpr.Syncer = sync.NewLogPipelineSyncer(client,
		daemonSetConfig,
		pipelineConfig,
	)
	lpr.FluentBitDaemonSet = daemonSetConfig.FluentBitDaemonSetName

	lpr.unsupportedTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "telemetry_plugins_unsupported_total",
		Help: "Number of custom filters or outputs to indicate unsupported mode.",
	})
	lpr.DaemonSetUtils = fluentbit.NewDaemonSetUtils(client, daemonSetConfig.FluentBitDaemonSetName, restartsTotal)

	metrics.Registry.MustRegister(lpr.unsupportedTotal)

	return &lpr
}

func secretsArePresent(logPipeline *telemetryv1alpha1.LogPipeline) bool {
	secretPresent := false
	if len(logPipeline.Spec.Variables) > 0 {
		for _, v := range logPipeline.Spec.Variables {
			if v.ValueFrom.IsSecretRef() {
				secretPresent = true
				break
			}
		}
	}
	if logPipeline.Spec.Output.HTTP.Host.IsDefined() {
		httpOutput := logPipeline.Spec.Output.HTTP
		if httpOutput.User.ValueFrom.IsSecretRef() || httpOutput.Password.ValueFrom.IsSecretRef() || httpOutput.Host.ValueFrom.IsSecretRef() {
			secretPresent = true
		}
	}
	return secretPresent
}

func firstHttpSecret(logPipeline *telemetryv1alpha1.LogPipeline) string {
	httpOutput := logPipeline.Spec.Output.HTTP
	if httpOutput.Host.ValueFrom.SecretKey.Name != "" {
		return httpOutput.Host.ValueFrom.SecretKey.Name
	}
	if httpOutput.User.ValueFrom.SecretKey.Name != "" {
		return httpOutput.Host.ValueFrom.SecretKey.Name
	}
	if httpOutput.Password.ValueFrom.SecretKey.Name != "" {
		return httpOutput.Host.ValueFrom.SecretKey.Name
	}
	return ""
}

func indexSecrets(fieldName string, mgr ctrl.Manager) error {
	return mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1alpha1.LogPipeline{}, fieldName, func(rawObj client.Object) []string {
		// Extract the secret name from the logPipeline Spec, if one is provided
		logPipeline := rawObj.(*telemetryv1alpha1.LogPipeline)
		// verify if the secret is being used in variables or httpOutput
		if !secretsArePresent(logPipeline) {
			return nil
		}

		var retStr []string
		if fieldName == httpOutputSecretPathRef {
			secretName := firstHttpSecret(logPipeline)
			if secretName == "" {
				return nil
			}
			retStr = append(retStr, secretName)
		} else if fieldName == varSecretPathRef {
			for _, v := range logPipeline.Spec.Variables {
				if v.ValueFrom.SecretKey.Name == "" {
					return nil
				}
				retStr = append(retStr, v.ValueFrom.SecretKey.Name)
			}
		}

		return retStr
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *LogPipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := indexSecrets(httpOutputSecretPathRef, mgr); err != nil {
		return err
	}
	if err := indexSecrets(varSecretPathRef, mgr); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1alpha1.LogPipeline{}).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.reconcilePipelineWithSecret),
		).
		Complete(r)
}

func (r *LogPipelineReconciler) fetchLogPipelineToReconcile(secret client.Object) *telemetryv1alpha1.LogPipelineList {
	var logPipelines telemetryv1alpha1.LogPipelineList
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(varSecretPathRef, secret.GetName()),
	}
	err := r.List(context.TODO(), &logPipelines, listOps)
	if err != nil {
		return nil
	}
	if len(logPipelines.Items) == 0 {
		listOps = &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(httpOutputSecretPathRef, secret.GetName()),
		}
	}
	err = r.List(context.TODO(), &logPipelines, listOps)
	if err != nil {
		return nil
	}
	return &logPipelines
}

func (r *LogPipelineReconciler) reconcilePipelineWithSecret(secret client.Object) []reconcile.Request {
	logPipelines := r.fetchLogPipelineToReconcile(secret)
	requests := make([]reconcile.Request, len(logPipelines.Items))
	for i, item := range logPipelines.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: item.GetName(),
			},
		}
	}
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
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *LogPipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.Info("RECONCILING!!")
	var logPipeline telemetryv1alpha1.LogPipeline
	if err := r.Get(ctx, req.NamespacedName, &logPipeline); err != nil {
		log.Info("Ignoring deleted LogPipeline")
		// Ignore not-found errors since we can get them on deleted requests
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	secretsOK := r.Syncer.SecretHelper.ValidateSecretsExist(ctx, &logPipeline)
	if !secretsOK {
		condition := telemetryv1alpha1.NewLogPipelineCondition(
			telemetryv1alpha1.SecretsNotPresent,
			telemetryv1alpha1.LogPipelinePending,
		)
		pipeLineUnsupported := sync.LogPipelineIsUnsupported(logPipeline)
		if err := r.updateLogPipelineStatus(ctx, req.NamespacedName, condition, pipeLineUnsupported); err != nil {
			return ctrl.Result{Requeue: shouldRetryOn(err)}, nil
		}

		return ctrl.Result{RequeueAfter: requeueTime}, nil
	}

	changed, err := r.Syncer.SyncAll(ctx, &logPipeline)
	if err != nil {
		return ctrl.Result{Requeue: shouldRetryOn(err)}, nil
	}

	r.unsupportedTotal.Set(float64(r.Syncer.UnsupportedPluginsTotal))

	if changed {
		log.V(1).Info("Fluent Bit configuration was updated. Restarting the DaemonSet due to logpipeline change")

		if err = r.Update(ctx, &logPipeline); err != nil {
			log.Error(err, "Failed updating log pipeline")
			return ctrl.Result{Requeue: shouldRetryOn(err)}, err
		}

		if err = r.DaemonSetUtils.RestartFluentBit(ctx); err != nil {
			log.Error(err, "Failed restarting fluent bit daemon set")
			return ctrl.Result{Requeue: shouldRetryOn(err)}, err
		}

		condition := telemetryv1alpha1.NewLogPipelineCondition(
			telemetryv1alpha1.FluentBitDSRestartedReason,
			telemetryv1alpha1.LogPipelinePending,
		)
		pipeLineUnsupported := sync.LogPipelineIsUnsupported(logPipeline)
		if err = r.updateLogPipelineStatus(ctx, req.NamespacedName, condition, pipeLineUnsupported); err != nil {
			return ctrl.Result{RequeueAfter: requeueTime}, err
		}

		return ctrl.Result{RequeueAfter: requeueTime}, nil
	}

	if logPipeline.Status.GetCondition(telemetryv1alpha1.LogPipelineRunning) == nil {
		var ready bool
		ready, err = r.DaemonSetUtils.IsFluentBitDaemonSetReady(ctx)
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
		pipeLineUnsupported := sync.LogPipelineIsUnsupported(logPipeline)

		if err = r.updateLogPipelineStatus(ctx, req.NamespacedName, condition, pipeLineUnsupported); err != nil {
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
		!errors.IsForbidden(err) &&
		!errors.IsNotFound(err)
}

func (r *LogPipelineReconciler) updateLogPipelineStatus(ctx context.Context, name types.NamespacedName, condition *telemetryv1alpha1.LogPipelineCondition, unSupported bool) error {
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
	logPipeline.Status.UnsupportedMode = unSupported

	if err := r.Status().Update(ctx, &logPipeline); err != nil {
		log.Error(err, fmt.Sprintf("Failed updating log pipeline status to %s", condition.Type))
		return err
	}
	return nil
}
