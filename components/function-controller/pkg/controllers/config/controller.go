package config

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/function-controller/pkg/configwatcher"
	"github.com/kyma-project/kyma/components/function-controller/pkg/container"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceType string

const (
	NamespaceType      ResourceType = "Namespace"
	ConfigMapType      ResourceType = "Configmap"
	SecretType         ResourceType = "Secret"
	ServiceAccountType ResourceType = "ServiceAccount"
)

// Reconciler reconciles a Namespace/ConfigMap/Secret object
type Reconciler struct {
	client.Client
	Log logr.Logger

	config       configwatcher.Config
	resourceType ResourceType
	services     *configwatcher.Services
}

func NewController(config configwatcher.Config, resourceType ResourceType, log logr.Logger, di *container.Container) *Reconciler {
	return &Reconciler{
		Client:       di.Manager.GetClient(),
		Log:          log,
		config:       config,
		resourceType: resourceType,
		services:     di.ResourceConfigServices,
	}
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(r.getResource()).
		WithEventFilter(r.getPredicates()).
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
	case SecretType:
		return r.reconcileCredentials(req)
	case ServiceAccountType:
		return r.reconcileServiceAccount(req)
	default:
		return r.reconcileNamespace(req)
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

	logger := r.Log.WithValues("kind", namespace.Kind, "name", namespace.Name)
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

	logger := r.Log.WithValues("kind", runtime.Kind, "namespace", runtime.Namespace, "name", runtime.Name)
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

	logger := r.Log.WithValues("kind", credentials.Kind, "namespace", credentials.Namespace, "name", credentials.Name)
	err := newHandler(logger, r.resourceType, r.services).Do(ctx, credentials)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileServiceAccount(req ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serviceAccount := &corev1.ServiceAccount{}
	if err := r.Client.Get(ctx, req.NamespacedName, serviceAccount); err != nil {
		if apiErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	logger := r.Log.WithValues("kind", serviceAccount.Kind, "namespace", serviceAccount.Namespace, "name", serviceAccount.Name)
	err := newHandler(logger, r.resourceType, r.services).Do(ctx, serviceAccount)
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
	case SecretType:
		return &corev1.Secret{}
	case ServiceAccountType:
		return &corev1.ServiceAccount{}
	default:
		return &corev1.Namespace{}
	}
}

func (r *Reconciler) getPredicates() predicate.Predicate {
	switch r.resourceType {
	case NamespaceType:
		return r.predicatesForNamespace()
	case ConfigMapType:
		return r.predicatesForRuntime()
	case SecretType:
		return r.predicatesForCredential()
	case ServiceAccountType:
		return r.predicatesForServiceAccount()
	default:
		return r.predicatesForNamespace()
	}
}

func (r *Reconciler) predicatesForNamespace() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			namespace, ok := e.Object.(*corev1.Namespace)
			if !ok {
				return false
			}
			return !r.services.Namespaces.IsExcludedNamespace(namespace.Name)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
}

func (r *Reconciler) predicatesForRuntime() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			runtime, ok := e.Object.(*corev1.ConfigMap)
			if !ok {
				return false
			}
			return r.services.Runtimes.IsBaseRuntime(runtime)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			runtime, ok := e.ObjectNew.(*corev1.ConfigMap)
			if !ok {
				return false
			}
			return r.services.Runtimes.IsBaseRuntime(runtime)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
}

func (r *Reconciler) predicatesForCredential() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			credential, ok := e.Object.(*corev1.Secret)
			if !ok {
				return false
			}
			return r.services.Credentials.IsBaseCredential(credential)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			credential, ok := e.ObjectNew.(*corev1.Secret)
			if !ok {
				return false
			}
			return r.services.Credentials.IsBaseCredential(credential)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
}

func (r *Reconciler) predicatesForServiceAccount() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			serviceAccount, ok := e.ObjectNew.(*corev1.ServiceAccount)
			if !ok {
				return false
			}
			return r.services.ServiceAccount.IsBaseServiceAccount(serviceAccount)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
}
