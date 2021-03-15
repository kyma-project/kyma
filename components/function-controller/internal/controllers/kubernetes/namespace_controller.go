package kubernetes

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type NamespaceReconciler struct {
	Log                logr.Logger
	client             client.Client
	config             Config
	configMapSvc       ConfigMapService
	secretSvc          SecretService
	serviceAccountSvc  ServiceAccountService
	roleService        RoleService
	roleBindingService RoleBindingService
}

func NewNamespace(client client.Client, log logr.Logger, config Config,
	configMapSvc ConfigMapService, secretSvc SecretService, serviceAccountSvc ServiceAccountService,
	roleService RoleService, roleBindingService RoleBindingService) *NamespaceReconciler {
	return &NamespaceReconciler{
		client:             client,
		Log:                log.WithName("controllers").WithName("namespace"),
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

func (r *NamespaceReconciler) Reconcile(request ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	instance := &corev1.Namespace{}
	if err := r.client.Get(ctx, request.NamespacedName, instance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger := r.Log.WithValues("name", instance.GetName())

	logger.Info(fmt.Sprintf("Updating ConfigMaps in namespace '%s'", instance.GetName()))
	configMaps, err := r.configMapSvc.ListBase(ctx)
	if err != nil {
		logger.Error(err, "Listing base ConfigMaps failed")
		return ctrl.Result{}, err
	}
	for _, configMap := range configMaps {
		if err = r.configMapSvc.UpdateNamespace(ctx, logger, instance.GetName(), &configMap); err != nil {
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
		if err = r.secretSvc.UpdateNamespace(ctx, logger, instance.GetName(), &secret); err != nil {
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
		if err = r.roleService.UpdateNamespace(ctx, logger, instance.GetName(), &role); err != nil {
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
		if err = r.roleBindingService.UpdateNamespace(ctx, logger, instance.GetName(), &roleBinding); err != nil {
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
		if err = r.serviceAccountSvc.UpdateNamespace(ctx, logger, instance.GetName(), &serviceAccount); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
