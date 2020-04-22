package knative

import (
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func (r *ServiceReconciler) getPredicates() predicate.Predicate {
	var log = r.Log.WithName("predicates")
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			srv, ok := e.Object.(*servingv1.Service)
			if !ok {
				return false
			}
			log.Info("Skipping reconciliation for dependent resource creation", "name", srv.Name, "namespace", srv.Namespace, "apiVersion", srv.GroupVersionKind().GroupVersion(), "kind", srv.GroupVersionKind().Kind)
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSrv, ok := e.ObjectNew.(*servingv1.Service)
			if !ok {
				return false
			}

			log.Info("Reconciling due to dependent resource update", "name", newSrv.GetName(), "namespace", newSrv.GetNamespace(), "apiVersion", newSrv.GroupVersionKind().GroupVersion(), "kind", newSrv.GroupVersionKind().Kind)
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			srv, ok := e.Object.(*servingv1.Service)
			if !ok {
				return false
			}
			log.Info("Reconcile due to generic event", "name", srv.GetName(), "namespace", srv.GetNamespace(), "apiVersion", srv.GroupVersionKind().Version, "kind", srv.GroupVersionKind().Kind)
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			srv, ok := e.Object.(*servingv1.Service)
			if !ok {
				return false
			}
			log.Info("Skipping reconciliation for dependent resource deletion", "name", srv.GetName(), "namespace", srv.GetNamespace(), "apiVersion", srv.GroupVersionKind().Version, "kind", srv.GroupVersionKind().Kind)
			return false
		},
	}
}
