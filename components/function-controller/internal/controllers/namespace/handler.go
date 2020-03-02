package namespace

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	resource_watcher "github.com/kyma-project/kyma/components/function-controller/internal/resource-watcher"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

type handler struct {
	log               logr.Logger
	services *resource_watcher.ResourceWatcherServices
}

func newHandler(log logr.Logger, resourceWatcherServices *resource_watcher.ResourceWatcherServices) *handler {
	return &handler{
		log:               log,
		services: resourceWatcherServices,
	}
}

func (h *handler) Do(ctx context.Context, namespace *corev1.Namespace) error {
	h.logInfof("Start Namespace handling")
	defer h.logInfof("Finish Namespace handling")

	switch {
	case h.isOnDelete(namespace):
		h.logInfof("On delete")
		return nil
	default:
		h.logInfof("On add or update")
		return h.onAddOrUpdate(ctx, namespace)
	}
}

func (*handler) isOnDelete(namespace *corev1.Namespace) bool {
	return namespace.DeletionTimestamp != nil || !namespace.DeletionTimestamp.IsZero()
}

func (h *handler) onAddOrUpdate(ctx context.Context, namespace *corev1.Namespace) error {
	var err error

	h.logInfof("Applying Registry Credentials")
	err = h.services.Credentials.ApplyCredentialsToNamespace(namespace.Name)
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

func (h *handler) logInfof(message string, args ...interface{}) {
	h.log.Info(fmt.Sprintf(message, args...))
}
