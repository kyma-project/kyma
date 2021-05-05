package backend

import (
	"github.com/go-logr/logr"

	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BackendReconciler controls the switching between different eventing controller backends.
type BackendReconciler struct {
	client.Client
	Log logr.Logger
}

func (r *BackendReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
func (r *BackendReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&v1.Secret{}).Complete(r)
}
