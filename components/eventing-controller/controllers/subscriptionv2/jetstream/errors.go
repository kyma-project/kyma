package jetstream

import "github.com/pkg/errors"

var (
	errFailedToUpdateStatus    = errors.New("failed to update JetStream subscription status")
	errFailedToDeleteSub       = errors.New("failed to delete JetStream subscription")
	errFailedToRemoveFinalizer = errors.New("failed to remove finalizer from subscription")
	errFailedToAddFinalizer    = errors.New("failed to add finalizer to subscription")
)
