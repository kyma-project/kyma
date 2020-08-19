package kubernetes

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type ServiceAccountReconciler struct {
	Log    logr.Logger
	client client.Client
	svc    ServiceAccountService
	config Config
}

func NewServiceAccount(client client.Client, log logr.Logger, config Config, serviceAccountSvc ServiceAccountService) *ServiceAccountReconciler {
	return &ServiceAccountReconciler{
		client: client,
		Log:    log.WithName("controllers").WithName("serviceaccount").WithValues("kind", "ServiceAccount"),
		config: config,
		svc:    serviceAccountSvc,
	}
}

func (r *ServiceAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("serviceaccount-controller").
		For(&corev1.ServiceAccount{}).
		WithEventFilter(r.predicate()).
		Complete(r)
}

func (r *ServiceAccountReconciler) predicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			runtime, ok := e.Object.(*corev1.ServiceAccount)
			if !ok {
				return false
			}
			return r.svc.IsBase(runtime)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			runtime, ok := e.ObjectNew.(*corev1.ServiceAccount)
			if !ok {
				return false
			}
			return r.svc.IsBase(runtime)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			runtime, ok := e.Object.(*corev1.ServiceAccount)
			if !ok {
				return false
			}
			return r.svc.IsBase(runtime)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
}

// Reconcile reads that state of the cluster for a ServiceAccount object and makes changes based
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

func (r *ServiceAccountReconciler) Reconcile(request ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	instance := &corev1.ServiceAccount{}
	if err := r.client.Get(ctx, request.NamespacedName, instance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger := r.Log.WithValues("namespace", instance.GetNamespace(), "name", instance.GetName())

	namespaces, err := getNamespaces(ctx, r.client, r.config.BaseNamespace, r.config.ExcludedNamespaces)
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, namespace := range namespaces {
		if err = r.svc.UpdateNamespace(ctx, logger, namespace, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: r.config.ServiceAccountRequeueDuration}, nil
}
