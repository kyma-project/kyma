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
	GetDeletionTimestamp() *v1.Time
	GetObjectKind() schema.ObjectKind
	DeepCopyObject() runtime.Object
}

type handler struct {
	log logr.Logger

	resourceType ResourceType
	services     *resource_watcher.ResourceWatcherServices
}

func newHandler(log logr.Logger, resourceType ResourceType, resourceWatcherServices *resource_watcher.ResourceWatcherServices) *handler {
	return &handler{
		log:          log,
		resourceType: resourceType,
		services:     resourceWatcherServices,
	}
}

func (h *handler) Do(ctx context.Context, obj MetaAccessor) error {
	h.logInfof("Start %s handling", h.resourceType)
	defer h.logInfof("Finish %s handling", h.resourceType)

	switch {
	case h.isOnDelete(obj):
		h.logInfof("On delete")
		return h.onDelete(ctx, obj)
	case h.isOnCreate(obj):
		h.logInfof("On create")
		return h.onDelete(ctx, obj)
	case h.isOnUpdate(obj):
		h.logInfof("On update")
		return h.onUpdate(ctx, obj)
	default:
		h.logInfof("Action not taken")
		return nil
	}
}

func (*handler) isOnCreate(obj MetaAccessor) bool {
	return obj.GetCreationTimestamp().IsZero() || obj.GetCreationTimestamp() == v1.Now()
}

func (*handler) isOnUpdate(obj MetaAccessor) bool {
	return !obj.GetCreationTimestamp().IsZero() && obj.GetCreationTimestamp() != v1.Now()
}

func (*handler) isOnDelete(obj MetaAccessor) bool {
	return !obj.GetDeletionTimestamp().IsZero()
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
	h.logInfof("Applying Registry Credentials")
	err := h.services.Credentials.ApplyCredentialsToNamespace(namespace.Name)
	if err != nil {
		return errors.Wrapf(err, "while applying Credentials in %s namespace", namespace.Name)
	}
	h.logInfof("Registry Credentials applied")

	h.logInfof("Applying Runtimes")
	err = h.services.Runtimes.ApplyRuntimesToNamespace(namespace.Name)
	if err != nil {
		return errors.Wrapf(err, "while applying Runtimes in %s namespace", namespace.Name)
	}
	h.logInfof("Runtimes applied")

	return nil
}

func (h *handler) onUpdate(ctx context.Context, obj MetaAccessor) error {
	switch object := obj.(type) {
	case *corev1.ConfigMap:
		return h.onUpdateConfigMap(ctx, object)
	case *corev1.Secret:
		return h.onUpdateSecret(ctx, object)
	default:
		return nil
	}
}

func (h *handler) onUpdateConfigMap(_ context.Context, configMap *corev1.ConfigMap) error {
	hasBaseNamespace := h.services.Namespaces.HasBaseNamespace(configMap.Namespace)

	return nil
}

func (h *handler) onUpdateSecret(_ context.Context, secret *corev1.Secret) error {
	hasBaseNamespace := h.services.Namespaces.HasBaseNamespace(secret.Namespace)

	return nil
}

func (h *handler) onDelete(ctx context.Context, obj MetaAccessor) error {
	switch object := obj.(type) {
	case *corev1.ConfigMap:
		return h.onDeleteConfigMap(ctx, object)
	case *corev1.Secret:
		return h.onDeleteSecret(ctx, object)
	default:
		return nil
	}
}

func (h *handler) onDeleteConfigMap(_ context.Context, configMap *corev1.ConfigMap) error {
	copiedCM := &corev1.ConfigMap{}
	configMap.DeepCopyInto(copiedCM)

	return nil
}

func (h *handler) onDeleteSecret(_ context.Context, secret *corev1.Secret) error {
	copiedS := &corev1.Secret{}
	secret.DeepCopyInto(copiedS)

	return nil
}

func (h *handler) logInfof(message string, args ...interface{}) {
	h.log.Info(fmt.Sprintf(message, args...))
}
