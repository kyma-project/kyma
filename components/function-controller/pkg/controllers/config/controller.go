package config

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/function-controller/pkg/container"
	resource_watcher "github.com/kyma-project/kyma/components/function-controller/pkg/resource-watcher"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceType string

const (
	NamespaceType ResourceType = "Namespace"
	ConfigMapType ResourceType = "Configmap"
	SecretType    ResourceType = "Secret"
)

// Reconciler reconciles a Namespace/ConfigMap/Secret object
type Reconciler struct {
	client.Client
	log logr.Logger

	config       resource_watcher.Config
	resourceType ResourceType
	services     *resource_watcher.Services
}

func NewController(config resource_watcher.Config, resourceType ResourceType, log logr.Logger, di *container.Container) *Reconciler {
	return &Reconciler{
		Client:       di.Manager.GetClient(),
		log:          log,
		config:       config,
		resourceType: resourceType,
		services:     di.ResourceWatcherServices,
	}
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(r.getResource()).
		Complete(r)
}

// Reconcile performs the reconciling for a single request object that can be used to fetch the configMap it represents from the cache
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces/status,verbs=get;update;patch;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;create;update;patch;delete
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	switch r.resourceType {
	case NamespaceType:
		return r.reconcileNamespace(req)
	case ConfigMapType:
		return r.reconcileRuntimes(req)
	default:
		return r.reconcileCredentials(req)
	}
}

func (r *Reconciler) reconcileNamespace(req ctrl.Request) (ctrl.Result, error) {
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
		r.log.Info(fmt.Sprintf("%s is not a excluded namespace. Skipping...", namespace.Name))
		return ctrl.Result{}, nil
	}

	logger := r.log.WithValues("kind", namespace.Kind, "name", namespace.Name)
	err := newHandler(logger, r.resourceType, r.services).Do(ctx, namespace)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{
		RequeueAfter: r.config.NamespaceRelistInterval,
	}, nil
}

func (r *Reconciler) reconcileRuntimes(req ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runtime := &corev1.ConfigMap{}
	if err := r.Client.Get(ctx, req.NamespacedName, runtime); err != nil {
		if apiErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if !r.services.Runtimes.IsBaseRuntime(runtime) {
		r.log.Info(fmt.Sprintf("%s in %s namespace is not a base runtime. Skipping...", runtime.Name, runtime.Namespace))
		return ctrl.Result{}, nil
	}

	logger := r.log.WithValues("kind", runtime.Kind, "namespace", runtime.Namespace, "name", runtime.Name)
	err := newHandler(logger, r.resourceType, r.services).Do(ctx, runtime)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileCredentials(req ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	credentials := &corev1.Secret{}
	if err := r.Client.Get(ctx, req.NamespacedName, credentials); err != nil {
		if apiErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if !r.services.Credentials.IsBaseCredentials(credentials) {
		r.log.Info(fmt.Sprintf("%s in %s namespace is not a base credentials. Skipping...", credentials.Name, credentials.Namespace))
		return ctrl.Result{}, nil
	}

	logger := r.log.WithValues("kind", credentials.Kind, "namespace", credentials.Namespace, "name", credentials.Name)
	err := newHandler(logger, r.resourceType, r.services).Do(ctx, credentials)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) getResource() runtime.Object {
	switch r.resourceType {
	case NamespaceType:
		return &corev1.Namespace{}
	case ConfigMapType:
		return &corev1.ConfigMap{}
	default:
		return &corev1.Secret{}
	}
}
