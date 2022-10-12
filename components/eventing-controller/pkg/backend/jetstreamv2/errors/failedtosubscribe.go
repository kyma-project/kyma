package errors

import (
	"fmt"

	"golang.org/x/xerrors"
)

type ErrFailedToSubscribeOnNATS = FailedToSubscribeOnNATSError

type FailedToSubscribeOnNATSError struct {
	OriginalError error
}

func NewFailedToSubscribeOnNATSError(reason error) error {
	return xerrors.Errorf("%w", &FailedToSubscribeOnNATSError{OriginalError: reason})
}

func (e *FailedToSubscribeOnNATSError) Error() string {
	return fmt.Sprintf("failed to create NATS JetStream subscription for subject: %v", e.OriginalError)
}

func (e *FailedToSubscribeOnNATSError) Is(target error) bool {
	if _, ok := target.(*FailedToSubscribeOnNATSError); !ok {
		return false
	}
	return true
}
