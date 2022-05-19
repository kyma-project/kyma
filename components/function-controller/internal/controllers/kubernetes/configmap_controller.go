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

type ConfigMapReconciler struct {
	Log    *zap.SugaredLogger
	client client.Client
	config Config
	svc    ConfigMapService
}

func NewConfigMap(client client.Client, log *zap.SugaredLogger, config Config, service ConfigMapService) *ConfigMapReconciler {
	return &ConfigMapReconciler{
		client: client,
		Log:    log.Named("controllers").Named("configmap"),
		config: config,
		svc:    service,
	}
}

func (r *ConfigMapReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("configmap-controller").
		For(&corev1.ConfigMap{}).
		WithEventFilter(r.predicate()).
		Complete(r)
}

func (r *ConfigMapReconciler) predicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			runtime, ok := e.Object.(*corev1.ConfigMap)
			if !ok {
				return false
			}
			return r.svc.IsBase(runtime)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			runtime, ok := e.ObjectNew.(*corev1.ConfigMap)
			if !ok {
				return false
			}
			return r.svc.IsBase(runtime)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			runtime, ok := e.Object.(*corev1.ConfigMap)
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

// Reconcile reads that state of the cluster for a ConfigMap object and makes changes based
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

func (r *ConfigMapReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	instance := &corev1.ConfigMap{}
	if err := r.client.Get(ctx, request.NamespacedName, instance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	r.client.Status()

	logger := r.Log.With("namespace", instance.GetNamespace(), "name", instance.GetName())

	namespaces, err := getNamespaces(ctx, r.client, r.config.BaseNamespace, r.config.ExcludedNamespaces)
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, namespace := range namespaces {
		if err = r.svc.UpdateNamespace(ctx, logger, namespace, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: r.config.ConfigMapRequeueDuration}, nil
}
