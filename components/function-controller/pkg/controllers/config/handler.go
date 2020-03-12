package config

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/function-controller/pkg/configwatcher"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type MetaAccessor interface {
	GetNamespace() string
	GetName() string
	GetObjectKind() schema.ObjectKind
}

type handler struct {
	log logr.Logger

	resourceType ResourceType
	services     *configwatcher.Services
}

func newHandler(log logr.Logger, resourceType ResourceType, services *configwatcher.Services) *handler {
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
		return h.handleNamespace(ctx, object)
	case *corev1.ConfigMap:
		return h.handleRuntime(ctx, object)
	case *corev1.Secret:
		return h.handleCredential(ctx, object)
	default:
		return nil
	}
}

func (h *handler) handleNamespace(_ context.Context, namespace *corev1.Namespace) error {
	namespaceName := namespace.Name

	h.logInfof("Applying Credentials")
	err := h.services.Credentials.UpdateCredentialsInNamespace(namespaceName)
	if err != nil {
		return errors.Wrapf(err, "while applying Credentials in %s namespace", namespaceName)
	}
	h.logInfof("Credentials applied")

	h.logInfof("Applying Service Account")
	err = h.services.ServiceAccount.CreateServiceAccountInNamespace(namespaceName)
	if err != nil {
		return errors.Wrapf(err, "while applying Service Account in %s namespace", namespaceName)
	}
	h.logInfof("Service Account applied")

	h.logInfof("Applying Runtimes")
	err = h.services.Runtimes.UpdateRuntimesInNamespace(namespaceName)
	if err != nil {
		return errors.Wrapf(err, "while applying Runtimes in %s namespace", namespaceName)
	}
	h.logInfof("Runtimes applied")

	return nil
}

func (h *handler) handleRuntime(_ context.Context, runtime *corev1.ConfigMap) error {
	h.logInfof("Updating Runtime in namespaces")

	runtimeName := runtime.Name
	err := h.services.Runtimes.UpdateCachedRuntime(runtime)
	if err != nil {
		return errors.Wrapf(err, "while propagating new Runtime %s to namespaces", runtimeName)
	}

	namespaces, err := h.services.Namespaces.GetNamespaces()
	if err != nil {
		return errors.Wrapf(err, "while propagating new Runtime %s to namespaces", runtimeName)
	}

	err = h.services.Runtimes.UpdateRuntimeInNamespaces(runtime, namespaces)
	if err != nil {
		return errors.Wrapf(err, "while propagating new Runtime %s to namespaces", runtimeName)
	}

	h.logInfof("Runtime updated in namespaces")
	return nil
}

func (h *handler) handleCredential(_ context.Context, credential *corev1.Secret) error {
	h.logInfof("Updating Credential in namespaces")

	credentialName := credential.Name
	err := h.services.Credentials.UpdateCachedCredential(credential)
	if err != nil {
		return errors.Wrapf(err, "while propagating new Credential %s to namespaces", credentialName)
	}

	namespaces, err := h.services.Namespaces.GetNamespaces()
	if err != nil {
		return errors.Wrapf(err, "while propagating new Credential %s to namespaces", credentialName)
	}

	err = h.services.Credentials.UpdateCredentialInNamespaces(credential, namespaces)
	if err != nil {
		return errors.Wrapf(err, "while propagating new Credential %s to namespaces", credentialName)
	}

	h.logInfof("Credential updated in namespaces")
	return nil
}

func (h *handler) logInfof(message string, args ...interface{}) {
	h.log.Info(fmt.Sprintf(message, args...))
}
