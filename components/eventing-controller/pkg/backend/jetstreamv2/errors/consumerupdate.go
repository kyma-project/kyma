package errors

import (
	"fmt"
)

type FailedToUpdateConsumerInfoError struct {
	OriginalError error
}

func NewFailedToUpdateConsumerInfoError(reason error) error {
	return fmt.Errorf("%w", &FailedToUpdateConsumerInfoError{OriginalError: reason})
}

func (e *FailedToUpdateConsumerInfoError) Error() string {
	return fmt.Sprintf("failed to update consumer info")
}

func (e *FailedToUpdateConsumerInfoError) Is(target error) bool {
	if _, ok := target.(*FailedToUpdateConsumerInfoError); !ok {
		return false
	}
	return true
}
