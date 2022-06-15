package kubernetes

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type NamespaceReconciler struct {
	Log                *zap.SugaredLogger
	client             client.Client
	config             Config
	configMapSvc       ConfigMapService
	secretSvc          SecretService
	serviceAccountSvc  ServiceAccountService
	roleService        RoleService
	roleBindingService RoleBindingService
}

func NewNamespace(client client.Client, log *zap.SugaredLogger, config Config,
	configMapSvc ConfigMapService, secretSvc SecretService, serviceAccountSvc ServiceAccountService,
	roleService RoleService, roleBindingService RoleBindingService) *NamespaceReconciler {
	return &NamespaceReconciler{
		client:             client,
		Log:                log.Named("controllers").Named("namespace"),
		config:             config,
		configMapSvc:       configMapSvc,
		secretSvc:          secretSvc,
		serviceAccountSvc:  serviceAccountSvc,
		roleService:        roleService,
		roleBindingService: roleBindingService,
	}
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("namespace-controller").
		For(&corev1.Namespace{}).
		WithEventFilter(r.predicate()).
		Complete(r)
}

func (r *NamespaceReconciler) predicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			namespace, ok := e.Object.(*corev1.Namespace)
			if !ok {
				return false
			}
			return !isExcludedNamespace(namespace.Name, r.config.BaseNamespace, r.config.ExcludedNamespaces)
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
}

// Reconcile reads that state of the cluster for a Namespace object and updates other resources based on it
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=configmaps;secrets;serviceaccounts,verbs=get;list;watch;create;update;patch;delete

func (r *NamespaceReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	instance := &corev1.Namespace{}
	if err := r.client.Get(ctx, request.NamespacedName, instance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger := r.Log.With("name", instance.GetName())

	logger.Info(fmt.Sprintf("Updating ConfigMaps in namespace '%s'", instance.GetName()))
	configMaps, err := r.configMapSvc.ListBase(ctx)
	if err != nil {
		logger.Error(err, "Listing base ConfigMaps failed")
		return ctrl.Result{}, err
	}
	for _, configMap := range configMaps {
		c := configMap
		if err := r.configMapSvc.UpdateNamespace(ctx, logger, instance.GetName(), &c); err != nil {
			return ctrl.Result{}, err
		}
	}

	logger.Info(fmt.Sprintf("Updating Secrets in namespace '%s'", instance.GetName()))
	secrets, err := r.secretSvc.ListBase(ctx)
	if err != nil {
		logger.Error(err, "Listing base Secrets failed")
		return ctrl.Result{}, err
	}
	for _, secret := range secrets {
		s := secret
		if err := r.secretSvc.UpdateNamespace(ctx, logger, instance.GetName(), &s); err != nil {
			return ctrl.Result{}, err
		}
	}

	logger.Info(fmt.Sprintf("Updating Roles in namespace '%s'", instance.GetName()))
	roles, err := r.roleService.ListBase(ctx)
	if err != nil {
		logger.Error(err, "Listing base Roles failed")
		return ctrl.Result{}, err
	}
	for _, role := range roles {
		rr := role
		if err := r.roleService.UpdateNamespace(ctx, logger, instance.GetName(), &rr); err != nil {
			return ctrl.Result{}, err
		}
	}

	logger.Info(fmt.Sprintf("Updating RoleBindings in namespace '%s'", instance.GetName()))
	roleBindings, err := r.roleBindingService.ListBase(ctx)
	if err != nil {
		logger.Error(err, "Listing base RoleBindings failed")
		return ctrl.Result{}, err
	}
	for _, roleBinding := range roleBindings {
		rb := roleBinding
		if err := r.roleBindingService.UpdateNamespace(ctx, logger, instance.GetName(), &rb); err != nil {
			return ctrl.Result{}, err
		}
	}

	logger.Info(fmt.Sprintf("Updating ServiceAccounts in namespace '%s'", instance.GetName()))
	serviceAccounts, err := r.serviceAccountSvc.ListBase(ctx)
	if err != nil {
		logger.Error(err, "Listing base ServiceAccounts failed")
		return ctrl.Result{}, err
	}
	for _, serviceAccount := range serviceAccounts {
		sa := serviceAccount
		if err := r.serviceAccountSvc.UpdateNamespace(ctx, logger, instance.GetName(), &sa); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
