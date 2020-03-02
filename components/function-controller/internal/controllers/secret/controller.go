package secret

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/function-controller/internal/container"
	resource_watcher "github.com/kyma-project/kyma/components/function-controller/internal/resource-watcher"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SecretReconciler reconciles a Secret object
type SecretReconciler struct {
	client.Client
	log logr.Logger

	services *resource_watcher.ResourceWatcherServices
}

func NewController(config resource_watcher.ResourceWatcherConfig, log logr.Logger, di *container.Container) *SecretReconciler {
	return &SecretReconciler{
		Client:   di.Manager.GetClient(),
		log:      log,
		services: di.ResourceWatcherServices,
	}
}

func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Complete(r)
}

// Reconcile performs the reconciling for a single request object that can be used to fetch the secret it represents from the cache
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;create;update;patch;delete
func (r *SecretReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	secret := &corev1.Secret{}
	if err := r.Client.Get(ctx, req.NamespacedName, secret); err != nil {
		if apiErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	secretLogger := r.log.WithValues("kind", configMap.Kind, "name", configMap.Name, "namespace", configMap.Namespace)
	handler := newHandler(secretLogger, r.services)
	err := handler.Do(ctx, configMap)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
