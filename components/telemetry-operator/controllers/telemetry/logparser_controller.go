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
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/parsers"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/sync"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// LogParserReconciler reconciles a LogParser object
type LogParserReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	FluentBitDaemonSet types.NamespacedName
	Parser             parsers.LogParserSyncer
	FluentBitUtils     *kubernetes.FluentBitUtils
	restartsTotal      prometheus.Counter
}

//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logparsers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logparsers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logparsers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LogParser object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile

func NewLogParserReconciler(client client.Client, scheme *runtime.Scheme, daemonSetConfig sync.FluentBitDaemonSetConfig, restartsTotal prometheus.Counter) *LogParserReconciler {
	var lpr LogParserReconciler
	lpr.Client = client
	lpr.Scheme = scheme
	lpr.FluentBitDaemonSet = daemonSetConfig.FluentBitDaemonSetName
	lpr.FluentBitUtils = kubernetes.NewFluentBit(client)
	lpr.restartsTotal = restartsTotal

	prometheus.MustRegister(lpr.restartsTotal)
	return &lpr
}

func (r *LogParserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	var logParser telemetryv1alpha1.LogParser
	if err := r.Get(ctx, req.NamespacedName, &logParser); err != nil {
		log.Info("Ignoring deleted LogParser")
		// Ignore not-found errors since we can get them on deleted requests
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var changed, err = r.Parser.SyncParsersConfigMap(ctx, &logParser)
	if err != nil {
		return ctrl.Result{Requeue: shouldRetryOn(err)}, nil
	}

	if changed {
		log.V(1).Info("Fluent Bit configuration was updated. Restarting the DaemonSet")
		if err = r.FluentBitUtils.RestartFluentBit(ctx, r.restartsTotal); err != nil {
			log.Error(err, "Failed restarting fluent bit daemon set")
			return ctrl.Result{Requeue: shouldRetryOn(err)}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LogParserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1alpha1.LogParser{}).
		Complete(r)
}
