package knative

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func (r *ServiceReconciler) getPredicate() predicate.Predicate {
	var log = r.Log.WithName("predicates")
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			o := e.Object.(*unstructured.Unstructured)
			log.Info("Skipping reconciliation for dependent resource creation", "name", o.GetName(), "namespace", o.GetNamespace(), "apiVersion", o.GroupVersionKind().GroupVersion(), "kind", o.GroupVersionKind().Kind)
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			newObj := e.ObjectNew.(*unstructured.Unstructured).DeepCopy()
			log.Info("Reconciling due to dependent resource update", "name", newObj.GetName(), "namespace", newObj.GetNamespace(), "apiVersion", newObj.GroupVersionKind().GroupVersion(), "kind", newObj.GroupVersionKind().Kind)
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			o := e.Object.(*unstructured.Unstructured)
			log.Info("Reconcile due to generic event", "name", o.GetName(), "namespace", o.GetNamespace(), "apiVersion", o.GroupVersionKind().GroupVersion(), "kind", o.GroupVersionKind().Kind)
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			o := e.Object.(*unstructured.Unstructured)
			log.Info("Skipping reconciliation for dependent resource deletion", "name", o.GetName(), "namespace", o.GetNamespace(), "apiVersion", o.GroupVersionKind().GroupVersion(), "kind", o.GroupVersionKind().Kind)
			return false
		},
	}
}
