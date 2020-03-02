package configmap

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/function-controller/internal/container"
	resource_watcher "github.com/kyma-project/kyma/components/function-controller/internal/resource-watcher"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConfigMapReconciler reconciles a ConfigMap object
type ConfigMapReconciler struct {
	client.Client
	log logr.Logger

	services *resource_watcher.ResourceWatcherServices
}

func NewController(config resource_watcher.ResourceWatcherConfig, log logr.Logger, di *container.Container) *ConfigMapReconciler {
	return &ConfigMapReconciler{
		Client: di.Manager.GetClient(),
		log:    log,
		services: di.ResourceWatcherServices,
	}
}

func (r *ConfigMapReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Complete(r)
}

// Reconcile performs the reconciling for a single request object that can be used to fetch the configMap it represents from the cache
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;create;update;patch;delete
func (r *ConfigMapReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configMap := &corev1.ConfigMap{}
	if err := r.Client.Get(ctx, req.NamespacedName, configMap); err != nil {
		if apiErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if !r.services.Runtimes.IsRuntime(configMap) {
		r.log.Info(fmt.Sprintf("%s in %s is not a Runtime ConfigMap. Skipping...", configMap.Name, configMap.Namespace))
		return ctrl.Result{}, nil
	}

	configMapLogger := r.log.WithValues("kind", configMap.Kind, "name", configMap.Name, "namespace", configMap.Namespace)
	handler := newHandler(configMapLogger, r.services)
	err := handler.Do(ctx, configMap)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}