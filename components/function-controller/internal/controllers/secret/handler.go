package secret

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	resource_watcher "github.com/kyma-project/kyma/components/function-controller/internal/resource-watcher"
	corev1 "k8s.io/api/core/v1"
)

type handler struct {
	log      logr.Logger
	services *resource_watcher.ResourceWatcherServices
}

func newHandler(log logr.Logger, resourceWatcherServices *resource_watcher.ResourceWatcherServices) *handler {
	return &handler{
		log:      log,
		services: resourceWatcherServices,
	}
}

func (h *handler) Do(ctx context.Context, secret *corev1.Secret) error {
	h.logInfof("Start Secret handling")
	defer h.logInfof("Finish Secret handling")

	switch {
	case h.isOnDelete(secret):
		h.logInfof("On delete")
		return nil
	default:
		h.logInfof("On add or update")
		return h.onAddOrUpdate(ctx, secret)
	}
}

func (*handler) isOnDelete(secret *corev1.Secret) bool {
	return secret.DeletionTimestamp != nil || !secret.DeletionTimestamp.IsZero()
}

func (h *handler) logInfof(message string, args ...interface{}) {
	h.log.Info(fmt.Sprintf(message, args...))
}
