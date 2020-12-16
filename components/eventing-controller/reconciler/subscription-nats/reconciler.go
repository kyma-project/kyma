package subscription_nats

import (
	"github.com/go-logr/logr"

	"k8s.io/client-go/tools/record"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

// Reconciler reconciles a Subscription object
type Reconciler struct {
	client.Client
	cache.Cache
	Log      logr.Logger
	recorder record.EventRecorder
}

func NewReconciler(client client.Client, cache cache.Cache, log logr.Logger, recorder record.EventRecorder) *Reconciler {
	return &Reconciler{
		Client:   client,
		Cache:    cache,
		Log:      log,
		recorder: recorder,
	}
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventingv1alpha1.Subscription{}).
		Complete(r)
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("received subscription reconciliation request", "namespace", req.Namespace, "name", req.Name)

	return ctrl.Result{}, nil
}
