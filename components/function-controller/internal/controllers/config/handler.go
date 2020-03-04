package config

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	resource_watcher "github.com/kyma-project/kyma/components/function-controller/internal/resource-watcher"
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

	switch {
	case h.isOnCreate(obj):
		h.logInfof("On create")
		return h.onCreate(ctx, obj)
	case h.isOnUpdate(obj):
		h.logInfof("On update")
		return h.onUpdate(ctx, obj)
	default:
		h.logInfof("Action not taken")
		return nil
	}
}

func (*handler) isOnCreate(obj MetaAccessor) bool {
	creationTimestamp := obj.GetCreationTimestamp()
	return (&creationTimestamp).IsZero() || creationTimestamp == v1.Now()
}

func (*handler) isOnUpdate(obj MetaAccessor) bool {
	creationTimestamp := obj.GetCreationTimestamp()
	return !(&creationTimestamp).IsZero() && creationTimestamp != v1.Now()
}

func (h *handler) onCreate(ctx context.Context, obj interface{}) error {
	switch object := obj.(type) {
	case *corev1.Namespace:
		return h.onCreateNamespace(ctx, object)
	default:
		return nil
	}
}

func (h *handler) onCreateNamespace(_ context.Context, namespace *corev1.Namespace) error {
	namespaceName := namespace.Name

	h.logInfof("Applying Registry Credentials")
	err := h.services.Credentials.CreateCredentialsInNamespace(namespaceName)
	if err != nil {
		return errors.Wrapf(err, "while applying Credentials in %s namespace", namespaceName)
	}
	h.logInfof("Registry Credentials applied")

	h.logInfof("Applying Runtimes")
	err = h.services.Runtimes.CreateRuntimesInNamespace(namespaceName)
	if err != nil {
		return errors.Wrapf(err, "while applying Runtimes in %s namespace", namespaceName)
	}
	h.logInfof("Runtimes applied")

	return nil
}

func (h *handler) onUpdate(ctx context.Context, obj MetaAccessor) error {
	switch object := obj.(type) {
	case *corev1.Namespace:
		return h.onUpdateNamespace(ctx, object)
	case *corev1.ConfigMap:
		return h.onUpdateConfigMap(ctx, object)
	case *corev1.Secret:
		return h.onUpdateSecret(ctx, object)
	default:
		return nil
	}
}

func (h *handler) onUpdateNamespace(_ context.Context, namespace *corev1.Namespace) error {
	namespaceName := namespace.Name

	err := h.services.Credentials.UpdateCredentialsInNamespace(namespaceName)
	if err != nil {
		return errors.Wrapf(err, "while reconciling namespace '%s' - update Registry Credentials", namespaceName)
	}

	err = h.services.Runtimes.UpdateRuntimesInNamespace(namespaceName)
	if err != nil {
		return errors.Wrapf(err, "while reconciling namespace '%s' - update Runtimes", namespaceName)
	}

	return nil
}

func (h *handler) onUpdateConfigMap(_ context.Context, configMap *corev1.ConfigMap) error {
	err := h.services.Runtimes.UpdateCachedRuntime(configMap)
	if err != nil {
		return errors.Wrapf(err, "while propagating new Runtime %v to namespaces", configMap)
	}

	namespaces, err := h.services.Namespaces.GetNamespaces()
	if err != nil {
		return errors.Wrapf(err, "while propagating new Runtime %v to namespaces", configMap)
	}

	err = h.services.Runtimes.UpdateRuntimeInNamespaces(configMap, namespaces)
	if err != nil {
		return errors.Wrapf(err, "while propagating new Runtime %v to namespaces", configMap)
	}

	return nil
}

func (h *handler) onUpdateSecret(_ context.Context, secret *corev1.Secret) error {
	h.services.Credentials.UpdateCachedCredentials(secret)

	namespaces, err := h.services.Namespaces.GetNamespaces()
	if err != nil {
		return errors.Wrapf(err, "while propagating new Registry Credentials %v to namespaces", secret)
	}

	err = h.services.Credentials.UpdateCredentialsInNamespaces(namespaces)
	if err != nil {
		return errors.Wrapf(err, "while propagating new Registry Credentials %v to namespaces", secret)
	}

	return nil
}

func (h *handler) logInfof(message string, args ...interface{}) {
	h.log.Info(fmt.Sprintf(message, args...))
}
