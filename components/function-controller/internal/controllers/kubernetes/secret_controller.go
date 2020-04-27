package kubernetes

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type SecretReconciler struct {
	Log    logr.Logger
	client client.Client
	config Config
	svc    SecretService
}

func NewSecret(client client.Client, log logr.Logger, config Config, secretSvc SecretService) *SecretReconciler {
	return &SecretReconciler{
		client: client,
		Log:    log.WithName("controllers").WithName("secret"),
		config: config,
		svc:    secretSvc,
	}
}

func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("secret-controller").
		For(&corev1.Secret{}).
		WithEventFilter(r.predicate()).
		Complete(r)
}

func (r *SecretReconciler) predicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			runtime, ok := e.Object.(*corev1.Secret)
			if !ok {
				return false
			}
			return r.svc.IsBase(runtime)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			runtime, ok := e.ObjectNew.(*corev1.Secret)
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

// Reconcile reads that state of the cluster for a Secret object and makes changes based
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch

func (r *SecretReconciler) Reconcile(request ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	instance := &corev1.Secret{}
	if err := r.client.Get(ctx, request.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
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

	return ctrl.Result{RequeueAfter: r.config.SecretRequeueDuration}, nil
}
