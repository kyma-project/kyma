package controllers

import (
	"context"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	client.Client
	log logr.Logger

	config NamespaceConfig
}

type NamespaceConfig struct {
	EnableController      bool     `envconfig:"default=true"`
	BaseNamespace         string   `envconfig:"default=kyma-system"`
	SecretName            string   `envconfig:"default=1"`
	RuntimeConfigMapNames []string `envconfig:"default=1"`
	ExcludedNamespaces    []string `envconfig:"default=kube-system,kyma-system"`
}

func NewNamespace(config NamespaceConfig, log logr.Logger, di *Container) *NamespaceReconciler {
	return &NamespaceReconciler{
		Client: di.Manager.GetClient(),
		log:    log,
		config: config,
	}
}

// Reconcile performs the reconciling for a single request object that can be used to fetch the namespace it represents from the cache
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces/status,verbs=get;update;patch;watch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=create;update;patch;delete
func (r *NamespaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	namespace := &corev1.Namespace{}
	if err := r.Client.Get(ctx, req.NamespacedName, namespace); err != nil {
		if apiErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}
	if !namespace.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	namespaceName := namespace.Name
	if r.isExcludedNamespace(namespaceName) {
		r.log.Info(fmt.Sprintf("%s is a system namespace. Skipping...", namespaceName))
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}

func (r *NamespaceReconciler) isExcludedNamespace(namespaceName string) bool {
	for _, name := range r.config.ExcludedNamespaces {
		if name == namespaceName {
			return true
		}
	}
	return false
}
