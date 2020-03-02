package config

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/function-controller/internal/container"
	resource_watcher "github.com/kyma-project/kyma/components/function-controller/internal/resource-watcher"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceType string

const (
	NamespaceType ResourceType = "namespace"
	ConfigMapType ResourceType = "configmap"
	SecretType ResourceType = "secret"
)

// ConfigReconciler reconciles a Namespace/ConfigMap/Secret object
type ConfigReconciler struct {
	client.Client
	log logr.Logger

	resourceType ResourceType
	services *resource_watcher.ResourceWatcherServices
}

func NewController(log logr.Logger, resourceType ResourceType, di *container.Container) *ConfigReconciler {
	return &ConfigReconciler{
		Client: di.Manager.GetClient(),
		resourceType: resourceType,
		log:    log,
		services: di.ResourceWatcherServices,
	}
}

var resourceMap = map[ResourceType]runtime.Object{
	NamespaceType: &corev1.Namespace{},
	ConfigMapType: &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Labels:
		},
	},
	SecretType:    &corev1.Secret{},
}

func (r *ConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For( [r.resourceType]).
		Complete(r)
}

// Reconcile performs the reconciling for a single request object that can be used to fetch the configMap it represents from the cache
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces/status,verbs=get;update;patch;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;create;update;patch;delete
func (r *ConfigReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	obj := *new(MetaAccessor)
	if err := r.Client.Get(ctx, req.NamespacedName, obj); err != nil {
		if apiErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if r.skipResource(obj) {
		r.log.Info(fmt.Sprintf("%v is not a appropriate object. Skipping...", obj))
		return ctrl.Result{}, nil
	}

	logger := r.log.WithValues("kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName(), "namespace", obj.GetNamespace())
	err := newHandler(logger, r.resourceType, r.services).Do(ctx, obj)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ConfigReconciler) skipResource(obj interface{}) bool {
	switch object := obj.(type) {
	case *corev1.Namespace:
		return r.services.Namespaces.IsExcludedNamespace(object.Name)
	case *corev1.ConfigMap:
		return r.services.Namespaces.IsExcludedNamespace(object.Namespace) && !r.services.Runtimes.IsRuntime(object)
	case *corev1.Secret:
		return r.services.Namespaces.IsExcludedNamespace(object.Namespace) && !r.services.Credentials.IsCredentials(object)
	default:
		return true
	}
}