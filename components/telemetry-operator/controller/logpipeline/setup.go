package logpipeline

import (
	"context"
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/setup"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1alpha1.LogPipeline{}).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.mapSecret),
			builder.WithPredicates(setup.CreateOrUpdate()),
		).
		Watches(
			&source.Kind{Type: &appsv1.DaemonSet{}},
			handler.EnqueueRequestsFromMapFunc(r.mapDaemonSet),
			builder.WithPredicates(setup.OnlyUpdate()),
		).
		Complete(r)
}

func (r *Reconciler) mapSecret(object client.Object) []reconcile.Request {
	secret := object.(*corev1.Secret)
	var pipelines telemetryv1alpha1.LogPipelineList
	var requests []reconcile.Request
	err := r.List(context.Background(), &pipelines)
	if err != nil {
		ctrl.Log.Error(err, "Secret UpdateEvent: fetching LogPipelineList failed!", err.Error())
		return requests
	}

	ctrl.Log.V(1).Info(fmt.Sprintf("Secret UpdateEvent: handling Secret: %s", secret.Name))
	for i := range pipelines.Items {
		var pipeline = pipelines.Items[i]
		if hasSecretRef(&pipeline, secret.Name, secret.Namespace) {
			request := reconcile.Request{NamespacedName: types.NamespacedName{Name: pipeline.Name}}
			requests = append(requests, request)
			ctrl.Log.V(1).Info(fmt.Sprintf("Secret UpdateEvent: added reconcile request for pipeline: %s", pipeline.Name))
		}
	}
	return requests
}

func (r *Reconciler) mapDaemonSet(object client.Object) []reconcile.Request {
	daemonSet := object.(*appsv1.DaemonSet)

	var requests []reconcile.Request
	if daemonSet.Name != r.config.DaemonSet.Name || daemonSet.Namespace != r.config.DaemonSet.Namespace {
		return requests
	}

	var allPipelines telemetryv1alpha1.LogPipelineList
	if err := r.List(context.Background(), &allPipelines); err != nil {
		ctrl.Log.Error(err, "DaemonSet UpdateEvent: fetching LogPipelineList failed!", err.Error())
		return requests
	}

	for _, pipeline := range allPipelines.Items {
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{Name: pipeline.Name}})
	}
	ctrl.Log.V(1).Info(fmt.Sprintf("DaemonSet changed event handling done: Created %d new reconciliation requests.\n", len(requests)))
	return requests
}
