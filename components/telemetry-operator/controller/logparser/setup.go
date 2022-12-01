package logparser

import (
	"context"
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	onlyUpdate = predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool { return false },
		DeleteFunc: func(e event.DeleteEvent) bool { return false },
		UpdateFunc: func(e event.UpdateEvent) bool {
			// We only need to check generation change here, because it is only updated on spec changes.
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}
)

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1alpha1.LogParser{}).
		Watches(
			&source.Kind{Type: &appsv1.DaemonSet{}},
			handler.EnqueueRequestsFromMapFunc(r.mapDaemonSets),
			builder.WithPredicates(onlyUpdate),
		).
		Complete(r)
}

func (r *Reconciler) mapDaemonSets(object client.Object) []reconcile.Request {
	daemonSet := object.(*appsv1.DaemonSet)

	var requests []reconcile.Request
	if daemonSet.Name != r.config.DaemonSet.Name || daemonSet.Namespace != r.config.DaemonSet.Namespace {
		return requests
	}

	var allParsers telemetryv1alpha1.LogParserList
	if err := r.List(context.Background(), &allParsers); err != nil {
		ctrl.Log.Error(err, "DamonSet UpdateEvent: fetching LogParserList failed!", err.Error())
		return requests
	}

	for _, parser := range allParsers.Items {
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{Name: parser.Name}})
	}
	ctrl.Log.V(1).Info(fmt.Sprintf("DaemonSet changed event handling done: Created %d new reconciliation requests.\n", len(requests)))
	return requests
}
