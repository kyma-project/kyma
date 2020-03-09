package config

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	resource_watcher "github.com/kyma-project/kyma/components/function-controller/pkg/resource-watcher"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type MetaAccessor interface {
	GetNamespace() string
	GetName() string
	GetCreationTimestamp() v1.Time
	GetDeletionTimestamp() *v1.Time
	GetObjectKind() schema.ObjectKind
	DeepCopyObject() runtime.Object
}

type handler struct {
	log logr.Logger

	resourceType ResourceType
	services     *resource_watcher.Services
}

func newHandler(log logr.Logger, resourceType ResourceType, services *resource_watcher.Services) *handler {
	return &handler{
		log:          log,
		resourceType: resourceType,
		services:     services,
	}
}

func (h *handler) Do(ctx context.Context, obj MetaAccessor) error {
	h.logInfof("Start %s handling", h.resourceType)
	defer h.logInfof("Finish %s handling", h.resourceType)

	switch object := obj.(type) {
	case *corev1.Namespace:
		return h.onCreateNamespace(ctx, object)
	case *corev1.ConfigMap:
		return h.onUpdateRuntime(ctx, object)
	case *corev1.Secret:
		return h.onUpdateCredential(ctx, object)
	default:
		return nil
	}

	//switch {
	//case h.isOnCreate(obj):
	//	h.logInfof("On create")
	//	return h.onCreate(ctx, obj)
	//case h.isOnUpdate(obj):
	//	h.logInfof("On update")
	//	return h.onUpdate(ctx, obj)
	//default:
	//	h.logInfof("Action not taken")
	//	return nil
	//}
}

//func (h *handler) onCreate(ctx context.Context, obj interface{}) error {
//	switch object := obj.(type) {
//	case *corev1.Namespace:
//		return h.onCreateNamespace(ctx, object)
//	default:
//		return nil
//	}
//}

func (h *handler) onCreateNamespace(_ context.Context, namespace *corev1.Namespace) error {
	namespaceName := namespace.Name

	h.logInfof("Applying Credentials in %s namespace", namespaceName)
	err := h.services.Credentials.CreateCredentialsInNamespace(namespaceName)
	if err != nil {
		return errors.Wrapf(err, "while applying Credentials in %s namespace", namespaceName)
	}
	h.logInfof("Credentials applied in %s namespace", namespaceName)

	h.logInfof("Applying Service Account in %s namespace", namespaceName)
	err = h.services.ServiceAccount.CreateServiceAccountInNamespace(namespaceName)
	if err != nil {
		return errors.Wrapf(err, "while applying Service Account in %s namespace", namespaceName)
	}
	h.logInfof("Service Account applied in %s namespace", namespaceName)

	h.logInfof("Applying Runtimes in %s namespace", namespaceName)
	err = h.services.Runtimes.CreateRuntimesInNamespace(namespaceName)
	if err != nil {
		return errors.Wrapf(err, "while applying Runtimes in %s namespace", namespaceName)
	}
	h.logInfof("Runtimes applied in %s namespace", namespaceName)

	return nil
}

//func (h *handler) onUpdate(ctx context.Context, obj MetaAccessor) error {
//	switch object := obj.(type) {
//	case *corev1.Namespace:
//		return h.onUpdateNamespace(ctx, object)
//	case *corev1.ConfigMap:
//		return h.onUpdateConfigMap(ctx, object)
//	case *corev1.Secret:
//		return h.onUpdateSecret(ctx, object)
//	//case *corev1.ServiceAccount:
//	//	return h.onUpdateServiceAccount(ctx, object)
//	default:
//		return nil
//	}
//}

func (h *handler) onUpdateNamespace(_ context.Context, namespace *corev1.Namespace) error {
	namespaceName := namespace.Name
	err := h.services.Credentials.UpdateCredentialsInNamespace(namespaceName)
	if err != nil {
		return errors.Wrapf(err, "while reconciling namespace '%s' - update Registry Credentials", namespaceName)
	}

	//err = h.services.ServiceAccount.UpdateServiceAccountInNamespace(namespaceName)
	//if err != nil {
	//	return errors.Wrapf(err, "while reconciling namespace '%s' - update Service Account", namespaceName)
	//}

	err = h.services.Runtimes.UpdateRuntimesInNamespace(namespaceName)
	if err != nil {
		return errors.Wrapf(err, "while reconciling namespace '%s' - update Runtimes", namespaceName)
	}

	return nil
}

func (h *handler) onUpdateRuntime(_ context.Context, configMap *corev1.ConfigMap) error {
	configMapName := configMap.Name
	err := h.services.Runtimes.UpdateCachedRuntime(configMap)
	if err != nil {
		return errors.Wrapf(err, "while propagating new Runtime %s to namespaces", configMapName)
	}

	namespaces, err := h.services.Namespaces.GetNamespaces()
	if err != nil {
		return errors.Wrapf(err, "while propagating new Runtime %s to namespaces", configMapName)
	}

	err = h.services.Runtimes.UpdateRuntimeInNamespaces(configMap, namespaces)
	if err != nil {
		return errors.Wrapf(err, "while propagating new Runtime %s to namespaces", configMapName)
	}

	return nil
}

func (h *handler) onUpdateCredential(_ context.Context, secret *corev1.Secret) error {
	secretName := secret.Name
	err := h.services.Credentials.UpdateCachedCredential(secret)
	if err != nil {
		return errors.Wrapf(err, "while propagating new Credential %s to namespaces", secretName)
	}

	namespaces, err := h.services.Namespaces.GetNamespaces()
	if err != nil {
		return errors.Wrapf(err, "while propagating new Credential %s to namespaces", secretName)
	}

	err = h.services.Credentials.UpdateCredentialsInNamespaces(namespaces)
	if err != nil {
		return errors.Wrapf(err, "while propagating new Credential %s to namespaces", secretName)
	}

	return nil
}

//func (h *handler) onUpdateServiceAccount(_ context.Context, serviceAccount *corev1.ServiceAccount) error {
//	err := h.services.ServiceAccount.UpdateCachedServiceAccount(serviceAccount)
//	if err != nil {
//		return errors.Wrapf(err, "while propagating new Service Account %v to namespaces", serviceAccount)
//	}
//
//	namespaces, err := h.services.Namespaces.GetNamespaces()
//	if err != nil {
//		return errors.Wrapf(err, "while propagating new Service Account %v to namespaces", serviceAccount)
//	}
//
//	err = h.services.ServiceAccount.UpdateServiceAccountInNamespaces(namespaces)
//	if err != nil {
//		return errors.Wrapf(err, "while propagating new Service Account %v to namespaces", serviceAccount)
//	}
//
//	return nil
//}

func (h *handler) logInfof(message string, args ...interface{}) {
	h.log.Info(fmt.Sprintf(message, args...))
}
