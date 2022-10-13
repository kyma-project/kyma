package errors

import (
	"fmt"
)

type FailedToFetchConsumerInfoError struct {
	OriginalError error
}

func NewFailedToFetchConsumerInfoError(reason error) error {
	return fmt.Errorf("%w", &FailedToFetchConsumerInfoError{OriginalError: reason})
}

func (e *FailedToFetchConsumerInfoError) Error() string {
	return fmt.Sprintf("failed to fetch consumer info : %v", e.OriginalError)
}

func (e *FailedToFetchConsumerInfoError) Is(target error) bool {
	if _, ok := target.(*FailedToFetchConsumerInfoError); !ok {
		return false
	}
	return true
}
