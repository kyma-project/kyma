package jetstream

import "github.com/pkg/errors"

var (
	errFailedToUpdateStatus     = errors.New("failed to update JetStream subscription status")
	errFailedToDeleteSub        = errors.New("failed to delete JetStream subscription")
	errFailedToUpdateFinalizers = errors.New("failed to update subscription's finalizers")
)
