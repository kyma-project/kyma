package errors

import (
	"fmt"
)

type FailedToReadConfigError struct {
	FieldName     string
	OriginalError error
}

func NewFailedToReadConfigError(fieldName string, reason error) error {
	return fmt.Errorf("%w", &FailedToReadConfigError{FieldName: fieldName, OriginalError: reason})
}

func (e *FailedToReadConfigError) Error() string {
	return fmt.Sprintf("failed to fetch consumer info: %v", e.OriginalError)
}

func (e *FailedToReadConfigError) Is(target error) bool {
	if _, ok := target.(*FailedToReadConfigError); !ok {
		return false
	}
	return true
}
