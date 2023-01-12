package jetstream

import "github.com/pkg/errors"

var (
	errFailedToUpdateStatus = errors.New("failed to update JetStream subscription status")
	errFailedToDeleteSub    = errors.New("failed to delete JetStream subscription")
	errFailedToAddFinalizer = errors.New("failed to add finalizer to subscription")
)
