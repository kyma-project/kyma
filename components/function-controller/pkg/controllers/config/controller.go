package config

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	default:
		return r.reconcileSecret(req)
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
		r.log.Info(fmt.Sprintf("%s is not a appropriate object. Skipping...", namespace.Name))
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

func (r *Reconciler) reconcileSecret(req ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	secret := &corev1.Secret{}
	if err := r.Client.Get(ctx, req.NamespacedName, secret); err != nil {
		if apiErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	logger := r.log.WithValues("kind", secret.Kind, "namespace", secret.Namespace, "name", secret.Name)
	err := newHandler(logger, r.resourceType, r.services).Do(ctx, secret)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) getResource() runtime.Object {
	switch r.resourceType {
	case ConfigMapType:
		return &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: r.config.BaseNamespace,
				Labels: map[string]string{
					resource_watcher.ConfigLabel: resource_watcher.RuntimeLabelValue,
				},
			},
		}
	case SecretType:
		return &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: r.config.BaseNamespace,
				Labels: map[string]string{
					resource_watcher.ConfigLabel: resource_watcher.RegistryCredentialsLabelValue,
				},
			},
		}
	default:
		return &corev1.Namespace{}
	}
}
