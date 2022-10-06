package kubernetes

import (
	"context"

	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type SecretReconciler struct {
	Log    *zap.SugaredLogger
	client client.Client
	config Config
	svc    SecretService
}

func NewSecret(client client.Client, log *zap.SugaredLogger, config Config, secretSvc SecretService) *SecretReconciler {
	return &SecretReconciler{
		client: client,
		Log:    log,
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
		GenericFunc: func(e event.GenericEvent) bool {
			runtime, ok := e.Object.(*corev1.Secret)
			if !ok {
				return false
			}
			return r.svc.IsBase(runtime)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			runtime, ok := e.Object.(*corev1.Secret)
			if !ok {
				return false
			}
			return r.svc.IsBase(runtime)
		},
	}
}

// Reconcile reads that state of the cluster for a Secret object and makes changes based
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

func (r *SecretReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	instance := &corev1.Secret{}
	if err := r.client.Get(ctx, request.NamespacedName, instance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger := r.Log.With("namespace", instance.GetNamespace(), "name", instance.GetName())

	namespaces, err := getNamespaces(ctx, r.client, r.config.BaseNamespace, r.config.ExcludedNamespaces)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.svc.HandleFinalizer(ctx, logger, instance, namespaces); err != nil {
		return ctrl.Result{}, err
	}
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	for _, namespace := range namespaces {
		if err = r.svc.UpdateNamespace(ctx, logger, namespace, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: r.config.SecretRequeueDuration}, nil
}
