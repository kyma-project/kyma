package namespace

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/function-controller/internal/container"
	resource_watcher "github.com/kyma-project/kyma/components/function-controller/internal/resource-watcher"
	"time"

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

	relistInterval time.Duration
	services *resource_watcher.ResourceWatcherServices
}

func NewController(config resource_watcher.ResourceWatcherConfig, log logr.Logger, di *container.Container) *NamespaceReconciler {
	return &NamespaceReconciler{
		Client: di.Manager.GetClient(),
		log:    log,
		relistInterval: config.NamespaceRelistInterval,
		services: di.ResourceWatcherServices,
	}
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}

// Reconcile performs the reconciling for a single request object that can be used to fetch the namespace it represents from the cache
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces/status,verbs=get;update;patch;watch
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

	if r.services.Namespaces.IsExcludedNamespace(namespace.Name) {
		r.log.Info(fmt.Sprintf("%s is a excluded/system namespace. Skipping...", namespace.Name))
		return ctrl.Result{}, nil
	}

	namespaceLogger := r.log.WithValues("kind", namespace.Kind, "name", namespace.Name)
	handler := newHandler(namespaceLogger, r.services)
	err := handler.Do(ctx, namespace)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{
		RequeueAfter: r.relistInterval,
	}, nil
}
