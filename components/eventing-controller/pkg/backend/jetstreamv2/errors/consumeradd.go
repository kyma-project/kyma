package errors

import (
	"fmt"
)

type FailedToAddConsumerError struct {
	OriginalError error
}

func NewFailedToAddConsumerError(reason error) error {
	return fmt.Errorf("%w", &FailedToAddConsumerError{OriginalError: reason})
}

func (e *FailedToAddConsumerError) Error() string {
	return fmt.Sprintf("failed add a consumer: %v", e.OriginalError)
}

func (e *FailedToAddConsumerError) Is(target error) bool {
	if _, ok := target.(*FailedToAddConsumerError); !ok {
		return false
	}
	return true
}
